package spot

// Error is the concrete error type returned from every SDK method.
//
// Callers can branch using errors.Is against the exported sentinel values:
//
//	if errors.Is(err, spot.ErrRestaurantNotFound) { ... }
//
// Or use errors.As to access HTTPStatus and Details for richer handling:
//
//	var e *spot.Error
//	if errors.As(err, &e) { /* read e.HTTPStatus, e.Details */ }
type Error struct {
	// Code is a stable machine-readable identifier for the error.
	Code string
	// Message is the human-readable description.
	Message string
	// HTTPStatus is the HTTP status code returned by the API. Zero for non-HTTP errors.
	HTTPStatus int
	// Details is an optional server-provided detail blob.
	Details map[string]any
	// Platform is populated only for ErrPlatformNotConnected; names the specific
	// booking platform that the authenticated user has not linked (e.g. "resy").
	Platform string
}

// Error implements the error interface.
func (e *Error) Error() string {
	return e.Message
}

// Is reports whether target matches this error by Code. It enables
// errors.Is comparisons against the sentinel values.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok || t == nil {
		return false
	}
	return e.Code == t.Code
}

// Sentinel errors. Compare using errors.Is.
var (
	// ErrUnauthenticated means no credentials were supplied or the token source returned nothing.
	ErrUnauthenticated = &Error{Code: "unauthenticated", Message: "unauthenticated"}
	// ErrAuthExpired means the access token could not be refreshed and a new login is required.
	ErrAuthExpired = &Error{Code: "auth_expired", Message: "authentication expired"}
	// ErrRestaurantNotFound is returned when a restaurant lookup fails.
	ErrRestaurantNotFound = &Error{Code: "restaurant_not_found", Message: "restaurant not found"}
	// ErrSearchNotFound is returned when a search ID does not resolve.
	ErrSearchNotFound = &Error{Code: "search_not_found", Message: "search not found"}
	// ErrReservationNotFound is returned when a reservation ID does not resolve.
	ErrReservationNotFound = &Error{Code: "reservation_not_found", Message: "reservation not found"}
	// ErrNoAvailability means no matching availability was found.
	ErrNoAvailability = &Error{Code: "no_availability", Message: "no availability"}
	// ErrConflict means the request conflicts with existing state (e.g., already booked).
	ErrConflict = &Error{Code: "conflict", Message: "conflict with existing state"}
	// ErrValidation means the request was malformed or failed server-side validation.
	ErrValidation = &Error{Code: "validation", Message: "validation error"}
	// ErrRateLimited means the server returned HTTP 429 (Too Many Requests).
	ErrRateLimited = &Error{Code: "rate_limited", Message: "rate limited"}
	// ErrServer means the server returned a 5xx response.
	ErrServer = &Error{Code: "server_error", Message: "server error"}
	// ErrSlotExpired means a ReservationSlot previously returned by
	// ReservationsService.Search is no longer bookable — its short TTL passed,
	// someone else booked it, or the booking provider rejected the token.
	// Callers should re-run Search.
	ErrSlotExpired = &Error{Code: "slot_expired", Message: "slot is no longer available"}
	// ErrPlatformNotConnected means the authenticated user has not linked
	// their credentials for the booking platform this slot belongs to. The
	// error's Platform field names which one (e.g. "resy").
	ErrPlatformNotConnected = &Error{Code: "platform_not_connected", Message: "booking platform is not connected"}
)
