package auth

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

func setupMockKeyring(t *testing.T) {
	t.Helper()
	keyring.MockInit()
	t.Cleanup(func() {
		keyring.MockInit()
	})
}

func TestKeyringStore_RoundTrip(t *testing.T) {
	setupMockKeyring(t)

	store := NewKeyringStore("test-account")
	creds := Credentials{
		AccessToken:  "access-abc",
		RefreshToken: "refresh-xyz",
		TokenType:    "Bearer",
	}

	require.NoError(t, store.Save(creds))

	loaded, err := store.Load()
	require.NoError(t, err)
	assert.Equal(t, "access-abc", loaded.AccessToken)
	assert.Equal(t, "refresh-xyz", loaded.RefreshToken)
	assert.Equal(t, "Bearer", loaded.TokenType)
}

func TestKeyringStore_Load_Missing(t *testing.T) {
	setupMockKeyring(t)

	store := NewKeyringStore("missing-account")
	_, err := store.Load()
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoCredentials))
}

func TestKeyringStore_Delete(t *testing.T) {
	setupMockKeyring(t)

	store := NewKeyringStore("deleteme")
	require.NoError(t, store.Save(Credentials{AccessToken: "x"}))

	require.NoError(t, store.Delete())

	_, err := store.Load()
	assert.True(t, errors.Is(err, ErrNoCredentials))
}

func TestKeyringStore_Delete_Missing_IsNoop(t *testing.T) {
	setupMockKeyring(t)

	store := NewKeyringStore("ghost")
	require.NoError(t, store.Delete())
}

func TestKeyringStore_DefaultAccount(t *testing.T) {
	setupMockKeyring(t)

	store := NewKeyringStore("")
	require.NoError(t, store.Save(Credentials{AccessToken: "a"}))

	loaded, err := store.Load()
	require.NoError(t, err)
	assert.Equal(t, "a", loaded.AccessToken)
}

func TestIsKeyringUnavailable(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{
			"linux no d-bus",
			errors.New("The name org.freedesktop.secrets was not provided by any .service files"),
			true,
		},
		{
			"generic d-bus name missing",
			errors.New("something org.freedesktop.secrets something"),
			true,
		},
		{
			"secret service missing",
			errors.New("SecretService not available"),
			true,
		},
		{"unrelated error", errors.New("permission denied"), false},
		{"not-found (should be handled separately)", keyring.ErrNotFound, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, isKeyringUnavailable(tc.err))
		})
	}
}
