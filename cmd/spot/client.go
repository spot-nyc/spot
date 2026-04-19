package main

import (
	"os"

	"github.com/spot-nyc/spot"
	"github.com/spot-nyc/spot/auth"
)

// SpotBaseURLEnv names an optional environment variable for overriding the
// morty API base URL. Unset in normal use; the CLI falls back to
// spot.DefaultBaseURL. Primarily intended for end-to-end CLI tests pointing
// at an httptest server.
const SpotBaseURLEnv = "SPOT_BASE_URL"

// newClient builds a spot.Client wired to the default token source (env →
// keyring → file) and, when set, honors SPOT_BASE_URL to override the API
// base URL.
func newClient() (*spot.Client, error) {
	opts := []spot.Option{spot.WithTokenSource(auth.DefaultTokenSource())}
	if base := os.Getenv(SpotBaseURLEnv); base != "" {
		opts = append(opts, spot.WithBaseURL(base))
	}
	return spot.NewClient(opts...)
}
