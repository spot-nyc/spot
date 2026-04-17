package spot

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestNewClient_DefaultBaseURL(t *testing.T) {
	c, err := NewClient(WithToken("test-token"))
	require.NoError(t, err)
	assert.Equal(t, "https://api.spot.nyc", c.BaseURL())
}

func TestNewClient_OverrideBaseURL(t *testing.T) {
	c, err := NewClient(
		WithToken("test-token"),
		WithBaseURL("http://localhost:8080"),
	)
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8080", c.BaseURL())
}

func TestNewClient_DefaultUserAgent(t *testing.T) {
	c, err := NewClient(WithToken("test-token"))
	require.NoError(t, err)
	assert.Contains(t, c.UserAgent(), "spot-sdk-go/")
}

func TestNewClient_OverrideUserAgent(t *testing.T) {
	c, err := NewClient(
		WithToken("test-token"),
		WithUserAgent("my-app/1.0"),
	)
	require.NoError(t, err)
	assert.Equal(t, "my-app/1.0", c.UserAgent())
}

func TestNewClient_RequiresTokenSource(t *testing.T) {
	_, err := NewClient()
	assert.Error(t, err, "NewClient should fail when no token source is configured")
}

func TestNewClient_WithTokenSource(t *testing.T) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "abc"})
	c, err := NewClient(WithTokenSource(ts))
	require.NoError(t, err)
	tok, err := c.tokenSource.Token()
	require.NoError(t, err)
	assert.Equal(t, "abc", tok.AccessToken)
}

func TestNewClient_DefaultHTTPClientHasTimeout(t *testing.T) {
	c, err := NewClient(WithToken("test-token"))
	require.NoError(t, err)
	assert.NotNil(t, c.httpClient)
	assert.Equal(t, 30*time.Second, c.httpClient.Timeout)
}

func TestClient_do_GET_SuccessfulResponse(t *testing.T) {
	type response struct {
		Name string `json:"name"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/ping", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Contains(t, r.Header.Get("User-Agent"), "spot-sdk-go/")
		assert.Equal(t, "application/json", r.Header.Get("Accept"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"name":"hello"}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	var out response
	err = c.do(context.Background(), http.MethodGet, "/ping", nil, &out)
	require.NoError(t, err)
	assert.Equal(t, "hello", out.Name)
}

func TestClient_do_POST_EncodesRequestBody(t *testing.T) {
	type payload struct {
		Greeting string `json:"greeting"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, _ := io.ReadAll(r.Body)
		var got payload
		_ = json.Unmarshal(body, &got)
		assert.Equal(t, "hi", got.Greeting)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	err = c.do(context.Background(), http.MethodPost, "/greet", payload{Greeting: "hi"}, nil)
	require.NoError(t, err)
}

func TestClient_do_UnauthenticatedWhenNoToken(t *testing.T) {
	ts := emptyTokenSource{}
	c, err := NewClient(WithTokenSource(ts), WithBaseURL("http://unreachable.example"))
	require.NoError(t, err)

	err = c.do(context.Background(), http.MethodGet, "/", nil, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthenticated)
}

// emptyTokenSource is a test helper that mimics a store with no credentials.
type emptyTokenSource struct{}

func (emptyTokenSource) Token() (*oauth2.Token, error) {
	return nil, errors.New("no credentials")
}
