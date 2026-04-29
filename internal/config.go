package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	ScopeCalendarEvents = "https://www.googleapis.com/auth/calendar.events"
	ScopeTasks          = "https://www.googleapis.com/auth/tasks"
	ScopeOpenID         = "openid"
	ScopeEmail          = "email"
)

// Config holds filesystem and runtime settings for Stan.
type Config struct {
	HomeDir             string
	ConfigDir           string
	TokenPath           string
	CredentialsPath     string
	CredentialsFallback string
}

// LoadConfig resolves local filesystem paths used by the CLI.
func LoadConfig() (Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("resolve home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "stan")
	cfg := Config{
		HomeDir:             homeDir,
		ConfigDir:           configDir,
		TokenPath:           filepath.Join(configDir, "token.json"),
		CredentialsFallback: filepath.Join(configDir, "client_secret.json"),
	}

	if _, err := os.Stat("client_secret.json"); err == nil {
		cfg.CredentialsPath = "client_secret.json"
	} else if _, err := os.Stat(cfg.CredentialsFallback); err == nil {
		cfg.CredentialsPath = cfg.CredentialsFallback
	}

	return cfg, nil
}

// EnsureConfigDir creates the Stan config directory when missing.
func EnsureConfigDir(cfg Config) error {
	if err := os.MkdirAll(cfg.ConfigDir, 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	return nil
}

// MissingCredentialsError renders the onboarding message required by the spec.
func MissingCredentialsError(cfg Config) error {
	return fmt.Errorf(
		"Missing credentials file.\nPlace client_secret.json in:\n- current directory\n- %s\nGoogle Cloud Console:\nhttps://console.cloud.google.com/",
		cfg.ConfigDir,
	)
}

// LoadOAuthConfig loads Google Desktop OAuth credentials and scopes.
func LoadOAuthConfig(cfg Config, redirectURL string) (*oauth2.Config, error) {
	if cfg.CredentialsPath == "" {
		return nil, MissingCredentialsError(cfg)
	}

	raw, err := os.ReadFile(cfg.CredentialsPath)
	if err != nil {
		return nil, fmt.Errorf("read credentials: %w", err)
	}

	conf, err := google.ConfigFromJSON(raw, ScopeCalendarEvents, ScopeTasks, ScopeOpenID, ScopeEmail)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}
	conf.RedirectURL = redirectURL
	return conf, nil
}

// DesktopClientSecret returns the configured client secret if present.
func DesktopClientSecret(cfg Config) (string, error) {
	if cfg.CredentialsPath == "" {
		return "", MissingCredentialsError(cfg)
	}

	raw, err := os.ReadFile(cfg.CredentialsPath)
	if err != nil {
		return "", fmt.Errorf("read credentials: %w", err)
	}

	var parsed struct {
		Installed struct {
			ClientSecret string `json:"client_secret"`
		} `json:"installed"`
	}

	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("parse credentials: %w", err)
	}
	if parsed.Installed.ClientSecret == "" {
		return "", errors.New("client_secret missing from Desktop App credentials")
	}
	return parsed.Installed.ClientSecret, nil
}
