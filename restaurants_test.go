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
				{
					"id": "rst_abc",
					"name": "Gramercy Tavern",
					"neighborhood": "Flatiron",
					"cuisine": "American",
					"phone": "212-477-0777",
					"website": "https://www.gramercytavern.com",
					"minimumPartySize": 1,
					"maximumPartySize": 8,
					"bookingDifficulty": 8,
					"zone": "NYC",
					"resyActive": true,
					"openTableActive": false,
					"sevenRoomsActive": false,
					"doorDashActive": false
				},
				{
					"id": "rst_def",
					"name": "Gramercy Park Hotel",
					"neighborhood": "Gramercy",
					"cuisine": "Hotel",
					"zone": "NYC",
					"resyActive": false,
					"openTableActive": true,
					"sevenRoomsActive": false,
					"doorDashActive": false
				}
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
	assert.Equal(t, "Flatiron", results[0].Neighborhood)
	assert.Equal(t, "American", results[0].Cuisine)
	assert.Equal(t, "NYC", results[0].Zone)
	assert.Equal(t, "212-477-0777", results[0].Phone)
	assert.Equal(t, "https://www.gramercytavern.com", results[0].Website)
	assert.Equal(t, 8, results[0].MaximumPartySize)
	assert.Equal(t, 8, results[0].BookingDifficulty)
	assert.True(t, results[0].ResyActive)
	assert.False(t, results[0].OpenTableActive)
	assert.Equal(t, []string{"Resy"}, results[0].Platforms())
	assert.Equal(t, []string{"OpenTable"}, results[1].Platforms())
}

func TestRestaurant_Platforms(t *testing.T) {
	cases := []struct {
		name       string
		restaurant Restaurant
		want       []string
	}{
		{"none active", Restaurant{}, []string{}},
		{"resy only", Restaurant{ResyActive: true}, []string{"Resy"}},
		{"opentable only", Restaurant{OpenTableActive: true}, []string{"OpenTable"}},
		{"sevenrooms only", Restaurant{SevenRoomsActive: true}, []string{"SevenRooms"}},
		{"doordash only", Restaurant{DoorDashActive: true}, []string{"DoorDash"}},
		{
			"all active",
			Restaurant{ResyActive: true, OpenTableActive: true, SevenRoomsActive: true, DoorDashActive: true},
			[]string{"Resy", "OpenTable", "SevenRooms", "DoorDash"},
		},
		{
			"resy and doordash",
			Restaurant{ResyActive: true, DoorDashActive: true},
			[]string{"Resy", "DoorDash"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.restaurant.Platforms())
		})
	}
}

func TestRestaurantsService_Get(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/restaurants/rst_abc", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"restaurant": {
				"id": "rst_abc",
				"name": "Gramercy Tavern",
				"neighborhood": "Flatiron",
				"cuisine": "American",
				"resyActive": true,
				"openTableActive": false,
				"sevenRoomsActive": false,
				"doorDashActive": false
			}
		}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	restaurant, err := c.Restaurants.Get(context.Background(), "rst_abc")
	require.NoError(t, err)
	require.NotNil(t, restaurant)
	assert.Equal(t, "rst_abc", restaurant.ID)
	assert.Equal(t, "Gramercy Tavern", restaurant.Name)
	assert.Equal(t, []string{"Resy"}, restaurant.Platforms())
}

func TestRestaurantsService_Get_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":"Restaurant not found"}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	_, err = c.Restaurants.Get(context.Background(), "missing")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRestaurantNotFound)
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
