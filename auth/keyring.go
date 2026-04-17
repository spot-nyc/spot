package auth

import (
	"encoding/json"
	"errors"

	"github.com/zalando/go-keyring"
)

// KeyringServiceName is the service identifier used when storing credentials
// in the OS keychain. macOS Keychain groups by service; Linux libsecret uses
// it as the collection key; Windows Credential Manager uses it as the target.
const KeyringServiceName = "com.spot-nyc.cli"

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
// no entry exists.
func (k *KeyringStore) Load() (Credentials, error) {
	data, err := keyring.Get(KeyringServiceName, k.account)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
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

// Delete removes credentials from the OS keychain. Missing entry is not an error.
func (k *KeyringStore) Delete() error {
	if err := keyring.Delete(KeyringServiceName, k.account); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil
		}
		return err
	}
	return nil
}
