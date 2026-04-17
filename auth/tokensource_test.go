package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
