package spot

import (
	"context"
	"net/http"
)

// ReservationsService handles the /reservations endpoints.
type ReservationsService struct {
	client *Client
}

// Reservation represents a booked table.
type Reservation struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	Cancelled bool   `json:"cancelled,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	Table     Table  `json:"table"`
}

// Table is the specific reservation slot held by a Reservation.
type Table struct {
	ID           string      `json:"id"`
	RestaurantID string      `json:"restaurantId,omitempty"`
	Date         string      `json:"date"`
	Time         string      `json:"time"`
	Party        int         `json:"party"`
	Seating      string      `json:"seating,omitempty"`
	Restaurant   *Restaurant `json:"restaurant,omitempty"`
}

// reservationsListResponse matches the {"reservations": [...]} envelope.
type reservationsListResponse struct {
	Reservations []Reservation `json:"reservations"`
}

// List returns the authenticated user's upcoming reservations.
func (s *ReservationsService) List(ctx context.Context) ([]Reservation, error) {
	var resp reservationsListResponse
	if err := s.client.do(ctx, http.MethodGet, "/searches/bookings", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Reservations, nil
}

// Cancel cancels an upcoming reservation by ID. Succeeds on 2xx responses.
func (s *ReservationsService) Cancel(ctx context.Context, id string) error {
	return s.client.do(ctx, http.MethodPost, "/reservations/"+id+"/cancel", nil, nil)
}
