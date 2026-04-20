package spot

import (
	"context"
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
	ID               string `json:"id"`
	Name             string `json:"name"`
	Cuisine          string `json:"cuisine,omitempty"`
	Address          string `json:"address,omitempty"`
	Neighborhood     string `json:"neighborhood,omitempty"`
	Zone             string `json:"zone,omitempty"`
	ResyActive       bool   `json:"resyActive"`
	OpenTableActive  bool   `json:"openTableActive"`
	SevenRoomsActive bool   `json:"sevenRoomsActive"`
	DoorDashActive   bool   `json:"doorDashActive"`
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
