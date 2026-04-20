package spot

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestaurantsService_Search(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/restaurants/search", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "gramercy", got["query"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"restaurants": [
				{"id": "rst_abc", "name": "Gramercy Tavern", "platform": "resy", "zone": "NYC"},
				{"id": "rst_def", "name": "Gramercy Park Hotel", "platform": "opentable", "zone": "NYC"}
			]
		}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	results, err := c.Restaurants.Search(context.Background(), "gramercy")
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "rst_abc", results[0].ID)
	assert.Equal(t, "Gramercy Tavern", results[0].Name)
	assert.Equal(t, "resy", results[0].Platform)
	assert.Equal(t, "NYC", results[0].Zone)
}

func TestRestaurantsService_Search_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"restaurants":[]}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	results, err := c.Restaurants.Search(context.Background(), "nowhere")
	require.NoError(t, err)
	assert.Empty(t, results)
}
