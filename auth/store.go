// Package auth handles OAuth 2.1 authentication against Supabase and
// persistence of credentials across processes.
package auth

import (
	"errors"
	"time"

	"golang.org/x/oauth2"
)

// ErrNoCredentials is returned by Store.Load when no credentials are present.
var ErrNoCredentials = errors.New("auth: no credentials stored")

// ErrReadOnly is returned by Save/Delete on read-only stores (e.g., EnvStore).
var ErrReadOnly = errors.New("auth: credential store is read-only")

// Credentials holds a persisted OAuth 2.0 token triple.
//
// This is a thin wrapper around oauth2.Token so our persistence schema stays
// under our control — if oauth2.Token ever gains fields we don't want on disk,
// we serialize explicitly here.
type Credentials struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
	Expiry       time.Time `json:"expiry,omitempty"`
}

// Token converts Credentials to an oauth2.Token.
func (c Credentials) Token() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  c.AccessToken,
		RefreshToken: c.RefreshToken,
		TokenType:    c.TokenType,
		Expiry:       c.Expiry,
	}
}

// CredentialsFromToken converts an oauth2.Token to Credentials.
func CredentialsFromToken(t *oauth2.Token) Credentials {
	if t == nil {
		return Credentials{}
	}
	return Credentials{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		TokenType:    t.TokenType,
		Expiry:       t.Expiry,
	}
}

// Store abstracts credential persistence. Implementations include EnvStore
// (read-only env vars — Task 7), FileStore (JSON on disk — M1b), and
// KeyringStore (OS keychain — M1b).
type Store interface {
	Load() (Credentials, error)
	Save(Credentials) error
	Delete() error
}
