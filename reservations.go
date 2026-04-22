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

// Table is the specific reservation slot held by a Reservation. Platform is
// the booking platform this table came from (e.g. "resy", "opentable"). The
// nested Restaurant is always populated by the API.
type Table struct {
	ID         string      `json:"id"`
	Date       string      `json:"date"`
	Time       string      `json:"time"`
	Party      int         `json:"party"`
	Seating    string      `json:"seating,omitempty"`
	Platform   string      `json:"platform,omitempty"`
	Restaurant *Restaurant `json:"restaurant,omitempty"`
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

// ReservationSlot is a table that's currently available to book. Slots are
// persisted server-side for a short TTL (around 5 minutes); after that, Book
// returns ErrSlotExpired.
type ReservationSlot struct {
	ID           string      `json:"id"`
	Platform     string      `json:"platform"`
	Date         string      `json:"date"`
	Time         string      `json:"time"`
	Party        int         `json:"party"`
	Seating      string      `json:"seating,omitempty"`
	RestaurantID string      `json:"restaurantId"`
	Restaurant   *Restaurant `json:"restaurant,omitempty"`
}

// SearchReservationsParams are inputs for ReservationsService.Search.
type SearchReservationsParams struct {
	RestaurantIDs []string `json:"restaurantIds"`
	Date          string   `json:"date"`
	StartTime     string   `json:"startTime"`
	EndTime       string   `json:"endTime"`
	Party         int      `json:"party"`
}

type reservationsSearchAvailability struct {
	Restaurant Restaurant        `json:"restaurant"`
	Slots      []ReservationSlot `json:"slots"`
}

type reservationsSearchResponse struct {
	Availability []reservationsSearchAvailability `json:"availability"`
}

// Search returns slots currently available at the requested restaurants
// within the date/time window. Each slot is persisted server-side with a
// short TTL; callers book by passing the slot ID to Book. The returned
// slots are flattened across restaurants, with each slot's Restaurant
// field populated.
func (s *ReservationsService) Search(ctx context.Context, params *SearchReservationsParams) ([]ReservationSlot, error) {
	var resp reservationsSearchResponse
	if err := s.client.do(ctx, http.MethodPost, "/reservations/search", params, &resp); err != nil {
		return nil, err
	}

	var slots []ReservationSlot
	for _, group := range resp.Availability {
		restaurant := group.Restaurant
		for _, slot := range group.Slots {
			slot.Restaurant = &restaurant
			slots = append(slots, slot)
		}
	}
	return slots, nil
}

type reservationsBookRequest struct {
	SlotID string `json:"slotId"`
}

type reservationsBookResponse struct {
	Reservation Reservation `json:"reservation"`
}

// Book books a slot previously returned by Search. Returns ErrSlotExpired if
// the slot's TTL has passed or it was already booked, or
// ErrPlatformNotConnected if the authenticated user hasn't linked their
// credentials for the slot's booking platform.
func (s *ReservationsService) Book(ctx context.Context, slotID string) (*Reservation, error) {
	var resp reservationsBookResponse
	body := reservationsBookRequest{SlotID: slotID}
	if err := s.client.do(ctx, http.MethodPost, "/reservations/book", body, &resp); err != nil {
		return nil, err
	}
	return &resp.Reservation, nil
}
