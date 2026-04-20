package spot

import (
	"context"
	"net/http"
)

// SearchesService handles the /searches endpoints.
type SearchesService struct {
	client *Client
}

// Search is a user's reservation search, as returned by the API.
type Search struct {
	ID            string         `json:"id"`
	UserID        string         `json:"userId"`
	Party         int            `json:"party"`
	StartDate     string         `json:"startDate"`
	EndDate       string         `json:"endDate"`
	StartTime     string         `json:"startTime"`
	EndTime       string         `json:"endTime"`
	Upgrade       bool           `json:"upgrade"`
	CreatedAt     string         `json:"createdAt,omitempty"`
	UpdatedAt     string         `json:"updatedAt,omitempty"`
	SearchTargets []SearchTarget `json:"searchTargets,omitempty"`
}

// SearchTarget is a restaurant attached to a search with a rank.
type SearchTarget struct {
	ID         string      `json:"id"`
	Rank       int         `json:"rank"`
	Restaurant *Restaurant `json:"restaurant,omitempty"`
}

// Restaurant is a restaurant the SDK cares to reference.
// Fields will grow as more commands need them.
type Restaurant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// searchesListResponse matches the {"searches": [...]} envelope returned by the API.
type searchesListResponse struct {
	Searches []Search `json:"searches"`
}

// searchDetailResponse matches the {"search": {...}} envelope for single-search reads.
type searchDetailResponse struct {
	Search Search `json:"search"`
}

// List returns the authenticated user's active searches.
func (s *SearchesService) List(ctx context.Context) ([]Search, error) {
	var resp searchesListResponse
	if err := s.client.do(ctx, http.MethodGet, "/searches/active", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Searches, nil
}

// Get fetches a single search by ID.
func (s *SearchesService) Get(ctx context.Context, id string) (*Search, error) {
	var resp searchDetailResponse
	if err := s.client.do(ctx, http.MethodGet, "/searches/"+id, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Search, nil
}

// Delete removes a search by ID. Succeeds on 2xx responses (including 204).
func (s *SearchesService) Delete(ctx context.Context, id string) error {
	return s.client.do(ctx, http.MethodDelete, "/searches/"+id, nil, nil)
}
