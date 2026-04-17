package spot

import (
	"context"
	"net/http"
)

// UsersService handles the /users endpoints.
type UsersService struct {
	client *Client
}

// User is the current user profile returned by UsersService.Me.
//
// The field set is conservative and expected to grow as the /users/me
// endpoint is validated against the real server in M1b.
type User struct {
	ID    string `json:"id"`
	Phone string `json:"phone,omitempty"`
	Name  string `json:"name,omitempty"`
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
