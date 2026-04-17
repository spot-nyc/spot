package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// fakeStore is a test-only Store implementation.
type fakeStore struct {
	creds Credentials
	err   error
}

func (f *fakeStore) Load() (Credentials, error) { return f.creds, f.err }
func (f *fakeStore) Save(c Credentials) error   { f.creds = c; return nil }
func (f *fakeStore) Delete() error              { f.creds = Credentials{}; return nil }

func TestNewTokenSource_ReturnsAccessTokenFromStore(t *testing.T) {
	store := &fakeStore{
		creds: Credentials{
			AccessToken:  "abc",
			RefreshToken: "refresh-1",
			TokenType:    "Bearer",
			Expiry:       time.Now().Add(1 * time.Hour),
		},
	}

	ts := NewTokenSource(store)
	tok, err := ts.Token()
	require.NoError(t, err)
	assert.Equal(t, "abc", tok.AccessToken)
	assert.Equal(t, "refresh-1", tok.RefreshToken)
	assert.Equal(t, "Bearer", tok.TokenType)
}

func TestNewTokenSource_PropagatesStoreError(t *testing.T) {
	store := &fakeStore{err: ErrNoCredentials}

	ts := NewTokenSource(store)
	_, err := ts.Token()
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoCredentials))
}

func TestRefreshingTokenSource_UsesStoredTokenWhenValid(t *testing.T) {
	future := time.Now().Add(1 * time.Hour)
	store := &fakeStore{
		creds: Credentials{
			AccessToken:  "still-valid",
			RefreshToken: "refresh-1",
			TokenType:    "Bearer",
			Expiry:       future,
		},
	}

	cfg := &oauth2.Config{
		ClientID: "test-client",
		Endpoint: oauth2.Endpoint{TokenURL: "https://example.invalid/never-called"},
	}

	ts := NewRefreshingTokenSource(context.Background(), cfg, store)
	tok, err := ts.Token()
	require.NoError(t, err)
	assert.Equal(t, "still-valid", tok.AccessToken)
}

func TestRefreshingTokenSource_RefreshesExpiredAndPersists(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		assert.Equal(t, "refresh_token", r.FormValue("grant_type"))
		assert.Equal(t, "refresh-1", r.FormValue("refresh_token"))
		assert.Equal(t, "test-client", r.FormValue("client_id"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "rotated-access",
			"refresh_token": "rotated-refresh",
			"token_type":    "bearer",
			"expires_in":    3600,
		})
	}))
	defer tokenSrv.Close()

	past := time.Now().Add(-1 * time.Hour)
	store := &fakeStore{
		creds: Credentials{
			AccessToken:  "expired-access",
			RefreshToken: "refresh-1",
			TokenType:    "Bearer",
			Expiry:       past,
		},
	}

	cfg := &oauth2.Config{
		ClientID: "test-client",
		Endpoint: oauth2.Endpoint{
			TokenURL:  tokenSrv.URL,
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	ts := NewRefreshingTokenSource(context.Background(), cfg, store)

	tok, err := ts.Token()
	require.NoError(t, err)
	assert.Equal(t, "rotated-access", tok.AccessToken)

	stored, err := store.Load()
	require.NoError(t, err)
	assert.Equal(t, "rotated-access", stored.AccessToken)
	assert.Equal(t, "rotated-refresh", stored.RefreshToken)
}

func TestRefreshingTokenSource_RefreshFailurePropagates(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer tokenSrv.Close()

	past := time.Now().Add(-1 * time.Hour)
	store := &fakeStore{
		creds: Credentials{
			AccessToken:  "expired",
			RefreshToken: "bad-refresh",
			Expiry:       past,
		},
	}

	cfg := &oauth2.Config{
		ClientID: "test-client",
		Endpoint: oauth2.Endpoint{TokenURL: tokenSrv.URL},
	}

	ts := NewRefreshingTokenSource(context.Background(), cfg, store)
	_, err := ts.Token()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid_grant")
}

func TestRefreshingTokenSource_StoreLoadErrorPropagates(t *testing.T) {
	storeErr := errors.New("disk broken")
	store := &fakeStore{err: storeErr}

	ts := NewRefreshingTokenSource(context.Background(), &oauth2.Config{}, store)
	_, err := ts.Token()
	require.Error(t, err)
	assert.ErrorIs(t, err, storeErr)
}
