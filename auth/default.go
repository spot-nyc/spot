package auth

import (
	"context"

	"golang.org/x/oauth2"
)

// DefaultTokenSource returns the default token source for the current process:
//
//  1. Credentials are resolved via DefaultStore (EnvStore → KeyringStore → FileStore).
//  2. Expired tokens auto-refresh via oauth2.Config.TokenSource against
//     DefaultOAuthConfig().
//  3. Rotated tokens are persisted back to the first writable store.
//
// The returned source uses context.Background() for refresh. Callers wanting
// to control refresh cancellation should construct their own source via
// NewRefreshingTokenSource(ctx, cfg, store).
func DefaultTokenSource() oauth2.TokenSource {
	return NewRefreshingTokenSource(context.Background(), DefaultOAuthConfig(), DefaultStore())
}
