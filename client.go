package spot

import (
	"errors"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

// DefaultBaseURL is the production API base URL.
const DefaultBaseURL = "https://api.spot.nyc"

// Client is the top-level Spot SDK client. Instantiate with NewClient.
type Client struct {
	baseURL     string
	userAgent   string
	httpClient  *http.Client
	tokenSource oauth2.TokenSource

	// Services
	Users *UsersService
}

// NewClient constructs a Client. A token source is required; provide one via
// WithTokenSource or the WithToken shortcut.
func NewClient(opts ...Option) (*Client, error) {
	o := &clientOptions{
		baseURL:   DefaultBaseURL,
		userAgent: "spot-sdk-go/" + Version,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(o)
	}

	if o.tokenSource == nil {
		return nil, errors.New("spot: NewClient requires a token source; use WithTokenSource or WithToken")
	}

	c := &Client{
		baseURL:     o.baseURL,
		userAgent:   o.userAgent,
		httpClient:  o.httpClient,
		tokenSource: o.tokenSource,
	}
	c.Users = &UsersService{client: c}
	return c, nil
}

// BaseURL returns the base URL the client is configured to talk to.
func (c *Client) BaseURL() string { return c.baseURL }

// UserAgent returns the User-Agent string the client will send.
func (c *Client) UserAgent() string { return c.userAgent }
