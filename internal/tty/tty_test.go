package tty

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsTerminal_RegularFileIsNotTerminal(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "not-a-terminal")
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	assert.False(t, IsTerminal(f.Fd()), "a regular file should not be reported as a terminal")
}

func TestIsTerminal_PipeIsNotTerminal(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer func() {
		_ = r.Close()
		_ = w.Close()
	}()

	assert.False(t, IsTerminal(r.Fd()), "a pipe read end should not be reported as a terminal")
	assert.False(t, IsTerminal(w.Fd()), "a pipe write end should not be reported as a terminal")
}
