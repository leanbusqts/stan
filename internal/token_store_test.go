package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestTokenStoreSaveLoadDeleteWithFileFallback(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{
		ConfigDir: dir,
		TokenPath: filepath.Join(dir, "token.json"),
	}
	store := NewTokenStore(cfg)

	record := &TokenRecord{
		Email: "user@example.com",
		Token: &oauth2.Token{
			AccessToken:  "access",
			RefreshToken: "refresh",
			Expiry:       time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC),
		},
	}

	if err := store.Save(record); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	if got := store.Kind(); got != StorageFile {
		t.Fatalf("store kind mismatch: got %q want %q", got, StorageFile)
	}

	info, err := os.Stat(cfg.TokenPath)
	if err != nil {
		t.Fatalf("stat token file: %v", err)
	}
	if perms := info.Mode().Perm(); perms != 0o600 {
		t.Fatalf("token file permissions mismatch: got %o want %o", perms, 0o600)
	}

	raw, err := os.ReadFile(cfg.TokenPath)
	if err != nil {
		t.Fatalf("read token file: %v", err)
	}
	var persisted TokenRecord
	if err := json.Unmarshal(raw, &persisted); err != nil {
		t.Fatalf("unmarshal token file: %v", err)
	}
	if persisted.Version != 1 {
		t.Fatalf("token version mismatch: got %d", persisted.Version)
	}
	if persisted.Checksum != RefreshChecksum(record.Token) {
		t.Fatalf("checksum mismatch: got %q want %q", persisted.Checksum, RefreshChecksum(record.Token))
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if loaded.Email != record.Email {
		t.Fatalf("loaded email mismatch: got %q want %q", loaded.Email, record.Email)
	}
	if loaded.Token.RefreshToken != record.Token.RefreshToken {
		t.Fatalf("loaded refresh token mismatch: got %q want %q", loaded.Token.RefreshToken, record.Token.RefreshToken)
	}

	if err := store.Delete(); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if _, err := os.Stat(cfg.TokenPath); !os.IsNotExist(err) {
		t.Fatalf("token file still exists after delete, err=%v", err)
	}
}

func TestDecodeTokenRecordRejectsEmptyToken(t *testing.T) {
	_, err := decodeTokenRecord([]byte(`{"version":1}`))
	if err == nil {
		t.Fatal("decodeTokenRecord unexpectedly succeeded")
	}
}

func TestTokenFileDir(t *testing.T) {
	cfg := Config{TokenPath: "/tmp/stan/token.json"}
	if got := TokenFileDir(cfg); got != "/tmp/stan" {
		t.Fatalf("TokenFileDir mismatch: %q", got)
	}
}
