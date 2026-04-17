package auth

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// readOnlyStore returns ErrReadOnly on writes; configurable creds and load error.
type readOnlyStore struct {
	creds Credentials
	err   error
}

func (r *readOnlyStore) Load() (Credentials, error) { return r.creds, r.err }
func (r *readOnlyStore) Save(Credentials) error     { return ErrReadOnly }
func (r *readOnlyStore) Delete() error              { return ErrReadOnly }

func TestChainedStore_Load_FirstNonEmpty(t *testing.T) {
	first := &readOnlyStore{err: ErrNoCredentials}
	second := &readOnlyStore{creds: Credentials{AccessToken: "from-second"}}
	third := &fakeStore{creds: Credentials{AccessToken: "from-third"}}

	chain := &ChainedStore{Stores: []Store{first, second, third}}

	creds, err := chain.Load()
	require.NoError(t, err)
	assert.Equal(t, "from-second", creds.AccessToken)
}

func TestChainedStore_Load_AllEmpty(t *testing.T) {
	chain := &ChainedStore{Stores: []Store{
		&readOnlyStore{err: ErrNoCredentials},
		&readOnlyStore{err: ErrNoCredentials},
	}}

	_, err := chain.Load()
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoCredentials))
}

func TestChainedStore_Load_PropagatesUnexpectedError(t *testing.T) {
	unexpected := errors.New("disk broken")
	chain := &ChainedStore{Stores: []Store{
		&readOnlyStore{err: unexpected},
		&readOnlyStore{creds: Credentials{AccessToken: "unreached"}},
	}}

	_, err := chain.Load()
	require.Error(t, err)
	assert.ErrorIs(t, err, unexpected)
}

func TestChainedStore_Save_SkipsReadOnly(t *testing.T) {
	readOnly := &readOnlyStore{}
	writable := &fakeStore{}

	chain := &ChainedStore{Stores: []Store{readOnly, writable}}

	err := chain.Save(Credentials{AccessToken: "saved"})
	require.NoError(t, err)
	assert.Equal(t, "saved", writable.creds.AccessToken)
}

func TestChainedStore_Save_NoWritableStores(t *testing.T) {
	chain := &ChainedStore{Stores: []Store{
		&readOnlyStore{},
		&readOnlyStore{},
	}}

	err := chain.Save(Credentials{AccessToken: "x"})
	require.Error(t, err)
}

func TestChainedStore_Delete_AllStores(t *testing.T) {
	first := &fakeStore{creds: Credentials{AccessToken: "a"}}
	second := &fakeStore{creds: Credentials{AccessToken: "b"}}

	chain := &ChainedStore{Stores: []Store{first, second}}

	require.NoError(t, chain.Delete())
	assert.Empty(t, first.creds.AccessToken)
	assert.Empty(t, second.creds.AccessToken)
}

func TestDefaultStore_Compose(t *testing.T) {
	s := DefaultStore()

	chain, ok := s.(*ChainedStore)
	require.True(t, ok, "DefaultStore should return *ChainedStore")
	require.Len(t, chain.Stores, 3)

	_, envOK := chain.Stores[0].(EnvStore)
	assert.True(t, envOK, "Stores[0] should be EnvStore, got %T", chain.Stores[0])

	_, krOK := chain.Stores[1].(*KeyringStore)
	assert.True(t, krOK, "Stores[1] should be *KeyringStore, got %T", chain.Stores[1])

	_, fsOK := chain.Stores[2].(*FileStore)
	assert.True(t, fsOK, "Stores[2] should be *FileStore, got %T", chain.Stores[2])
}
