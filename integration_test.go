package spot

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spot-nyc/spot/auth"
)

// TestEndToEnd_EnvTokenToUsersMe exercises the full stack:
// SPOT_TOKEN env → auth.DefaultTokenSource → Client → UsersService.Me.
//
// It mocks the server with httptest and verifies the bearer token flows through.
func TestEndToEnd_EnvTokenToUsersMe(t *testing.T) {
	t.Setenv(auth.EnvTokenVar, "env-provided-token")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer env-provided-token", r.Header.Get("Authorization"))
		assert.Equal(t, "/users/me", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"user":{"id":"u-123","phone":"+15555551234"}}`)
	}))
	defer srv.Close()

	client, err := NewClient(
		WithTokenSource(auth.DefaultTokenSource()),
		WithBaseURL(srv.URL),
	)
	require.NoError(t, err)

	user, err := client.Users.Me(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "u-123", user.ID)
	assert.Equal(t, "+15555551234", user.Phone)
}
