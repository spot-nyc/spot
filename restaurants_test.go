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

func TestRestaurantsService_Discover(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/restaurants/search", r.URL.Path)

		query := r.URL.Query()
		assert.Equal(t, "pasta", query.Get("q"))
		assert.Equal(t, "italian", query.Get("cuisine"))
		assert.Equal(t, "flatiron", query.Get("neighborhood"))
		assert.Equal(t, "new-york", query.Get("market"))
		assert.Equal(t, "canonical", query.Get("scope"))
		assert.Equal(t, "distance", query.Get("sort"))
		assert.Equal(t, "5", query.Get("limit"))
		assert.Equal(t, "40.7128", query.Get("lat"))
		assert.Equal(t, "-74.006", query.Get("lon"))
		assert.Equal(t, "1200", query.Get("radiusMeters"))

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{
			"results": [
				{
					"restaurant": {
						"id": "rst_lodi",
						"name": "Lodi",
						"cuisine": "Italian",
						"neighborhood": "Rockefeller Center",
						"address": "1 Rockefeller Plaza",
						"coordinates": {"x": -73.978, "y": 40.758},
						"availability": null,
						"active": true,
						"imageUrl": null,
						"imageUrls": ["https://example.com/lodi.jpg"],
						"thumbnailUrl": null,
						"headerUrl": null,
						"priceTier": "$$$",
						"description": "All-day Italian restaurant",
						"editorial": null,
						"policy": null,
						"email": null,
						"phone": "212-555-0100",
						"website": "https://example.com/lodi",
						"instagram": null,
						"googleId": "google_lodi",
						"googleMapsUrl": "https://maps.example.com/lodi",
						"dishes": [
							{
								"name": "Agnolotti",
								"quote": "Order the agnolotti.",
								"imageUrl": "https://example.com/agnolotti.jpg",
								"standout": true,
								"shouldOrder": true,
								"recommendation": "recommended",
								"attributions": [
									{
										"origin": "editorial",
										"source": "infatuation-nyc",
										"articleUrl": "https://example.com/pasta",
										"articleType": "guide",
										"articleTitle": "Best Pasta",
										"articleDate": "2026-05-01"
									}
								]
							}
						],
						"ratings": [
							{
								"source": "infatuation-nyc",
								"articleUrl": "https://example.com/lodi-review",
								"articleDate": "2026-05-20",
								"label": null,
								"score": 8.2,
								"max": 10
							}
						],
						"mentions": [
							{
								"source": "infatuation-nyc",
								"articleUrl": "https://example.com/pasta",
								"articleType": "guide",
								"articleTitle": "Best Pasta",
								"articleSummary": "Pasta picks",
								"articleDate": "2026-05-01"
							}
						],
						"resyId": "resy_lodi",
						"resyUrl": "https://resy.example.com/lodi",
						"resyActive": true,
						"openTableId": null,
						"openTableUrl": null,
						"openTableActive": false,
						"sevenRoomsId": null,
						"sevenRoomsSlug": null,
						"sevenRoomsUrl": null,
						"sevenRoomsActive": false,
						"doorDashId": null,
						"doorDashUrl": null,
						"doorDashActive": false,
						"minimumPartySize": 1,
						"maximumPartySize": 6,
						"bookingDifficulty": 7,
						"bookingDifficultyDetails": "Popular dinner reservations",
						"depositFeeAmount": null,
						"depositFeeCutoffTime": null,
						"depositFeeCutoffWindow": null,
						"cancellationFeeAmount": null,
						"cancellationFeeCutoffTime": null,
						"cancellationFeeCutoffWindow": null
					},
					"score": 0.98,
					"distanceMeters": 321.5
				}
			],
			"nextCursor": null
		}`)
	}))
	defer srv.Close()

	c, err := NewClient(WithToken("test-token"), WithBaseURL(srv.URL))
	require.NoError(t, err)

	lat := 40.7128
	lon := -74.006
	response, err := c.Restaurants.Discover(context.Background(), &RestaurantSearchParams{
		Q:            "pasta",
		Cuisine:      "italian",
		Neighborhood: "flatiron",
		Market:       "new-york",
		Scope:        RestaurantSearchScopeCanonical,
		Sort:         RestaurantSearchSortDistance,
		Limit:        5,
		Lat:          &lat,
		Lon:          &lon,
		RadiusMeters: 1200,
	})
	require.NoError(t, err)
	require.Len(t, response.Results, 1)
	assert.Nil(t, response.NextCursor)

	result := response.Results[0]
	assert.Equal(t, "rst_lodi", result.Restaurant.ID)
	assert.Equal(t, "Lodi", result.Restaurant.Name)
	assert.Equal(t, "Italian", result.Restaurant.Cuisine)
	assert.Equal(t, "Rockefeller Center", result.Restaurant.Neighborhood)
	assert.Equal(t, "$$$", result.Restaurant.PriceTier)
	assert.Equal(t, "https://example.com/lodi", result.Restaurant.Website)
	assert.Equal(t, []string{"https://example.com/lodi.jpg"}, result.Restaurant.ImageURLs)
	require.NotNil(t, result.Restaurant.Coordinates)
	assert.Equal(t, -73.978, result.Restaurant.Coordinates.X)
	assert.Equal(t, 40.758, result.Restaurant.Coordinates.Y)
	require.Len(t, result.Restaurant.Dishes, 1)
	assert.Equal(t, "Agnolotti", result.Restaurant.Dishes[0].Name)
	require.NotNil(t, result.Restaurant.Dishes[0].Quote)
	assert.Equal(t, "Order the agnolotti.", *result.Restaurant.Dishes[0].Quote)
	require.Len(t, result.Restaurant.Dishes[0].Attributions, 1)
	assert.Equal(t, "Best Pasta", result.Restaurant.Dishes[0].Attributions[0].ArticleTitle)
	require.Len(t, result.Restaurant.Ratings, 1)
	assert.Equal(t, "infatuation-nyc", result.Restaurant.Ratings[0].Source)
	ratingJSON, err := json.Marshal(result.Restaurant.Ratings[0])
	require.NoError(t, err)
	assert.NotContains(t, string(ratingJSON), "articleDate")
	require.Len(t, result.Restaurant.Mentions, 1)
	assert.Equal(t, "Best Pasta", result.Restaurant.Mentions[0].ArticleTitle)
	assert.Equal(t, []string{"Resy"}, result.Restaurant.Platforms())
	require.NotNil(t, result.DistanceMeters)
	assert.Equal(t, 321.5, *result.DistanceMeters)
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
