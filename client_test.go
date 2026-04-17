package spot

import (
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
