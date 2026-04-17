package auth

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvStore_Load_FromSpotToken(t *testing.T) {
	t.Setenv("SPOT_TOKEN", "my-access-token")
	t.Setenv("SPOT_REFRESH_TOKEN", "my-refresh-token")

	s := EnvStore{}
	creds, err := s.Load()
	require.NoError(t, err)
	assert.Equal(t, "my-access-token", creds.AccessToken)
	assert.Equal(t, "my-refresh-token", creds.RefreshToken)
}

func TestEnvStore_Load_OnlyAccessToken(t *testing.T) {
	t.Setenv("SPOT_TOKEN", "my-access-token")
	t.Setenv("SPOT_REFRESH_TOKEN", "")

	s := EnvStore{}
	creds, err := s.Load()
	require.NoError(t, err)
	assert.Equal(t, "my-access-token", creds.AccessToken)
	assert.Empty(t, creds.RefreshToken)
}

func TestEnvStore_Load_NoTokenSet(t *testing.T) {
	t.Setenv("SPOT_TOKEN", "")
	t.Setenv("SPOT_REFRESH_TOKEN", "")

	s := EnvStore{}
	_, err := s.Load()
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoCredentials))
}

func TestEnvStore_Save_IsReadOnly(t *testing.T) {
	s := EnvStore{}
	err := s.Save(Credentials{AccessToken: "nope"})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrReadOnly))
}

func TestEnvStore_Delete_IsReadOnly(t *testing.T) {
	s := EnvStore{}
	err := s.Delete()
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrReadOnly))
}
