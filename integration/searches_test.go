//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spot-nyc/spot"
)

// TestIntegration_SearchesCRUD exercises the full create → list → get →
// update → delete lifecycle. Each sub-assertion proves the SDK and server
// agree on a particular mutation's shape. The search is cleaned up at the
// end regardless of outcome so the CI test user never accumulates cruft.
func TestIntegration_SearchesCRUD(t *testing.T) {
	client := requireClient(t)
	ctx := context.Background()

	// Pick a real restaurant ID to avoid hardcoding.
	restaurants, err := client.Restaurants.Search(ctx, searchProbe)
	require.NoError(t, err)
	require.NotEmpty(t, restaurants, "need a valid restaurant ID to create a search")
	targetID := restaurants[0].ID

	// Create a search targeting tomorrow. Using tomorrow keeps us from
	// conflicting with real upcoming reservations the test user may hold.
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	created, err := client.Searches.Create(ctx, &spot.CreateSearchParams{
		Party:         2,
		StartDate:     tomorrow,
		EndDate:       tomorrow,
		StartTime:     "18:00:00",
		EndTime:       "21:00:00",
		RestaurantIDs: []string{targetID},
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.ID)
	t.Cleanup(func() {
		// Best-effort. Delete is soft, so a second call still succeeds.
		_ = client.Searches.Delete(ctx, created.ID)
	})

	assert.Equal(t, 2, created.Party)
	assert.Equal(t, tomorrow, created.StartDate)

	// List should include our new search.
	all, err := client.Searches.List(ctx)
	require.NoError(t, err)
	found := false
	for _, s := range all {
		if s.ID == created.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "newly created search should appear in List")

	// Get returns the same entity with populated search targets.
	got, err := client.Searches.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	require.NotEmpty(t, got.SearchTargets, "Get should include search targets")

	// Update the party size. Only the explicitly set field should apply.
	newParty := 4
	updated, err := client.Searches.Update(ctx, created.ID, &spot.UpdateSearchParams{
		Party: &newParty,
	})
	require.NoError(t, err)
	assert.Equal(t, 4, updated.Party)
	assert.Equal(t, tomorrow, updated.StartDate, "unset fields should not change")

	// Delete and verify the entity is gone.
	err = client.Searches.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = client.Searches.Get(ctx, created.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, spot.ErrSearchNotFound)
}
