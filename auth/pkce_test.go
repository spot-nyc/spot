package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
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

// Full-flow Login test. Mocks the Supabase token endpoint with httptest, and
// simulates the browser step with a fake opener that directly hits the
// callback URL.
func TestLogin_FullFlow_Success(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.NoError(t, r.ParseForm())

		assert.Equal(t, "authorization_code", r.FormValue("grant_type"))
		assert.Equal(t, "code-from-callback", r.FormValue("code"))
		assert.Equal(t, "test-client", r.FormValue("client_id"))
		assert.NotEmpty(t, r.FormValue("redirect_uri"), "redirect_uri must be echoed")
		assert.NotEmpty(t, r.FormValue("code_verifier"), "code_verifier must be present")

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "access-token-123",
			"refresh_token": "refresh-token-456",
			"token_type":    "bearer",
			"expires_in":    3600,
			"scope":         "phone profile",
		})
	}))
	defer tokenSrv.Close()

	opener := func(rawURL string) error {
		u, err := neturl.Parse(rawURL)
		if err != nil {
			return err
		}
		state := u.Query().Get("state")
		redirectURI := u.Query().Get("redirect_uri")
		if state == "" || redirectURI == "" {
			return fmt.Errorf("expected state and redirect_uri in authorize URL")
		}

		go func() {
			cb := redirectURI + "?code=code-from-callback&state=" + state
			resp, hErr := http.Get(cb)
			if hErr == nil {
				_ = resp.Body.Close()
			}
		}()
		return nil
	}

	cfg := &oauth2.Config{
		ClientID: "test-client",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://example.invalid/authorize",
			TokenURL:  tokenSrv.URL,
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes: []string{"phone", "profile"},
	}

	creds, err := Login(context.Background(), cfg, LoginOptions{
		Opener:  opener,
		Timeout: 10 * time.Second,
	})
	require.NoError(t, err)
	assert.Equal(t, "access-token-123", creds.AccessToken)
	assert.Equal(t, "refresh-token-456", creds.RefreshToken)
}

func TestLogin_StateMismatch(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("token endpoint should NOT be called on state mismatch")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer tokenSrv.Close()

	opener := func(rawURL string) error {
		u, _ := neturl.Parse(rawURL)
		redirectURI := u.Query().Get("redirect_uri")
		go func() {
			cb := redirectURI + "?code=any-code&state=WRONG-STATE"
			resp, hErr := http.Get(cb)
			if hErr == nil {
				_ = resp.Body.Close()
			}
		}()
		return nil
	}

	cfg := &oauth2.Config{
		ClientID: "test-client",
		Endpoint: oauth2.Endpoint{AuthURL: "https://example.invalid/authorize", TokenURL: tokenSrv.URL},
	}

	_, err := Login(context.Background(), cfg, LoginOptions{
		Opener:  opener,
		Timeout: 5 * time.Second,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "state")
}

func TestLogin_ProviderError(t *testing.T) {
	opener := func(rawURL string) error {
		u, _ := neturl.Parse(rawURL)
		redirectURI := u.Query().Get("redirect_uri")
		state := u.Query().Get("state")
		go func() {
			cb := redirectURI + "?error=access_denied&error_description=user+denied&state=" + state
			resp, hErr := http.Get(cb)
			if hErr == nil {
				_ = resp.Body.Close()
			}
		}()
		return nil
	}

	cfg := &oauth2.Config{
		ClientID: "test-client",
		Endpoint: oauth2.Endpoint{AuthURL: "https://example.invalid/authorize", TokenURL: "https://example.invalid/token"},
	}

	_, err := Login(context.Background(), cfg, LoginOptions{
		Opener:  opener,
		Timeout: 5 * time.Second,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "access_denied")
}

func TestLogin_Timeout(t *testing.T) {
	opener := func(string) error {
		// Simulate: browser opens but user never completes.
		return nil
	}

	cfg := &oauth2.Config{
		ClientID: "test-client",
		Endpoint: oauth2.Endpoint{AuthURL: "https://example.invalid/authorize", TokenURL: "https://example.invalid/token"},
	}

	_, err := Login(context.Background(), cfg, LoginOptions{
		Opener:  opener,
		Timeout: 200 * time.Millisecond,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
}
