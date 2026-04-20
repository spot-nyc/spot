package spot

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReservationsService_List(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/searches/bookings", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"reservations": [
				{
					"id": "rsv_abc",
					"userId": "u1",
					"table": {
						"id": "tbl_1",
						"date": "2026-05-01",
						"time": "19:00:00",
						"party": 2,
						"restaurant": {"id": "rst_abc", "name": "Gramercy Tavern"}
					}
				},
				{
					"id": "rsv_def",
					"userId": "u1",
					"cancelled": false,
					"table": {
						"id": "tbl_2",
						"date": "2026-05-15",
						"time": "20:30:00",
						"party": 4,
						"restaurant": {"id": "rst_def", "name": "Shuko"}
					}
				}
			]
		}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	reservations, err := c.Reservations.List(context.Background())
	require.NoError(t, err)
	require.Len(t, reservations, 2)

	assert.Equal(t, "rsv_abc", reservations[0].ID)
	assert.Equal(t, "2026-05-01", reservations[0].Table.Date)
	assert.Equal(t, "19:00:00", reservations[0].Table.Time)
	assert.Equal(t, 2, reservations[0].Table.Party)
	require.NotNil(t, reservations[0].Table.Restaurant)
	assert.Equal(t, "Gramercy Tavern", reservations[0].Table.Restaurant.Name)

	assert.Equal(t, "rsv_def", reservations[1].ID)
	assert.False(t, reservations[1].Cancelled)
}

func TestReservationsService_Cancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/reservations/rsv_abc/cancel", r.URL.Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	err = c.Reservations.Cancel(context.Background(), "rsv_abc")
	require.NoError(t, err)
}

func TestReservationsService_Cancel_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":"Reservation not found"}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	err = c.Reservations.Cancel(context.Background(), "rsv_missing")
	require.Error(t, err)
}
