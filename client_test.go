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

func TestClient_do_Maps401_ToErrUnauthenticated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":"Missing authorization header"}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	err = c.do(context.Background(), http.MethodGet, "/anything", nil, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthenticated)

	var spotErr *Error
	require.True(t, errors.As(err, &spotErr))
	assert.Equal(t, http.StatusUnauthorized, spotErr.HTTPStatus)
	assert.Equal(t, "Missing authorization header", spotErr.Message)
}

func TestClient_do_Maps409_ToErrConflict(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = io.WriteString(w, `{"error":"already booked"}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	err = c.do(context.Background(), http.MethodPost, "/", nil, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConflict)
}

func TestClient_do_Maps422_ToErrValidation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = io.WriteString(w, `{"error":"invalid party size"}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	err = c.do(context.Background(), http.MethodPost, "/", nil, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrValidation)
}

func TestClient_do_Maps429_ToErrRateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	err = c.do(context.Background(), http.MethodGet, "/", nil, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRateLimited)
}

func TestClient_do_Maps500_ToErrServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":"Internal server error"}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	err = c.do(context.Background(), http.MethodGet, "/", nil, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrServer)
}

func TestClient_do_404_ReturnsGenericNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":"not found"}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	err = c.do(context.Background(), http.MethodGet, "/", nil, nil)
	require.Error(t, err)
	var spotErr *Error
	require.True(t, errors.As(err, &spotErr))
	assert.Equal(t, "not_found", spotErr.Code)
	assert.Equal(t, http.StatusNotFound, spotErr.HTTPStatus)
}

func TestClient_do_UnparseableErrorBody_StillReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(w, `<html>gateway error</html>`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	err = c.do(context.Background(), http.MethodGet, "/", nil, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrServer)
	var spotErr *Error
	require.True(t, errors.As(err, &spotErr))
	assert.Equal(t, http.StatusBadGateway, spotErr.HTTPStatus)
}

// Hono's HTTPException.getResponse() returns the message as text/plain. The
// SDK must surface that as the error Message rather than falling back to
// http.StatusText.
func TestClient_do_PlainTextErrorBody_PreservesMessage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		w.WriteHeader(http.StatusGone)
		_, _ = io.WriteString(w, "Slot is no longer available")
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	err = c.do(context.Background(), http.MethodGet, "/", nil, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSlotExpired)

	var spotErr *Error
	require.True(t, errors.As(err, &spotErr))
	assert.Equal(t, "Slot is no longer available", spotErr.Message)
}
