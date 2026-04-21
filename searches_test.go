package spot

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchesService_List(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/searches/active", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"searches": [
				{
					"id": "srch_abc",
					"userId": "user-1",
					"party": 2,
					"startDate": "2026-05-01",
					"endDate": "2026-05-03",
					"startTime": "18:00:00",
					"endTime": "21:00:00",
					"searchTargets": [
						{"restaurant": {"id": "rst_a", "name": "Gramercy Tavern"}}
					]
				},
				{
					"id": "srch_def",
					"userId": "user-1",
					"party": 4,
					"startDate": "2026-05-15",
					"endDate": "2026-05-15",
					"startTime": "19:00:00",
					"endTime": "20:00:00",
					"searchTargets": []
				}
			]
		}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	searches, err := c.Searches.List(context.Background())
	require.NoError(t, err)
	require.Len(t, searches, 2)

	assert.Equal(t, "srch_abc", searches[0].ID)
	assert.Equal(t, 2, searches[0].Party)
	assert.Equal(t, "2026-05-01", searches[0].StartDate)
	require.Len(t, searches[0].SearchTargets, 1)
	assert.Equal(t, "Gramercy Tavern", searches[0].SearchTargets[0].Restaurant.Name)

	assert.Equal(t, "srch_def", searches[1].ID)
	assert.Empty(t, searches[1].SearchTargets)
}

func TestSearchesService_List_Unauthenticated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":"Invalid or expired token"}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	_, err = c.Searches.List(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthenticated)
}

func TestSearchesService_Get(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/searches/srch_abc", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"search": {
				"id": "srch_abc",
				"userId": "u1",
				"party": 2,
				"startDate": "2026-05-01",
				"endDate": "2026-05-01",
				"startTime": "18:00:00",
				"endTime": "21:00:00",
				"searchTargets": [
					{"restaurant": {"id": "r1", "name": "Gramercy Tavern"}}
				]
			}
		}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	search, err := c.Searches.Get(context.Background(), "srch_abc")
	require.NoError(t, err)
	assert.Equal(t, "srch_abc", search.ID)
	assert.Equal(t, 2, search.Party)
	require.Len(t, search.SearchTargets, 1)
	assert.Equal(t, "Gramercy Tavern", search.SearchTargets[0].Restaurant.Name)
}

func TestSearchesService_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":"Search not found"}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	_, err = c.Searches.Get(context.Background(), "missing")
	require.Error(t, err)
	var spotErr *Error
	require.True(t, errors.As(err, &spotErr))
	assert.Equal(t, "not_found", spotErr.Code)
	assert.Equal(t, http.StatusNotFound, spotErr.HTTPStatus)
}

func TestSearchesService_Delete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/searches/srch_abc", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	err = c.Searches.Delete(context.Background(), "srch_abc")
	require.NoError(t, err)
}

func TestSearchesService_Create(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/searches", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.EqualValues(t, 2, got["party"])
		assert.Equal(t, "2026-05-01", got["startDate"])
		assert.Equal(t, "2026-05-01", got["endDate"])
		assert.Equal(t, "18:00:00", got["startTime"])
		assert.Equal(t, "21:00:00", got["endTime"])
		restaurants, ok := got["restaurants"].([]any)
		require.True(t, ok)
		assert.Len(t, restaurants, 2)
		assert.Equal(t, "rst_abc", restaurants[0])
		assert.Equal(t, "rst_def", restaurants[1])
		// No "upgrade" field in the request body — morty rejects it on create.
		_, hasUpgrade := got["upgrade"]
		assert.False(t, hasUpgrade, "CreateSearchParams must not send upgrade on create")

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"search":{"id":"srch_new","party":2,"startDate":"2026-05-01","endDate":"2026-05-01","startTime":"18:00:00","endTime":"21:00:00"}}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	params := &CreateSearchParams{
		Party:       2,
		StartDate:   "2026-05-01",
		EndDate:     "2026-05-01",
		StartTime:   "18:00:00",
		EndTime:     "21:00:00",
		Restaurants: []string{"rst_abc", "rst_def"},
	}

	created, err := c.Searches.Create(context.Background(), params)
	require.NoError(t, err)
	assert.Equal(t, "srch_new", created.ID)
	assert.Equal(t, 2, created.Party)
}

func TestSearchesService_Create_ValidationError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = io.WriteString(w, `{"error":"party must be positive"}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	_, err = c.Searches.Create(context.Background(), &CreateSearchParams{Party: 0})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrValidation)
}
