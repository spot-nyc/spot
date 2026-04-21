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
	StartDate     string         `json:"startDate"`
	EndDate       string         `json:"endDate"`
	StartTime     string         `json:"startTime"`
	EndTime       string         `json:"endTime"`
	Party         int            `json:"party"`
	CreatedAt     string         `json:"createdAt,omitempty"`
	UpdatedAt     string         `json:"updatedAt,omitempty"`
	SearchTargets []SearchTarget `json:"searchTargets,omitempty"`
}

// SearchTarget is a restaurant attached to a search. search_targets has a
// compound primary key of (searchId, restaurantId), so the target itself has
// no surface-level ID worth exposing.
type SearchTarget struct {
	Restaurant *Restaurant `json:"restaurant,omitempty"`
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

// CreateSearchParams holds the inputs for SearchesService.Create.
//
// Party, StartDate, EndDate, StartTime, EndTime, and Restaurants are all required.
// Time fields use "HH:MM:SS" format (the server accepts HH:MM:SS).
// Dates use "YYYY-MM-DD" format.
//
// Note: there is no Upgrade field — morty rejects upgrade on create. Enable
// upgrade mode via SearchesService.Update (Tier 2, M6).
type CreateSearchParams struct {
	Party       int      `json:"party"`
	StartDate   string   `json:"startDate"`
	EndDate     string   `json:"endDate"`
	StartTime   string   `json:"startTime"`
	EndTime     string   `json:"endTime"`
	Restaurants []string `json:"restaurants"`
}

// Create creates a new reservation search.
func (s *SearchesService) Create(ctx context.Context, params *CreateSearchParams) (*Search, error) {
	var resp searchDetailResponse
	if err := s.client.do(ctx, http.MethodPost, "/searches", params, &resp); err != nil {
		return nil, err
	}
	return &resp.Search, nil
}
