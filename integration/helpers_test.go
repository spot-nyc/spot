//go:build integration

package integration

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/spot-nyc/spot"
)

const (
	envAccessToken = "SPOT_TEST_ACCESS_TOKEN"
	envBaseURL     = "SPOT_TEST_BASE_URL"
	envUserID      = "SPOT_TEST_USER_ID"
)

// requireClient builds an authenticated Spot SDK client for integration tests.
// Fatals the test if SPOT_TEST_ACCESS_TOKEN is not set. Honors SPOT_TEST_BASE_URL
// for pointing at a non-prod endpoint.
func requireClient(t *testing.T) *spot.Client {
	t.Helper()

	token := os.Getenv(envAccessToken)
	if token == "" {
		t.Fatalf("%s not set — cannot run integration tests. See docs/integration-testing.md.", envAccessToken)
	}

	opts := []spot.Option{spot.WithToken(token)}
	if base := os.Getenv(envBaseURL); base != "" {
		opts = append(opts, spot.WithBaseURL(base))
	}

	client, err := spot.NewClient(opts...)
	require.NoError(t, err)
	return client
}

// integrationEnabled reports whether the env looks ready for integration tests.
// Used by TestMain to skip cleanly when secrets are missing.
func integrationEnabled() bool {
	return os.Getenv(envAccessToken) != ""
}

// logSkip prints a single consistent skip message to stderr.
func logSkip(reason string) {
	fmt.Fprintf(os.Stderr, "[integration] skipping: %s\n", reason)
}
