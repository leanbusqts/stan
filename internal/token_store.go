package internal

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

// StorageKind identifies the token persistence backend in use.
type StorageKind string

const (
	StorageKeychain StorageKind = "keychain"
	StorageFile     StorageKind = "file"
)

// TokenRecord is the persisted token format used by Stan.
type TokenRecord struct {
	Version  int           `json:"version"`
	Email    string        `json:"email"`
	Checksum string        `json:"checksum"`
	Token    *oauth2.Token `json:"token"`
}

// CompositeTokenStore prefers keychain and falls back to the token file.
type CompositeTokenStore struct {
	cfg      Config
	keychain KeychainStore
	lastKind StorageKind
}

// NewTokenStore creates a token store with the required priority order.
func NewTokenStore(cfg Config) *CompositeTokenStore {
	return &CompositeTokenStore{cfg: cfg}
}

// Kind reports the backend used in the last successful operation.
func (s *CompositeTokenStore) Kind() StorageKind {
	if s.lastKind == "" {
		return StorageFile
	}
	return s.lastKind
}

// Load reads an existing token envelope from keychain or file.
func (s *CompositeTokenStore) Load() (*TokenRecord, error) {
	if raw, err := s.keychain.Get(); err == nil {
		record, err := decodeTokenRecord([]byte(raw))
		if err == nil {
			s.lastKind = StorageKeychain
			return record, nil
		}
	}

	raw, err := os.ReadFile(s.cfg.TokenPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, os.ErrNotExist
		}
		return nil, fmt.Errorf("read token file: %w", err)
	}

	record, err := decodeTokenRecord(raw)
	if err != nil {
		return nil, err
	}
	s.lastKind = StorageFile
	return record, nil
}

// Save persists the token envelope using the preferred backend.
func (s *CompositeTokenStore) Save(record *TokenRecord) error {
	record.Version = 1
	record.Checksum = RefreshChecksum(record.Token)

	payload, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal token: %w", err)
	}

	if err := s.keychain.Set(string(payload)); err == nil {
		s.lastKind = StorageKeychain
		return nil
	}

	if err := EnsureConfigDir(s.cfg); err != nil {
		return err
	}

	tmpPath := s.cfg.TokenPath + ".tmp"
	file, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open temp token file: %w", err)
	}
	if _, err := file.Write(payload); err != nil {
		file.Close()
		return fmt.Errorf("write temp token file: %w", err)
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return fmt.Errorf("sync temp token file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close temp token file: %w", err)
	}
	if err := os.Rename(tmpPath, s.cfg.TokenPath); err != nil {
		return fmt.Errorf("replace token file: %w", err)
	}
	if err := os.Chmod(s.cfg.TokenPath, 0o600); err != nil {
		return fmt.Errorf("chmod token file: %w", err)
	}

	s.lastKind = StorageFile
	return nil
}

// Delete removes both keychain and file-backed tokens.
func (s *CompositeTokenStore) Delete() error {
	var errs []error
	if err := s.keychain.Delete(); err != nil {
		errs = append(errs, fmt.Errorf("delete keychain token: %w", err))
	}
	if err := os.Remove(s.cfg.TokenPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		errs = append(errs, fmt.Errorf("delete token file: %w", err))
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

func decodeTokenRecord(raw []byte) (*TokenRecord, error) {
	var record TokenRecord
	if err := json.Unmarshal(raw, &record); err != nil {
		return nil, fmt.Errorf("decode token: %w", err)
	}
	if record.Token == nil {
		return nil, errors.New("stored token is empty")
	}
	expected := RefreshChecksum(record.Token)
	if record.Checksum != "" && expected != record.Checksum {
		fmt.Fprintln(os.Stderr, "warning: stored token checksum mismatch")
	}
	return &record, nil
}

// RefreshChecksum computes the SHA256 of the refresh token.
func RefreshChecksum(token *oauth2.Token) string {
	sum := sha256.Sum256([]byte(token.RefreshToken))
	return hex.EncodeToString(sum[:])
}

// TokenFileDir returns the token directory path.
func TokenFileDir(cfg Config) string {
	return filepath.Dir(cfg.TokenPath)
}
