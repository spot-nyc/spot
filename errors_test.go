package spot

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Error(t *testing.T) {
	e := &Error{Code: "restaurant_not_found", Message: "restaurant not found: gramercy tavern"}
	assert.Equal(t, "restaurant not found: gramercy tavern", e.Error())
}

func TestError_Is_MatchesSentinelByCode(t *testing.T) {
	e := &Error{Code: "restaurant_not_found", Message: "restaurant not found"}
	assert.True(t, errors.Is(e, ErrRestaurantNotFound), "errors.Is should match sentinel by code")
}

func TestError_Is_DifferentCodeDoesNotMatch(t *testing.T) {
	e := &Error{Code: "restaurant_not_found"}
	assert.False(t, errors.Is(e, ErrSearchNotFound), "errors.Is should not match different code")
}

func TestError_Is_NonSpotErrorTargetDoesNotMatch(t *testing.T) {
	e := &Error{Code: "x"}
	assert.False(t, errors.Is(e, errors.New("some other error")))
}

func TestError_Is_NilSpotErrorTargetDoesNotMatch(t *testing.T) {
	e := &Error{Code: "x"}
	var nilTarget *Error
	assert.False(t, errors.Is(e, nilTarget), "errors.Is should not match a nil *Error target")
}

func TestSentinels_AllHaveCodes(t *testing.T) {
	sentinels := []*Error{
		ErrUnauthenticated,
		ErrAuthExpired,
		ErrRestaurantNotFound,
		ErrSearchNotFound,
		ErrReservationNotFound,
		ErrNoAvailability,
		ErrConflict,
		ErrValidation,
		ErrRateLimited,
		ErrServer,
	}
	seen := make(map[string]bool)
	for _, s := range sentinels {
		assert.NotEmpty(t, s.Code, "sentinel code must be non-empty")
		assert.False(t, seen[s.Code], "sentinel code %q is duplicated", s.Code)
		seen[s.Code] = true
	}
}
