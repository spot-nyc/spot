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
						"platform": "resy",
						"date": "2026-05-01",
						"time": "19:00:00",
						"party": 2,
						"seating": "Dining Room",
						"restaurant": {"id": "rst_abc", "name": "Gramercy Tavern"}
					}
				},
				{
					"id": "rsv_def",
					"userId": "u1",
					"cancelled": false,
					"table": {
						"id": "tbl_2",
						"platform": "opentable",
						"date": "2026-05-15",
						"time": "20:30:00",
						"party": 4,
						"seating": "Bar",
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
	assert.Equal(t, "Dining Room", reservations[0].Table.Seating)
	assert.Equal(t, "resy", reservations[0].Table.Platform)
	require.NotNil(t, reservations[0].Table.Restaurant)
	assert.Equal(t, "Gramercy Tavern", reservations[0].Table.Restaurant.Name)

	assert.Equal(t, "rsv_def", reservations[1].ID)
	assert.False(t, reservations[1].Cancelled)
	assert.Equal(t, "Bar", reservations[1].Table.Seating)
	assert.Equal(t, "opentable", reservations[1].Table.Platform)
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

func TestReservationsService_Search(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/reservations/search", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		restaurantIDs, ok := got["restaurantIds"].([]any)
		require.True(t, ok)
		assert.Len(t, restaurantIDs, 2)
		assert.Equal(t, "rst_a", restaurantIDs[0])
		assert.Equal(t, "2026-05-15", got["date"])
		assert.Equal(t, "18:00:00", got["startTime"])
		assert.EqualValues(t, 2, got["party"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"availability": [
				{
					"restaurant": {
						"id": "rst_a",
						"name": "Gramercy Tavern",
						"neighborhood": "Flatiron",
						"cuisine": "American",
						"resyActive": true,
						"openTableActive": false,
						"sevenRoomsActive": false,
						"doorDashActive": false
					},
					"slots": [
						{
							"id": "slot_1",
							"platform": "resy",
							"date": "2026-05-15",
							"time": "19:00:00",
							"party": 2,
							"seating": "Dining Room",
							"restaurantId": "rst_a"
						},
						{
							"id": "slot_2",
							"platform": "resy",
							"date": "2026-05-15",
							"time": "20:00:00",
							"party": 2,
							"seating": "Dining Room",
							"restaurantId": "rst_a"
						}
					]
				},
				{
					"restaurant": {
						"id": "rst_b",
						"name": "Shuko",
						"neighborhood": "Union Square",
						"cuisine": "Japanese",
						"resyActive": false,
						"openTableActive": true,
						"sevenRoomsActive": false,
						"doorDashActive": false
					},
					"slots": [
						{
							"id": "slot_3",
							"platform": "opentable",
							"date": "2026-05-15",
							"time": "19:30:00",
							"party": 2,
							"seating": "default",
							"restaurantId": "rst_b"
						}
					]
				}
			]
		}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	slots, err := c.Reservations.Search(context.Background(), &SearchReservationsParams{
		RestaurantIDs: []string{"rst_a", "rst_b"},
		Date:          "2026-05-15",
		StartTime:     "18:00:00",
		EndTime:       "21:00:00",
		Party:         2,
	})
	require.NoError(t, err)
	require.Len(t, slots, 3)

	assert.Equal(t, "slot_1", slots[0].ID)
	assert.Equal(t, "resy", slots[0].Platform)
	assert.Equal(t, "2026-05-15", slots[0].Date)
	assert.Equal(t, "19:00:00", slots[0].Time)
	assert.Equal(t, "Dining Room", slots[0].Seating)
	assert.Equal(t, "rst_a", slots[0].RestaurantID)
	require.NotNil(t, slots[0].Restaurant)
	assert.Equal(t, "Gramercy Tavern", slots[0].Restaurant.Name)

	assert.Equal(t, "slot_3", slots[2].ID)
	require.NotNil(t, slots[2].Restaurant)
	assert.Equal(t, "Shuko", slots[2].Restaurant.Name)
}

func TestReservationsService_Search_MissingRestaurant_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Morty throws HTTPException, which Hono serializes as text/plain.
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, "Restaurant not found: rst_missing")
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	_, err = c.Reservations.Search(context.Background(), &SearchReservationsParams{
		RestaurantIDs: []string{"rst_missing"},
		Date:          "2026-05-15",
		StartTime:     "18:00:00",
		EndTime:       "21:00:00",
		Party:         2,
	})
	require.Error(t, err)

	var spotErr *Error
	require.True(t, errors.As(err, &spotErr))
	assert.Equal(t, http.StatusNotFound, spotErr.HTTPStatus)
	assert.Equal(t, "Restaurant not found: rst_missing", spotErr.Message)
}

func TestReservationsService_Search_Empty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"availability":[]}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	slots, err := c.Reservations.Search(context.Background(), &SearchReservationsParams{
		RestaurantIDs: []string{"rst_a"},
		Date:          "2026-05-15",
		StartTime:     "18:00:00",
		EndTime:       "21:00:00",
		Party:         2,
	})
	require.NoError(t, err)
	assert.Empty(t, slots)
}

func TestReservationsService_Book(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/reservations/book", r.URL.Path)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		assert.Equal(t, "slot_abc", got["slotId"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"reservation": {
				"id": "rsv_abc",
				"userId": "u1",
				"table": {
					"id": "tbl_abc",
					"platform": "resy",
					"date": "2026-05-15",
					"time": "19:00:00",
					"party": 2,
					"seating": "Dining Room",
					"restaurant": {"id": "rst_a", "name": "Gramercy Tavern"}
				}
			}
		}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	reservation, err := c.Reservations.Book(context.Background(), "slot_abc")
	require.NoError(t, err)
	require.NotNil(t, reservation)
	assert.Equal(t, "rsv_abc", reservation.ID)
	require.NotNil(t, reservation.Table.Restaurant)
	assert.Equal(t, "Gramercy Tavern", reservation.Table.Restaurant.Name)
}

func TestReservationsService_Book_SlotExpired(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusGone)
		_, _ = io.WriteString(w, `{"error":"Slot is no longer available"}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	_, err = c.Reservations.Book(context.Background(), "slot_expired")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSlotExpired)
}

func TestReservationsService_Book_PlatformNotConnected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusPreconditionFailed)
		_, _ = io.WriteString(w, `{"error":"Platform not connected: resy","platform":"resy"}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	_, err = c.Reservations.Book(context.Background(), "slot_abc")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPlatformNotConnected)

	var spotErr *Error
	require.True(t, errors.As(err, &spotErr))
	assert.Equal(t, "resy", spotErr.Platform)
}
