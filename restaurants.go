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
// need them; missing fields are omitted from JSON.
type Restaurant struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Platform string `json:"platform,omitempty"`
	Zone     string `json:"zone,omitempty"`
	Cuisine  string `json:"cuisine,omitempty"`
	Address  string `json:"address,omitempty"`
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
