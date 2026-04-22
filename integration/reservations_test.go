//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/spot-nyc/spot"
)

// TestIntegration_ReservationsList confirms the upcoming-reservations
// endpoint responds without error. Content depends on the test user's state
// so we don't assert on specific entries — just that the response decodes
// and the shape matches our struct.
func TestIntegration_ReservationsList(t *testing.T) {
	client := requireClient(t)
	_, err := client.Reservations.List(context.Background())
	require.NoError(t, err, "list should succeed regardless of whether user has upcoming reservations")
}

// TestIntegration_ReservationsHistory confirms the full-log endpoint
// responds without error. Same shape rationale as List.
func TestIntegration_ReservationsHistory(t *testing.T) {
	client := requireClient(t)
	_, err := client.Reservations.History(context.Background())
	require.NoError(t, err, "history should succeed regardless of content")
}

// TestIntegration_ReservationsSearch exercises the search endpoint with a
// real restaurant. Whether slots come back depends on availability, so we
// only assert the call itself succeeded.
func TestIntegration_ReservationsSearch_Roundtrip(t *testing.T) {
	client := requireClient(t)
	ctx := context.Background()

	restaurants, err := client.Restaurants.Search(ctx, searchProbe)
	require.NoError(t, err)
	require.NotEmpty(t, restaurants)

	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	_, err = client.Reservations.Search(ctx, &spot.SearchReservationsParams{
		RestaurantIDs: []string{restaurants[0].ID},
		Date:          tomorrow,
		StartTime:     "18:00:00",
		EndTime:       "21:00:00",
		Party:         2,
	})
	require.NoError(t, err)
}
