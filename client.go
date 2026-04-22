package spot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
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
	Users        *UsersService
	Searches     *SearchesService
	Restaurants  *RestaurantsService
	Reservations *ReservationsService
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
	c.Searches = &SearchesService{client: c}
	c.Restaurants = &RestaurantsService{client: c}
	c.Reservations = &ReservationsService{client: c}
	return c, nil
}

// BaseURL returns the base URL the client is configured to talk to.
func (c *Client) BaseURL() string { return c.baseURL }

// UserAgent returns the User-Agent string the client will send.
func (c *Client) UserAgent() string { return c.userAgent }

// do executes an authenticated HTTP request against the Spot API.
//
// method is the HTTP verb, path is the URL path (joined with the configured
// base URL), body is optional and will be JSON-encoded if non-nil, and out is
// an optional destination for JSON decoding of the response.
//
// Successful responses (2xx) decode into out if non-nil. Error responses are
// mapped to *Error via mapErrorResponse.
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	tok, err := c.tokenSource.Token()
	if err != nil || tok == nil || tok.AccessToken == "" {
		return ErrUnauthenticated
	}

	var bodyReader io.Reader
	if body != nil {
		buf, jerr := json.Marshal(body)
		if jerr != nil {
			return fmt.Errorf("spot: encode request body: %w", jerr)
		}
		bodyReader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("spot: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("spot: http request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if out == nil || resp.StatusCode == http.StatusNoContent {
			return nil
		}
		if derr := json.NewDecoder(resp.Body).Decode(out); derr != nil {
			return fmt.Errorf("spot: decode response: %w", derr)
		}
		return nil
	}

	return mapErrorResponse(resp)
}

// apiErrorBody matches the JSON shape the Spot API uses for responses it
// writes explicitly (e.g. 412 from /reservations/book, carrying a `platform`
// field). Some error paths return plain text instead of JSON —
// mapErrorResponse falls back to reading the raw body for those.
type apiErrorBody struct {
	Error    string `json:"error"`
	Platform string `json:"platform,omitempty"`
}

// mapErrorResponse translates a non-2xx API response into a *Error backed by
// one of the sentinel codes. The body supplies the message (parsed as JSON
// when possible, otherwise read as plain text); the HTTP status code is
// always recorded in Error.HTTPStatus.
func mapErrorResponse(resp *http.Response) error {
	raw, _ := io.ReadAll(resp.Body)

	var body apiErrorBody
	_ = json.Unmarshal(raw, &body)
	msg := body.Error
	if msg == "" {
		msg = strings.TrimSpace(string(raw))
	}

	var code string
	switch {
	case resp.StatusCode == http.StatusUnauthorized, resp.StatusCode == http.StatusForbidden:
		code = ErrUnauthenticated.Code
	case resp.StatusCode == http.StatusNotFound:
		code = "not_found"
	case resp.StatusCode == http.StatusConflict:
		code = ErrConflict.Code
	case resp.StatusCode == http.StatusGone:
		code = ErrSlotExpired.Code
	case resp.StatusCode == http.StatusPreconditionFailed:
		code = ErrPlatformNotConnected.Code
	case resp.StatusCode == http.StatusBadRequest, resp.StatusCode == http.StatusUnprocessableEntity:
		code = ErrValidation.Code
	case resp.StatusCode == http.StatusTooManyRequests:
		code = ErrRateLimited.Code
	case resp.StatusCode >= 500:
		code = ErrServer.Code
	default:
		code = "http_error"
	}

	if msg == "" {
		msg = http.StatusText(resp.StatusCode)
	}

	err := &Error{
		Code:       code,
		Message:    msg,
		HTTPStatus: resp.StatusCode,
	}
	if code == ErrPlatformNotConnected.Code {
		err.Platform = body.Platform
	}
	return err
}
