package auth

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"stan/internal"
)

const (
	userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	revokeURL   = "https://oauth2.googleapis.com/revoke"
)

// Manager handles Stan authentication and token lifecycle operations.
type Manager struct {
	cfg   internal.Config
	store *internal.CompositeTokenStore
}

// Status describes the current auth state.
type Status struct {
	Email   string               `json:"email"`
	Expiry  time.Time            `json:"expiry"`
	Storage internal.StorageKind `json:"storage"`
}

// NewManager creates a new auth manager.
func NewManager(cfg internal.Config) *Manager {
	return &Manager{
		cfg:   cfg,
		store: internal.NewTokenStore(cfg),
	}
}

// Login performs the OAuth desktop auth flow and persists the token.
func (m *Manager) Login(ctx context.Context) error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("start local listener: %w", err)
	}
	defer listener.Close()

	redirectURL := "http://" + listener.Addr().String()
	conf, err := internal.LoadOAuthConfig(m.cfg, redirectURL)
	if err != nil {
		return err
	}

	state, err := randomURLSafe(32)
	if err != nil {
		return fmt.Errorf("generate state: %w", err)
	}
	verifier, err := randomURLSafe(64)
	if err != nil {
		return fmt.Errorf("generate PKCE verifier: %w", err)
	}
	authURL := buildAuthCodeURL(conf, state, verifier)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	server := &http.Server{Handler: callbackHandler(state, codeCh, errCh)}

	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("callback server failed: %w", err)
		}
	}()
	defer server.Shutdown(context.Background())

	fmt.Println("Open this URL to authenticate:")
	fmt.Println(authURL)

	if headless() || !tryOpenBrowser(authURL) {
		return m.completeManualFlow(ctx, conf, verifier)
	}

	fmt.Println("Press Enter to switch to manual auth.")

	stdinCh := make(chan string, 1)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		stdinCh <- strings.TrimSpace(line)
	}()

	select {
	case code := <-codeCh:
		return m.exchangeAndSave(ctx, conf, code, verifier)
	case err := <-errCh:
		return err
	case line := <-stdinCh:
		if line == "" {
			return m.completeManualFlow(ctx, conf, verifier)
		}
		code, err := parseRedirectCode(line)
		if err != nil {
			return err
		}
		return m.exchangeAndSave(ctx, conf, code, verifier)
	case <-time.After(30 * time.Second):
		fmt.Println("Paste the full redirect URL:")
		select {
		case code := <-codeCh:
			return m.exchangeAndSave(ctx, conf, code, verifier)
		case err := <-errCh:
			return err
		case line := <-stdinCh:
			code, err := parseRedirectCode(line)
			if err != nil {
				return err
			}
			return m.exchangeAndSave(ctx, conf, code, verifier)
		case <-ctx.Done():
			return ctx.Err()
		}
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Status reads the persisted login state.
func (m *Manager) Status(ctx context.Context) (*Status, error) {
	record, err := m.store.Load()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errors.New("not logged in")
		}
		return nil, err
	}

	status := &Status{
		Email:   record.Email,
		Expiry:  record.Token.Expiry,
		Storage: m.store.Kind(),
	}
	if status.Email == "" {
		status.Email = "unknown"
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return status, nil
}

// Logout revokes the remote token best-effort and removes local state.
func (m *Manager) Logout(ctx context.Context) error {
	record, err := m.store.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if record != nil {
		_ = revokeToken(ctx, record.Token)
	}
	if err := m.store.Delete(); err != nil {
		return err
	}
	fmt.Println("Logged out.")
	return nil
}

// AuthorizedHTTPClient returns an HTTP client with auto-refresh and auto-save.
func (m *Manager) AuthorizedHTTPClient(ctx context.Context) (*http.Client, error) {
	record, err := m.store.Load()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errors.New("not logged in\nRun: stan auth login")
		}
		return nil, err
	}

	conf, err := internal.LoadOAuthConfig(m.cfg, "")
	if err != nil {
		return nil, err
	}

	base := conf.TokenSource(ctx, record.Token)
	source := &savingTokenSource{
		base:   base,
		store:  m.store,
		record: record,
	}

	return oauth2.NewClient(ctx, source), nil
}

