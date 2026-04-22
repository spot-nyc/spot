package spot

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersService_Me(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/users/me", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"user":{"id":"user-abc","phone":"+15555551234","name":"Brian","resyConnected":true,"openTableConnected":true,"sevenRoomsConnected":false,"doorDashConnected":false}}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	user, err := c.Users.Me(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "user-abc", user.ID)
	assert.Equal(t, "+15555551234", user.Phone)
	assert.Equal(t, "Brian", user.Name)
	assert.True(t, user.ResyConnected)
	assert.True(t, user.OpenTableConnected)
	assert.False(t, user.SevenRoomsConnected)
	assert.False(t, user.DoorDashConnected)
	assert.Equal(t, []string{"Resy", "OpenTable"}, user.ConnectedPlatforms())
}

func TestUser_ConnectedPlatforms(t *testing.T) {
	cases := []struct {
		name string
		user User
		want []string
	}{
		{"none", User{}, []string{}},
		{"resy only", User{ResyConnected: true}, []string{"Resy"}},
		{"opentable only", User{OpenTableConnected: true}, []string{"OpenTable"}},
		{"sevenrooms only", User{SevenRoomsConnected: true}, []string{"SevenRooms"}},
		{"doordash only", User{DoorDashConnected: true}, []string{"DoorDash"}},
		{
			"all",
			User{ResyConnected: true, OpenTableConnected: true, SevenRoomsConnected: true, DoorDashConnected: true},
			[]string{"Resy", "OpenTable", "SevenRooms", "DoorDash"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.user.ConnectedPlatforms())
		})
	}
}

func TestUsersService_Me_Unauthenticated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":"Invalid or expired token"}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	_, err = c.Users.Me(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthenticated)
}
