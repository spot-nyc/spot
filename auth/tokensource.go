package auth

import "golang.org/x/oauth2"

// NewTokenSource returns an oauth2.TokenSource backed by a Store.
//
// Each call to Token() reads from the store. No refresh is performed at this
// stage — M1b adds a refresh-aware wrapper that uses oauth2.Config.TokenSource
// to refresh expired tokens and persist rotated refresh tokens back to the
// store.
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
