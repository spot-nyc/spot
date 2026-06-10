package spot

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// RestaurantsService handles the /restaurants endpoints.
type RestaurantsService struct {
	client *Client
}

// Restaurant is a restaurant in the Spot catalog. Fields grow as more commands
// need them; missing string fields are omitted from JSON. The "*Active"
// booleans indicate which booking platforms currently monitor this restaurant.
type Restaurant struct {
	ID                          string                  `json:"id"`
	Name                        string                  `json:"name"`
	Cuisine                     string                  `json:"cuisine,omitempty"`
	Address                     string                  `json:"address,omitempty"`
	Neighborhood                string                  `json:"neighborhood,omitempty"`
	Zone                        string                  `json:"zone,omitempty"`
	Coordinates                 *RestaurantCoordinates  `json:"coordinates,omitempty"`
	Availability                *RestaurantAvailability `json:"availability,omitempty"`
	Active                      bool                    `json:"active"`
	Hours                       string                  `json:"hours,omitempty"`
	ImageURL                    string                  `json:"imageUrl,omitempty"`
	ImageURLs                   []string                `json:"imageUrls,omitempty"`
	ThumbnailURL                string                  `json:"thumbnailUrl,omitempty"`
	HeaderURL                   string                  `json:"headerUrl,omitempty"`
	PriceTier                   string                  `json:"priceTier,omitempty"`
	Description                 string                  `json:"description,omitempty"`
	Editorial                   string                  `json:"editorial,omitempty"`
	Policy                      string                  `json:"policy,omitempty"`
	Email                       string                  `json:"email,omitempty"`
	Phone                       string                  `json:"phone,omitempty"`
	Website                     string                  `json:"website,omitempty"`
	Instagram                   string                  `json:"instagram,omitempty"`
	GoogleID                    string                  `json:"googleId,omitempty"`
	GoogleMapsURL               string                  `json:"googleMapsUrl,omitempty"`
	Ratings                     []RestaurantRating      `json:"ratings,omitempty"`
	Mentions                    []RestaurantMention     `json:"mentions,omitempty"`
	ResyID                      string                  `json:"resyId,omitempty"`
	ResyURL                     string                  `json:"resyUrl,omitempty"`
	ResyActive                  bool                    `json:"resyActive"`
	OpenTableID                 string                  `json:"openTableId,omitempty"`
	OpenTableURL                string                  `json:"openTableUrl,omitempty"`
	OpenTableActive             bool                    `json:"openTableActive"`
	SevenRoomsID                string                  `json:"sevenRoomsId,omitempty"`
	SevenRoomsSlug              string                  `json:"sevenRoomsSlug,omitempty"`
	SevenRoomsURL               string                  `json:"sevenRoomsUrl,omitempty"`
	SevenRoomsActive            bool                    `json:"sevenRoomsActive"`
	DoorDashID                  string                  `json:"doorDashId,omitempty"`
	DoorDashURL                 string                  `json:"doorDashUrl,omitempty"`
	DoorDashActive              bool                    `json:"doorDashActive"`
	MinimumPartySize            int                     `json:"minimumPartySize,omitempty"`
	MaximumPartySize            int                     `json:"maximumPartySize,omitempty"`
	BookingDifficulty           int                     `json:"bookingDifficulty,omitempty"`
	BookingDifficultyDetails    string                  `json:"bookingDifficultyDetails,omitempty"`
	DepositFeeAmount            string                  `json:"depositFeeAmount,omitempty"`
	DepositFeeCutoffTime        string                  `json:"depositFeeCutoffTime,omitempty"`
	DepositFeeCutoffWindow      int                     `json:"depositFeeCutoffWindow,omitempty"`
	CancellationFeeAmount       string                  `json:"cancellationFeeAmount,omitempty"`
	CancellationFeeCutoffTime   string                  `json:"cancellationFeeCutoffTime,omitempty"`
	CancellationFeeCutoffWindow int                     `json:"cancellationFeeCutoffWindow,omitempty"`
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

type RestaurantSearchScope string

const (
	RestaurantSearchScopeAll       RestaurantSearchScope = "all"
	RestaurantSearchScopeCanonical RestaurantSearchScope = "canonical"
)

type RestaurantSearchSort string

const (
	RestaurantSearchSortRelevance RestaurantSearchSort = "relevance"
	RestaurantSearchSortDistance  RestaurantSearchSort = "distance"
	RestaurantSearchSortRecent    RestaurantSearchSort = "recent"
)

// RestaurantSearchParams are inputs for RestaurantsService.Discover.
//
// String fields are omitted when blank after trimming. Numeric fields are
// omitted when zero, except Lat and Lon which use pointers so zero is a valid
// coordinate.
type RestaurantSearchParams struct {
	Q            string
	Cuisine      string
	Neighborhood string
	Market       string
	Scope        RestaurantSearchScope
	Sort         RestaurantSearchSort
	Limit        int
	Lat          *float64
	Lon          *float64
	RadiusMeters int
}

type RestaurantCoordinates struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type MealService string

const (
	MealServiceBreakfast MealService = "breakfast"
	MealServiceBrunch    MealService = "brunch"
	MealServiceLunch     MealService = "lunch"
	MealServiceDinner    MealService = "dinner"
)

type RestaurantAvailability struct {
	All   []MealService               `json:"all"`
	Daily RestaurantDailyAvailability `json:"daily"`
}

type RestaurantDailyAvailability struct {
	Monday    []MealService `json:"monday"`
	Tuesday   []MealService `json:"tuesday"`
	Wednesday []MealService `json:"wednesday"`
	Thursday  []MealService `json:"thursday"`
	Friday    []MealService `json:"friday"`
	Saturday  []MealService `json:"saturday"`
	Sunday    []MealService `json:"sunday"`
}

type RestaurantRating struct {
	Source      string  `json:"source"`
	ArticleURL  string  `json:"articleUrl"`
	ArticleDate *string `json:"articleDate,omitempty"`
	Label       *string `json:"label,omitempty"`
	Score       float64 `json:"score"`
	Max         float64 `json:"max"`
}

type RestaurantMention struct {
	Source         string  `json:"source"`
	ArticleURL     string  `json:"articleUrl"`
	ArticleType    string  `json:"articleType"`
	ArticleTitle   string  `json:"articleTitle"`
	ArticleSummary *string `json:"articleSummary,omitempty"`
	ArticleDate    *string `json:"articleDate,omitempty"`
}

type RestaurantSearchResult struct {
	Restaurant     Restaurant `json:"restaurant"`
	Score          float64    `json:"score"`
	DistanceMeters *float64   `json:"distanceMeters,omitempty"`
}

func (r *RestaurantSearchResult) UnmarshalJSON(data []byte) error {
	type restaurantSearchResult RestaurantSearchResult
	var aux struct {
		restaurantSearchResult
		LegacyDistanceMeters *float64 `json:"distance_meters"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	*r = RestaurantSearchResult(aux.restaurantSearchResult)
	if r.DistanceMeters == nil {
		r.DistanceMeters = aux.LegacyDistanceMeters
	}
	return nil
}

// RestaurantSearchResponse is returned by the restaurant discovery endpoint.
type RestaurantSearchResponse struct {
	Results    []RestaurantSearchResult `json:"results"`
	NextCursor *string                  `json:"nextCursor"`
}

func (r *RestaurantSearchResponse) UnmarshalJSON(data []byte) error {
	var aux struct {
		Results          []RestaurantSearchResult `json:"results"`
		NextCursor       *string                  `json:"nextCursor"`
		LegacyNextCursor *string                  `json:"next_cursor"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	r.Results = aux.Results
	r.NextCursor = aux.NextCursor
	if r.NextCursor == nil {
		r.NextCursor = aux.LegacyNextCursor
	}
	return nil
}

type restaurantsSearchRequest struct {
	Query string `json:"query"`
}

type restaurantsSearchResponse struct {
	Restaurants []Restaurant `json:"restaurants"`
}

// Search returns restaurants matching the query string. It is a convenience
// helper for the legacy restaurant search response used by booking flows.
func (s *RestaurantsService) Search(ctx context.Context, query string) ([]Restaurant, error) {
	var resp restaurantsSearchResponse
	body := restaurantsSearchRequest{Query: query}
	if err := s.client.do(ctx, http.MethodPost, "/restaurants/search", body, &resp); err != nil {
		return nil, err
	}
	return resp.Restaurants, nil
}

// Discover searches the restaurant catalog using the GET /restaurants/search
// discovery endpoint.
func (s *RestaurantsService) Discover(ctx context.Context, params *RestaurantSearchParams) (*RestaurantSearchResponse, error) {
	path := "/restaurants/search"
	if params != nil {
		if query := params.queryString(); query != "" {
			path += "?" + query
		}
	}

	var resp RestaurantSearchResponse
	if err := s.client.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (p *RestaurantSearchParams) queryString() string {
	values := url.Values{}
	addTrimmed := func(key, value string) {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			values.Set(key, trimmed)
		}
	}

	addTrimmed("q", p.Q)
	addTrimmed("cuisine", p.Cuisine)
	addTrimmed("neighborhood", p.Neighborhood)
	addTrimmed("market", p.Market)
	addTrimmed("scope", string(p.Scope))
	addTrimmed("sort", string(p.Sort))
	if p.Limit > 0 {
		values.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Lat != nil {
		values.Set("lat", strconv.FormatFloat(*p.Lat, 'f', -1, 64))
	}
	if p.Lon != nil {
		values.Set("lon", strconv.FormatFloat(*p.Lon, 'f', -1, 64))
	}
	if p.RadiusMeters > 0 {
		values.Set("radiusMeters", strconv.Itoa(p.RadiusMeters))
	}

	return values.Encode()
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
