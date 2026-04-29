package auth

import (
	"errors"
	"testing"
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
