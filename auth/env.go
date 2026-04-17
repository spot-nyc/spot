package auth

import "os"

// EnvTokenVar names the env var that supplies a statically-provided access token.
const EnvTokenVar = "SPOT_TOKEN"

// EnvRefreshTokenVar names the env var that optionally supplies a refresh token.
const EnvRefreshTokenVar = "SPOT_REFRESH_TOKEN"

// EnvStore reads credentials from env vars. It is read-only; Save and Delete
// return ErrReadOnly.
//
// When SPOT_TOKEN is set, EnvStore.Load returns those credentials.
// When SPOT_TOKEN is unset or empty, EnvStore.Load returns ErrNoCredentials.
//
// EnvStore exists primarily for CI environments, MCP-server-mode credential
// passing, and local development where a token has been manually extracted
// from another source.
type EnvStore struct{}

// Load reads credentials from SPOT_TOKEN (and optionally SPOT_REFRESH_TOKEN).
func (EnvStore) Load() (Credentials, error) {
	token := os.Getenv(EnvTokenVar)
	if token == "" {
		return Credentials{}, ErrNoCredentials
	}
	return Credentials{
		AccessToken:  token,
		RefreshToken: os.Getenv(EnvRefreshTokenVar),
		TokenType:    "Bearer",
	}, nil
}

// Save always returns ErrReadOnly.
func (EnvStore) Save(Credentials) error {
	return ErrReadOnly
}

// Delete always returns ErrReadOnly.
func (EnvStore) Delete() error {
	return ErrReadOnly
}
