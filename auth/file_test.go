package auth

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStore_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")

	store := NewFileStore(path)

	creds := Credentials{
		AccessToken:  "access-abc",
		RefreshToken: "refresh-xyz",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(1 * time.Hour).UTC().Truncate(time.Second),
	}

	require.NoError(t, store.Save(creds))

	loaded, err := store.Load()
	require.NoError(t, err)
	assert.Equal(t, creds.AccessToken, loaded.AccessToken)
	assert.Equal(t, creds.RefreshToken, loaded.RefreshToken)
	assert.Equal(t, creds.TokenType, loaded.TokenType)
	assert.True(t, creds.Expiry.Equal(loaded.Expiry), "expiry round-trip: want %v, got %v", creds.Expiry, loaded.Expiry)
}

func TestFileStore_Load_MissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "does-not-exist.json")

	store := NewFileStore(path)

	_, err := store.Load()
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoCredentials))
}

func TestFileStore_Load_EmptyAccessToken(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"access_token":""}`), 0600))

	store := NewFileStore(path)

	_, err := store.Load()
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoCredentials))
}

func TestFileStore_Delete(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")

	store := NewFileStore(path)
	require.NoError(t, store.Save(Credentials{AccessToken: "x"}))

	require.FileExists(t, path)

	require.NoError(t, store.Delete())
	assert.NoFileExists(t, path)
}

func TestFileStore_Delete_MissingFile_IsNoop(t *testing.T) {
	store := NewFileStore(filepath.Join(t.TempDir(), "ghost.json"))

	require.NoError(t, store.Delete())
}

func TestFileStore_Save_SetsRestrictivePerms(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file mode bits not meaningful on Windows")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "credentials.json")
	store := NewFileStore(path)

	require.NoError(t, store.Save(Credentials{AccessToken: "x"}))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(),
		"credentials.json should have 0600 perms, got %o", info.Mode().Perm())
}

func TestFileStore_Save_CreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "spot", "credentials.json")

	store := NewFileStore(path)

	require.NoError(t, store.Save(Credentials{AccessToken: "x"}))
	require.FileExists(t, path)
}
