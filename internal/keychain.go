package internal

import (
	"errors"

	"github.com/zalando/go-keyring"
)

const (
	keychainService = "stan"
	keychainUser    = "oauth-token"
)

// KeychainStore persists token envelopes in the OS keychain.
type KeychainStore struct{}

// Get reads the current token envelope from keychain.
func (KeychainStore) Get() (string, error) {
	return keyring.Get(keychainService, keychainUser)
}

// Set writes the current token envelope to keychain.
func (KeychainStore) Set(value string) error {
	return keyring.Set(keychainService, keychainUser, value)
}

// Delete removes the keychain entry when present.
func (KeychainStore) Delete() error {
	err := keyring.Delete(keychainService, keychainUser)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}
