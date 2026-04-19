package render

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve_ForceJSON(t *testing.T) {
	// Even if out is a TTY, forceJSON wins.
	assert.Equal(t, FormatJSON, Resolve(true, os.Stdout))
}

func TestResolve_NonFileWriterDefaultsToJSON(t *testing.T) {
	// When writing to something that isn't an *os.File (e.g. bytes.Buffer),
	// we can't ask the OS if it's a TTY — default to JSON.
	var buf bytes.Buffer
	assert.Equal(t, FormatJSON, ResolveWriter(false, &buf))
}

func TestResolve_RedirectedFileIsJSON(t *testing.T) {
	// Regular file (not a TTY) → JSON.
	f, err := os.CreateTemp(t.TempDir(), "stdout")
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	assert.Equal(t, FormatJSON, Resolve(false, f))
}

func TestJSON_PrettyPrintsWithTwoSpaceIndent(t *testing.T) {
	var buf bytes.Buffer

	err := JSON(&buf, map[string]any{"id": "abc", "count": 2})
	require.NoError(t, err)

	got := buf.String()
	// Stable key ordering in encoding/json map output is alphabetical.
	assert.Contains(t, got, "\"count\": 2")
	assert.Contains(t, got, "\"id\": \"abc\"")
	// Trailing newline from encoder.Encode.
	assert.True(t, strings.HasSuffix(got, "\n"))
	// Indented with two spaces.
	assert.Contains(t, got, "\n  \"")
}

func TestTable_AlignsColumns(t *testing.T) {
	var buf bytes.Buffer

	tw := Table(&buf)
	_, _ = tw.Write([]byte("ID\tPARTY\tDATES\n"))
	_, _ = tw.Write([]byte("srch_abc\t2\t2026-05-01 → 2026-05-03\n"))
	_, _ = tw.Write([]byte("srch_def\t12\t2026-05-15\n"))
	require.NoError(t, tw.Flush())

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	require.Len(t, lines, 3)

	// All three lines should be the same width after tab expansion.
	assert.Greater(t, len(lines[0]), 20)
	assert.Equal(t, strings.Index(lines[0], "PARTY"), strings.Index(lines[1], "2"))
}
