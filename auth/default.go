package auth

import "golang.org/x/oauth2"

// DefaultTokenSource returns the default token source for the current process.
//
// In M1a, this is simply an EnvStore-backed TokenSource. M1b extends the
// resolution chain to env → keyring → file, with env always taking precedence
// when SPOT_TOKEN is set.
func DefaultTokenSource() oauth2.TokenSource {
	return NewTokenSource(EnvStore{})
}
