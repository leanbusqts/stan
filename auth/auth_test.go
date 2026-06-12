package auth

import (
	"errors"
	"net/url"
	"testing"

	"golang.org/x/oauth2"
)

func TestParseRedirectCode(t *testing.T) {
	code, err := parseRedirectCode("http://127.0.0.1:1234/?code=abc123&state=x")
	if err != nil {
		t.Fatalf("parseRedirectCode returned error: %v", err)
	}
	if code != "abc123" {
		t.Fatalf("parseRedirectCode mismatch: got %q", code)
	}
}

func TestParseRedirectCodeRejectsInvalidInput(t *testing.T) {
	if _, err := parseRedirectCode(""); err == nil {
		t.Fatal("expected error for empty input")
	}
	if _, err := parseRedirectCode("http://127.0.0.1:1234/?state=x"); err == nil {
		t.Fatal("expected error when code is missing")
	}
}

func TestClassifyAuthError(t *testing.T) {
	got := classifyAuthError(errors.New("invalid_grant"))
	if got == nil || got.Error() != "Session expired or revoked.\nRun: stan auth login" {
		t.Fatalf("classifyAuthError mismatch: %v", got)
	}
}

func TestBuildAuthCodeURLUsesVerifierForPKCEChallenge(t *testing.T) {
	conf := &oauth2.Config{
		ClientID:    "client-id",
		Endpoint:    oauth2.Endpoint{AuthURL: "https://accounts.example.test/auth"},
		RedirectURL: "http://127.0.0.1:1234",
	}
	verifier := "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFG"

	authURL := buildAuthCodeURL(conf, "state-value", verifier)
	parsed, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("parse auth URL: %v", err)
	}
	query := parsed.Query()

	if got, want := query.Get("code_challenge"), oauth2.S256ChallengeFromVerifier(verifier); got != want {
		t.Fatalf("code_challenge mismatch: got %q want %q", got, want)
	}
	if got := query.Get("code_challenge_method"); got != "S256" {
		t.Fatalf("code_challenge_method mismatch: got %q", got)
	}
	if got := query.Get("access_type"); got != "offline" {
		t.Fatalf("access_type mismatch: got %q", got)
	}
	if got := query.Get("prompt"); got != "consent" {
		t.Fatalf("prompt mismatch: got %q", got)
	}
}
