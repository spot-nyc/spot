//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
