package spot

import (
	"context"
	"errors"
	"net/http"
)

// RestaurantsService handles the /restaurants endpoints.
type RestaurantsService struct {
	client *Client
}

// Restaurant is a restaurant in the Spot catalog. Fields grow as more commands
// need them; missing string fields are omitted from JSON. The "*Active"
// booleans indicate which booking platforms currently monitor this restaurant.
type Restaurant struct {
	ID                       string `json:"id"`
	Name                     string `json:"name"`
	Cuisine                  string `json:"cuisine,omitempty"`
	Address                  string `json:"address,omitempty"`
	Neighborhood             string `json:"neighborhood,omitempty"`
	Zone                     string `json:"zone,omitempty"`
	Hours                    string `json:"hours,omitempty"`
	Phone                    string `json:"phone,omitempty"`
	Website                  string `json:"website,omitempty"`
	ResyURL                  string `json:"resyUrl,omitempty"`
	OpenTableURL             string `json:"openTableUrl,omitempty"`
	SevenRoomsURL            string `json:"sevenRoomsUrl,omitempty"`
	DoorDashURL              string `json:"doorDashUrl,omitempty"`
	MinimumPartySize         int    `json:"minimumPartySize,omitempty"`
	MaximumPartySize         int    `json:"maximumPartySize,omitempty"`
	BookingDifficulty        int    `json:"bookingDifficulty,omitempty"`
	BookingDifficultyDetails string `json:"bookingDifficultyDetails,omitempty"`
	ResyActive               bool   `json:"resyActive"`
	OpenTableActive          bool   `json:"openTableActive"`
	SevenRoomsActive         bool   `json:"sevenRoomsActive"`
	DoorDashActive           bool   `json:"doorDashActive"`
}

// Platforms returns the display names of the booking platforms currently
// active for this restaurant, in a stable order.
func (r Restaurant) Platforms() []string {
	platforms := make([]string, 0, 4)
	if r.ResyActive {
		platforms = append(platforms, "Resy")
	}
	if r.OpenTableActive {
		platforms = append(platforms, "OpenTable")
	}
	if r.SevenRoomsActive {
		platforms = append(platforms, "SevenRooms")
	}
	if r.DoorDashActive {
		platforms = append(platforms, "DoorDash")
	}
	return platforms
}

type restaurantsSearchRequest struct {
	Query string `json:"query"`
}

type restaurantsSearchResponse struct {
	Restaurants []Restaurant `json:"restaurants"`
}

// Search returns restaurants matching the query string.
func (s *RestaurantsService) Search(ctx context.Context, query string) ([]Restaurant, error) {
	var resp restaurantsSearchResponse
	body := restaurantsSearchRequest{Query: query}
	if err := s.client.do(ctx, http.MethodPost, "/restaurants/search", body, &resp); err != nil {
		return nil, err
	}
	return resp.Restaurants, nil
}

type restaurantsGetResponse struct {
	Restaurant Restaurant `json:"restaurant"`
}

// Get fetches a single restaurant by ID. Returns ErrRestaurantNotFound if no
// active restaurant matches.
func (s *RestaurantsService) Get(ctx context.Context, id string) (*Restaurant, error) {
	var resp restaurantsGetResponse
	if err := s.client.do(ctx, http.MethodGet, "/restaurants/"+id, nil, &resp); err != nil {
		var spotErr *Error
		if errors.As(err, &spotErr) && spotErr.HTTPStatus == http.StatusNotFound {
			spotErr.Code = ErrRestaurantNotFound.Code
		}
		return nil, err
	}
	return &resp.Restaurant, nil
}