func (m *Manager) completeManualFlow(ctx context.Context, conf *oauth2.Config, verifier string) error {
	fmt.Println("Paste the full redirect URL:")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("read redirect URL: %w", err)
	}
	line = strings.TrimSpace(line)
	code, err := parseRedirectCode(line)
	if err != nil {
		return err
	}
	return m.exchangeAndSave(ctx, conf, code, verifier)
}

func (m *Manager) exchangeAndSave(ctx context.Context, conf *oauth2.Config, code string, verifier string) error {
	token, err := conf.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return classifyAuthError(err)
	}

	email, err := fetchEmail(ctx, token.AccessToken)
	if err != nil {
		email = ""
	}
	record := &internal.TokenRecord{
		Email: email,
		Token: token,
	}
	if err := m.store.Save(record); err != nil {
		return err
	}

	if email == "" {
		fmt.Printf("Login complete. Token stored in %s.\n", m.store.Kind())
		return nil
	}
	fmt.Printf("Login complete. Logged in as %s.\n", email)
	return nil
}

func callbackHandler(expectedState string, codeCh chan<- string, errCh chan<- error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != expectedState {
			http.Error(w, "invalid state", http.StatusBadRequest)
			errCh <- errors.New("invalid OAuth state")
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			errCh <- errors.New("callback missing code")
			return
		}
		_, _ = io.WriteString(w, "Authentication complete. You can close this window.")
		select {
		case codeCh <- code:
		default:
		}
	})
}

func tryOpenBrowser(target string) bool {
	var cmd string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}
	if _, err := exec.LookPath(cmd); err != nil {
		return false
	}
	return exec.Command(cmd, target).Start() == nil
}

func headless() bool {
	if runtime.GOOS == "linux" && os.Getenv("DISPLAY") == "" {
		return true
	}
	return false
}

func randomURLSafe(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func buildAuthCodeURL(conf *oauth2.Config, state string, verifier string) string {
	return conf.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.S256ChallengeOption(verifier),
		oauth2.SetAuthURLParam("prompt", "consent"),
	)
}

func fetchEmail(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userInfoURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("userinfo request failed: %s", resp.Status)
	}

	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}
	return body.Email, nil
}

func revokeToken(ctx context.Context, token *oauth2.Token) error {
	value := token.RefreshToken
	if value == "" {
		value = token.AccessToken
	}
	if value == "" {
		return nil
	}

	form := url.Values{}
	form.Set("token", value)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, revokeURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

type savingTokenSource struct {
	base   oauth2.TokenSource
	store  *internal.CompositeTokenStore
	record *internal.TokenRecord
}

func (s *savingTokenSource) Token() (*oauth2.Token, error) {
	token, err := s.base.Token()
	if err != nil {
		return nil, classifyAuthError(err)
	}
	if token.RefreshToken == "" {
		token.RefreshToken = s.record.Token.RefreshToken
	}
	if s.record.Token.AccessToken == token.AccessToken &&
		s.record.Token.RefreshToken == token.RefreshToken &&
		s.record.Token.Expiry.Equal(token.Expiry) {
		return token, nil
	}

	s.record.Token = token
	if err := s.store.Save(s.record); err != nil {
		return nil, err
	}
	return token, nil
}

func classifyAuthError(err error) error {
	message := err.Error()
	if strings.Contains(message, "invalid_grant") || strings.Contains(message, "401") || strings.Contains(message, "400") {
		return errors.New("Session expired or revoked.\nRun: stan auth login")
	}
	return err
}

func parseRedirectCode(line string) (string, error) {
	if strings.TrimSpace(line) == "" {
		return "", errors.New("no redirect URL provided")
	}
	parsed, err := url.Parse(strings.TrimSpace(line))
	if err != nil {
		return "", fmt.Errorf("invalid redirect URL: %w", err)
	}
	code := parsed.Query().Get("code")
	if code == "" {
		return "", errors.New("redirect URL does not contain code=")
	}
	return code, nil
}
