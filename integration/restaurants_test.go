//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spot-nyc/spot"
)

// A common NYC-restaurant query guaranteed to match several entries. If
// upstream data ever changes such that this query returns zero results,
// swap it for something broader like "restaurant" — but "gramercy" is
// specific enough to stay meaningful.
const searchProbe = "gramercy"

func TestIntegration_RestaurantsSearch_ReturnsResults(t *testing.T) {
	client := requireClient(t)

	results, err := client.Restaurants.Search(context.Background(), searchProbe)
	require.NoError(t, err)
	require.NotEmpty(t, results, "search %q should return at least one restaurant", searchProbe)

	first := results[0]
	assert.NotEmpty(t, first.ID, "restaurant must have an ID")
	assert.NotEmpty(t, first.Name, "restaurant must have a name")
}

func TestIntegration_RestaurantsGet_MatchesSearchResult(t *testing.T) {
	client := requireClient(t)
	ctx := context.Background()

	results, err := client.Restaurants.Search(ctx, searchProbe)
	require.NoError(t, err)
	require.NotEmpty(t, results)

	probe := results[0]
	detail, err := client.Restaurants.Get(ctx, probe.ID)
	require.NoError(t, err)
	require.NotNil(t, detail)

	assert.Equal(t, probe.ID, detail.ID)
	assert.Equal(t, probe.Name, detail.Name)
	// Get returns a superset of search fields; we don't assert every field,
	// just that the basic identity matches.
}

func TestIntegration_RestaurantsDiscover_ReturnsRankedResults(t *testing.T) {
	client := requireClient(t)

	response, err := client.Restaurants.Discover(context.Background(), &spot.RestaurantSearchParams{
		Neighborhood: "east village",
		Market:       "new-york",
		Limit:        3,
	})
	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotEmpty(t, response.Results, "discovery should return East Village restaurants")

	first := response.Results[0]
	assert.NotEmpty(t, first.Restaurant.ID, "restaurant must have an ID")
	assert.NotEmpty(t, first.Restaurant.Name, "restaurant must have a name")
	assert.GreaterOrEqual(t, first.Score, 0.0, "discovery score should be non-negative")
	assert.True(t,
		first.Restaurant.ResyActive || first.Restaurant.OpenTableActive || first.Restaurant.SevenRoomsActive || first.Restaurant.DoorDashActive,
		"discovered restaurant should be bookable on at least one platform",
	)
}
