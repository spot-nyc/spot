package spot

import (
	"context"
	"errors"
	"net/http"
)

// UsersService handles the /users endpoints.
type UsersService struct {
	client *Client
}

// User is the current user profile returned by UsersService.Me.
//
// The "*Connected" booleans indicate which booking platforms the user has
// linked their credentials to. Use ConnectedPlatforms for a display-ready
// list.
type User struct {
	ID                  string `json:"id"`
	Phone               string `json:"phone,omitempty"`
	Name                string `json:"name,omitempty"`
	ResyConnected       bool   `json:"resyConnected"`
	OpenTableConnected  bool   `json:"openTableConnected"`
	SevenRoomsConnected bool   `json:"sevenRoomsConnected"`
	DoorDashConnected   bool   `json:"doorDashConnected"`
}

// ConnectedPlatforms returns the display names of booking platforms the user
// has linked, in a stable order. Mirrors Restaurant.Platforms.
func (u User) ConnectedPlatforms() []string {
	platforms := make([]string, 0, 4)
	if u.ResyConnected {
		platforms = append(platforms, "Resy")
	}
	if u.OpenTableConnected {
		platforms = append(platforms, "OpenTable")
	}
	if u.SevenRoomsConnected {
		platforms = append(platforms, "SevenRooms")
	}
	if u.DoorDashConnected {
		platforms = append(platforms, "DoorDash")
	}
	return platforms
}

// meResponse matches the {"user": {...}} envelope the server wraps /users/me in.
type meResponse struct {
	User User `json:"user"`
}

// Me returns the currently-authenticated user's profile.
func (s *UsersService) Me(ctx context.Context) (*User, error) {
	var resp meResponse
	if err := s.client.do(ctx, http.MethodGet, "/users/me", nil, &resp); err != nil {
		return nil, err
	}
	return &resp.User, nil
}

// Logout revokes the user's session server-side via the Spot API, which
// forwards to Supabase's admin.signOut. scope accepts:
//   - ""        — omitted from body; server defaults to "local" (current session)
//   - "local"   — current session only (explicit)
//   - "global"  — every active session for this user
//   - "others"  — every session except the calling one
//
// Idempotent: returns nil when the end state ("user is logged out") is
// already true — both for a local client with no valid token
// (ErrUnauthenticated from the pre-flight token check) and for a server-side
// 401 (auth middleware rejected an already-revoked token). Any other error
// means the revocation was not confirmed; callers should still proceed with
// local credential cleanup but surface a warning.
func (s *UsersService) Logout(ctx context.Context, scope string) error {
	body := struct {
		Scope string `json:"scope,omitempty"`
	}{Scope: scope}

	err := s.client.do(ctx, http.MethodPost, "/users/me/logout", body, nil)
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrUnauthenticated) {
		return nil
	}
	return err
}
