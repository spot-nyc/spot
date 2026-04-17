package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCodeVerifier_LengthAndAlphabet(t *testing.T) {
	v, err := newCodeVerifier()
	require.NoError(t, err)

	// Spec: 43–128 chars, unreserved URL chars (A-Z a-z 0-9 - . _ ~).
	// base64.RawURLEncoding produces A-Z a-z 0-9 - _ (no padding).
	assert.GreaterOrEqual(t, len(v), 43, "verifier should be at least 43 chars")
	assert.LessOrEqual(t, len(v), 128, "verifier should be at most 128 chars")
	for _, r := range v {
		isAlpha := (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
		isDigit := r >= '0' && r <= '9'
		isExtra := r == '-' || r == '_'
		assert.True(t, isAlpha || isDigit || isExtra, "unexpected char %q in verifier", r)
	}
}

func TestNewCodeVerifier_UniquePerCall(t *testing.T) {
	a, err := newCodeVerifier()
	require.NoError(t, err)
	b, err := newCodeVerifier()
	require.NoError(t, err)
	assert.NotEqual(t, a, b, "two verifiers should not collide")
}

func TestCodeChallenge_IsBase64URLEncodedSHA256(t *testing.T) {
	verifier := "abc123"

	got := codeChallenge(verifier)

	want := base64.RawURLEncoding.EncodeToString(sha256Digest([]byte(verifier)))
	assert.Equal(t, want, got)
}

func TestCodeChallenge_KnownVector(t *testing.T) {
	// RFC 7636 Appendix B: verifier "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	// yields challenge "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM".
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	want := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"

	assert.Equal(t, want, codeChallenge(verifier))
}

func TestNewState_UniquePerCall(t *testing.T) {
	a, err := newState()
	require.NoError(t, err)
	b, err := newState()
	require.NoError(t, err)

	assert.NotEqual(t, a, b)
	assert.GreaterOrEqual(t, len(a), 20, "state should be reasonably long")
	assert.False(t, strings.ContainsAny(a, "+/="), "state should be URL-safe (no +/=)")
}

// sha256Digest is a test helper wrapping sha256.Sum256 to return a slice.
func sha256Digest(b []byte) []byte {
	s := sha256.Sum256(b)
	return s[:]
}
