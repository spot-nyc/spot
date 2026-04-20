package spot

import (
	"context"
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
					"upgrade": false,
					"searchTargets": [
						{"id": "tgt_1", "rank": 0, "restaurant": {"id": "rst_a", "name": "Gramercy Tavern"}}
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
					"upgrade": true,
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
	assert.True(t, searches[1].Upgrade)
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
				"upgrade": false,
				"searchTargets": [
					{"id": "t1", "rank": 0, "restaurant": {"id": "r1", "name": "Gramercy Tavern"}}
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
