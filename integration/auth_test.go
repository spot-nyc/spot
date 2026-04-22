//go:build integration

package integration

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_AuthWhoami verifies the SDK can hit /users/me with the
// provided access token and get back a valid user profile. If SPOT_TEST_USER_ID
// is set, we also assert the returned ID matches.
//
// Note: we do NOT integration-test UsersService.Logout here. Logging out the
// session's access token breaks every subsequent test in the suite, and
// minting a disposable session requires rotating the refresh token from
// inside Go, which conflicts with the workflow's already-rotated secret.
// Logout is covered by unit tests + manual smoke. Adding integration coverage
// is tracked as future work (see docs/integration-testing.md, "Known gaps").
func TestIntegration_AuthWhoami(t *testing.T) {
	client := requireClient(t)

	user, err := client.Users.Me(context.Background())
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.NotEmpty(t, user.ID, "user profile should include an ID")

	if expected := os.Getenv(envUserID); expected != "" {
		assert.Equal(t, expected, user.ID, "whoami ID should match SPOT_TEST_USER_ID")
	}
}
