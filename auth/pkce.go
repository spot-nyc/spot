package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

// ============================================================================
// PKCE primitives
// ============================================================================

// newCodeVerifier returns a PKCE code verifier: 32 cryptographically-random
// bytes, base64url-encoded (producing a 43-character string, within the
// RFC 7636 43-128 char range).
func newCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("auth: generate code verifier: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// codeChallenge computes the S256 challenge for a verifier:
// base64url(SHA-256(verifier)), no padding.
func codeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// newState returns a CSRF state parameter: 16 cryptographically-random bytes,
// base64url-encoded.
func newState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("auth: generate state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// ============================================================================
// Loopback server
// ============================================================================

// tryBindLoopback attempts to bind a TCP listener on one of the pre-registered
// LoopbackPorts. Returns the listener and chosen port on success.
func tryBindLoopback() (net.Listener, int, error) {
	for _, port := range LoopbackPorts {
		lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			return lis, port, nil
		}
	}
	return nil, 0, errors.New("auth: all pre-registered loopback ports are in use; close one and retry")
}

// ============================================================================
// Login flow
// ============================================================================

// LoginOptions tunes the Login flow.
type LoginOptions struct {
	// Opener is called with the authorize URL. It typically opens the system
	// browser. Defaults to browser.OpenURL.
	Opener func(url string) error

	// Timeout bounds the total login wait (from browser open to callback).
	// Defaults to 5 minutes.
	Timeout time.Duration
}

// DefaultLogin runs the PKCE login flow against the default Spot OAuth
// configuration (from DefaultOAuthConfig).
func DefaultLogin(ctx context.Context, opts LoginOptions) (Credentials, error) {
	return Login(ctx, DefaultOAuthConfig(), opts)
}

// Login runs the PKCE authorization code flow:
//
//  1. Generate code verifier, code challenge, and CSRF state.
//  2. Bind a loopback listener on the first available pre-registered port.
//  3. Build the authorize URL and open it in the browser via opts.Opener.
//  4. Wait for the callback on the loopback listener (or error/timeout).
//  5. Validate state, extract the authorization code.
//  6. Exchange the code for tokens via the OAuth token endpoint.
//  7. Return Credentials ready to persist.
func Login(ctx context.Context, cfg *oauth2.Config, opts LoginOptions) (Credentials, error) {
	verifier, err := newCodeVerifier()
	if err != nil {
		return Credentials{}, err
	}
	challenge := codeChallenge(verifier)
	state, err := newState()
	if err != nil {
		return Credentials{}, err
	}

	lis, port, err := tryBindLoopback()
	if err != nil {
		return Credentials{}, err
	}
	defer func() { _ = lis.Close() }()

	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	// Scoped copy of the oauth2.Config with the selected redirect URI.
	effectiveCfg := *cfg
	effectiveCfg.RedirectURL = redirectURI

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	server := &http.Server{
		ReadHeaderTimeout: 5 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/callback" {
				http.NotFound(w, r)
				return
			}

			q := r.URL.Query()

			if errCode := q.Get("error"); errCode != "" {
				desc := q.Get("error_description")
				msg := fmt.Sprintf("auth: %s", errCode)
				if desc != "" {
					msg += ": " + desc
				}
				errCh <- errors.New(msg)
				http.Error(w, msg, http.StatusBadRequest)
				return
			}

			if got := q.Get("state"); got != state {
				errCh <- errors.New("auth: state mismatch; possible CSRF")
				http.Error(w, "state mismatch", http.StatusBadRequest)
				return
			}

			code := q.Get("code")
			if code == "" {
				errCh <- errors.New("auth: callback missing code")
				http.Error(w, "missing code", http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = io.WriteString(w, successHTML)
			codeCh <- code
		}),
	}

	go func() {
		_ = server.Serve(lis)
	}()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	authURL := effectiveCfg.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	opener := opts.Opener
	if opener == nil {
		opener = browser.OpenURL
	}
	if err := opener(authURL); err != nil {
		return Credentials{}, fmt.Errorf("auth: open browser: %w", err)
	}

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	select {
	case <-ctx.Done():
		return Credentials{}, ctx.Err()
	case err := <-errCh:
		return Credentials{}, err
	case <-time.After(timeout):
		return Credentials{}, fmt.Errorf("auth: login timed out after %v", timeout)
	case code := <-codeCh:
		tok, err := effectiveCfg.Exchange(ctx, code,
			oauth2.SetAuthURLParam("code_verifier", verifier),
		)
		if err != nil {
			return Credentials{}, fmt.Errorf("auth: token exchange: %w", err)
		}
		return CredentialsFromToken(tok), nil
	}
}

// successHTML is the page the browser shows after a successful callback.
const successHTML = `<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>Signed in to Spot</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
           max-width: 480px; margin: 80px auto; padding: 0 24px; color: #111; }
    h1 { font-size: 20px; margin-bottom: 8px; }
    p { color: #555; line-height: 1.5; }
  </style>
</head>
<body>
  <h1>Signed in to Spot</h1>
  <p>You can close this tab and return to your terminal.</p>
</body>
</html>`
