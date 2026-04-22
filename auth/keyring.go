package auth

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/zalando/go-keyring"
)

// KeyringServiceName is the service identifier used when storing credentials
// in the OS keychain. macOS Keychain groups by service; Linux libsecret uses
// it as the collection key; Windows Credential Manager uses it as the target.
const KeyringServiceName = "nyc.spot.cli"

// DefaultKeyringAccount is the account identifier used when no explicit
// account is supplied. Enables per-profile accounts in the future.
const DefaultKeyringAccount = "default"

// KeyringStore persists credentials in the OS keychain via go-keyring.
// macOS uses Keychain; Linux uses the Secret Service API via D-Bus; Windows
// uses Credential Manager. Tests should call keyring.MockInit() to swap in
// an in-memory backend.
type KeyringStore struct {
	account string
}

// NewKeyringStore constructs a KeyringStore. An empty account falls back to
// DefaultKeyringAccount.
func NewKeyringStore(account string) *KeyringStore {
	if account == "" {
		account = DefaultKeyringAccount
	}
	return &KeyringStore{account: account}
}

// Load reads credentials from the OS keychain. Returns ErrNoCredentials when
// no entry exists or the keyring backend is unavailable (e.g., Linux without
// a running Secret Service / D-Bus). The "unavailable" case is equivalent to
// "no credentials could have been saved here," so the chain can fall through
// to the file store.
func (k *KeyringStore) Load() (Credentials, error) {
	data, err := keyring.Get(KeyringServiceName, k.account)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) || isKeyringUnavailable(err) {
			return Credentials{}, ErrNoCredentials
		}
		return Credentials{}, err
	}

	var creds Credentials
	if err := json.Unmarshal([]byte(data), &creds); err != nil {
		return Credentials{}, err
	}
	if creds.AccessToken == "" {
		return Credentials{}, ErrNoCredentials
	}
	return creds, nil
}

// Save writes credentials to the OS keychain as a JSON-encoded string.
func (k *KeyringStore) Save(creds Credentials) error {
	data, err := json.Marshal(creds)
	if err != nil {
		return err
	}
	return keyring.Set(KeyringServiceName, k.account, string(data))
}

// Delete removes credentials from the OS keychain. Missing entry is not an
// error, nor is the keyring backend being unavailable — if the keyring can't
// be reached, there was nothing it could hold to begin with, so logout is
// still complete from this store's perspective.
func (k *KeyringStore) Delete() error {
	if err := keyring.Delete(KeyringServiceName, k.account); err != nil {
		if errors.Is(err, keyring.ErrNotFound) || isKeyringUnavailable(err) {
			return nil
		}
		return err
	}
	return nil
}

// isKeyringUnavailable reports whether err indicates the underlying keyring
// backend could not be reached at all. On Linux, go-keyring uses the Secret
// Service via D-Bus; headless environments (CI, minimal containers) often
// lack it and produce an error like "The name org.freedesktop.secrets was
// not provided by any .service files." Treat these as "no storage" so the
// CLI can still function via the file-backed fallback store.
func isKeyringUnavailable(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "org.freedesktop.secrets") ||
		strings.Contains(msg, "not provided by any .service files") ||
		strings.Contains(msg, "SecretService")
}
