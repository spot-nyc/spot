package auth

import (
	"errors"
	"fmt"
)

// ChainedStore resolves credentials across multiple underlying stores in order.
//
// Load: returns the first non-empty credentials, or ErrNoCredentials if all
// stores are empty. Unexpected errors from any store short-circuit.
//
// Save: tries each store in order, skipping those that return ErrReadOnly.
// Returns success on the first writer that succeeds.
//
// Delete: best-effort across all stores. ErrNoCredentials and ErrReadOnly are
// swallowed; the first real error (if any) is returned.
type ChainedStore struct {
	Stores []Store
}

// Load returns the first non-empty credentials across the chain.
func (c *ChainedStore) Load() (Credentials, error) {
	for _, s := range c.Stores {
		creds, err := s.Load()
		if err == nil {
			return creds, nil
		}
		if !errors.Is(err, ErrNoCredentials) {
			return Credentials{}, err
		}
	}
	return Credentials{}, ErrNoCredentials
}

// Save writes to the first writable store (skipping read-only).
func (c *ChainedStore) Save(creds Credentials) error {
	var lastErr error
	for _, s := range c.Stores {
		err := s.Save(creds)
		if err == nil {
			return nil
		}
		if errors.Is(err, ErrReadOnly) {
			continue
		}
		lastErr = err
	}
	if lastErr != nil {
		return fmt.Errorf("auth: save failed across all stores: %w", lastErr)
	}
	return errors.New("auth: no writable store available")
}

// Delete attempts to remove credentials from all stores. Missing and read-only
// stores are skipped silently.
func (c *ChainedStore) Delete() error {
	var firstErr error
	for _, s := range c.Stores {
		err := s.Delete()
		if err == nil {
			continue
		}
		if errors.Is(err, ErrNoCredentials) || errors.Is(err, ErrReadOnly) {
			continue
		}
		if firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// DefaultStore returns the default credential store chain:
// EnvStore → KeyringStore → FileStore.
//
// EnvStore takes precedence so CI and scripted use cases can override any
// persisted credentials. KeyringStore is next for interactive users.
// FileStore is the fallback for headless Linux without a keyring backend.
func DefaultStore() Store {
	return &ChainedStore{
		Stores: []Store{
			EnvStore{},
			NewKeyringStore(""),
			NewFileStore(""),
		},
	}
}
