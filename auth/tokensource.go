package auth

import (
	"context"
	"sync"

	"golang.org/x/oauth2"
)

// NewTokenSource returns an oauth2.TokenSource backed by a Store.
//
// Each call to Token() reads from the store. No refresh is performed —
// callers wanting refresh should use NewRefreshingTokenSource. This simple
// source is appropriate for read-only stores (e.g. EnvStore) where tokens
// cannot be rotated back in place.
func NewTokenSource(store Store) oauth2.TokenSource {
	return &storeTokenSource{store: store}
}

type storeTokenSource struct {
	store Store
}

func (t *storeTokenSource) Token() (*oauth2.Token, error) {
	creds, err := t.store.Load()
	if err != nil {
		return nil, err
	}
	return creds.Token(), nil
}

// NewRefreshingTokenSource returns an oauth2.TokenSource that:
//
//  1. Lazily loads credentials from the underlying Store on first Token() call.
//  2. Returns the cached token while it's still valid.
//  3. When the token is expired (or within the oauth2 library's built-in
//     expiry delta), refreshes via cfg.TokenSource(ctx, tok) which uses the
//     refresh_token grant against cfg.Endpoint.TokenURL.
//  4. If the refresh returns a rotated token, persists it to the store.
//  5. Caches the fresh token in memory for subsequent calls.
//
// Refresh errors propagate to the caller. Save errors are non-fatal — the
// returned token is still valid, just not persisted.
func NewRefreshingTokenSource(ctx context.Context, cfg *oauth2.Config, store Store) oauth2.TokenSource {
	return &refreshingTokenSource{
		ctx:   ctx,
		cfg:   cfg,
		store: store,
	}
}

type refreshingTokenSource struct {
	ctx    context.Context
	cfg    *oauth2.Config
	store  Store
	mu     sync.Mutex
	cached *oauth2.Token
}

func (r *refreshingTokenSource) Token() (*oauth2.Token, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cached == nil {
		creds, err := r.store.Load()
		if err != nil {
			return nil, err
		}
		r.cached = creds.Token()
	}

	if r.cached.Valid() {
		return r.cached, nil
	}

	src := r.cfg.TokenSource(r.ctx, r.cached)
	newTok, err := src.Token()
	if err != nil {
		return nil, err
	}

	rotated := newTok.AccessToken != r.cached.AccessToken ||
		newTok.RefreshToken != r.cached.RefreshToken
	if rotated {
		// Non-fatal on save error: we still return the valid token.
		_ = r.store.Save(CredentialsFromToken(newTok))
	}

	r.cached = newTok
	return newTok, nil
}
