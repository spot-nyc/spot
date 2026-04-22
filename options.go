package spot

import (
	"net/http"

	"golang.org/x/oauth2"
)

// Option configures a Client. Options are applied in order by NewClient.
type Option func(*clientOptions)

// clientOptions holds the mutable config built up by Option values before
// NewClient finalizes it into a Client.
type clientOptions struct {
	baseURL     string
	userAgent   string
	httpClient  *http.Client
	tokenSource oauth2.TokenSource
}

// WithBaseURL overrides the Spot API base URL. Default: https://api.spot.nyc.
func WithBaseURL(url string) Option {
	return func(o *clientOptions) {
		o.baseURL = url
	}
}

// WithUserAgent sets the User-Agent header for outbound requests.
// Default: "spot-sdk-go/<Version>".
func WithUserAgent(ua string) Option {
	return func(o *clientOptions) {
		o.userAgent = ua
	}
}

// WithHTTPClient supplies a custom *http.Client. Useful for tests, tracing,
// or custom timeouts. Default: &http.Client{Timeout: 30 * time.Second}.
func WithHTTPClient(c *http.Client) Option {
	return func(o *clientOptions) {
		o.httpClient = c
	}
}

// WithTokenSource supplies a token source. This is the usual auth path.
func WithTokenSource(ts oauth2.TokenSource) Option {
	return func(o *clientOptions) {
		o.tokenSource = ts
	}
}

// WithToken is a shortcut for static token use (e.g., tests, scripts with a
// manually-issued token). Wraps the token in an oauth2.StaticTokenSource.
func WithToken(token string) Option {
	return func(o *clientOptions) {
		o.tokenSource = oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	}
}
