package auth

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// DefaultFileStorePath returns the default location for the credentials file.
// It respects XDG_CONFIG_HOME when set, falling back to $HOME/.config.
func DefaultFileStorePath() (string, error) {
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "spot", "credentials.json"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "spot", "credentials.json"), nil
}

// FileStore persists credentials as JSON at a fixed path with mode 0600.
type FileStore struct {
	path string
}

// NewFileStore constructs a FileStore at the given path. If path is empty,
// the default path (XDG_CONFIG_HOME/spot/credentials.json or
// $HOME/.config/spot/credentials.json) is resolved lazily on first use.
func NewFileStore(path string) *FileStore {
	return &FileStore{path: path}
}

func (f *FileStore) resolve() (string, error) {
	if f.path != "" {
		return f.path, nil
	}
	return DefaultFileStorePath()
}

// Load reads and parses the credentials file. Returns ErrNoCredentials when
// the file does not exist or contains no access token.
func (f *FileStore) Load() (Credentials, error) {
	path, err := f.resolve()
	if err != nil {
		return Credentials{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Credentials{}, ErrNoCredentials
		}
		return Credentials{}, err
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return Credentials{}, err
	}
	if creds.AccessToken == "" {
		return Credentials{}, ErrNoCredentials
	}
	return creds, nil
}

// Save writes the credentials file with 0600 perms, creating parent directories
// as needed (with 0700 perms).
func (f *FileStore) Save(creds Credentials) error {
	path, err := f.resolve()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(creds)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// Delete removes the credentials file. Missing file is not an error.
func (f *FileStore) Delete() error {
	path, err := f.resolve()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
