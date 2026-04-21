package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spot-nyc/spot"
	"github.com/spot-nyc/spot/internal/render"
)

func TestExitCodeFor_Sentinels(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, 0},
		{"unauthenticated", spot.ErrUnauthenticated, 3},
		{"auth expired", spot.ErrAuthExpired, 4},
		{"restaurant not found", spot.ErrRestaurantNotFound, 5},
		{"search not found", spot.ErrSearchNotFound, 5},
		{"reservation not found", spot.ErrReservationNotFound, 5},
		{"no availability", spot.ErrNoAvailability, 5},
		{"conflict", spot.ErrConflict, 6},
		{"validation", spot.ErrValidation, 7},
		{"rate limited", spot.ErrRateLimited, 8},
		{"server", spot.ErrServer, 9},
		{"platform not connected", spot.ErrPlatformNotConnected, 10},
		{"slot expired", spot.ErrSlotExpired, 11},
		{"generic not-found code", &spot.Error{Code: "not_found", HTTPStatus: 404}, 5},
		{"unknown spot error", &spot.Error{Code: "something_weird"}, 1},
		{"plain error", errors.New("boom"), 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, ExitCodeFor(tc.err))
		})
	}
}

func TestRenderError_JSON(t *testing.T) {
	var buf bytes.Buffer
	RenderError(&buf, render.FormatJSON, spot.ErrUnauthenticated)

	var envelope struct {
		Error struct {
			Code       string `json:"code"`
			Message    string `json:"message"`
			HTTPStatus int    `json:"httpStatus,omitempty"`
		} `json:"error"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &envelope))
	assert.Equal(t, "unauthenticated", envelope.Error.Code)
	assert.Equal(t, "unauthenticated", envelope.Error.Message)
}

func TestRenderError_JSON_PlainError(t *testing.T) {
	var buf bytes.Buffer
	RenderError(&buf, render.FormatJSON, errors.New("boom"))

	var envelope struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &envelope))
	assert.Equal(t, "error", envelope.Error.Code)
	assert.Equal(t, "boom", envelope.Error.Message)
}

func TestRenderError_Table_FriendlyUnauthenticated(t *testing.T) {
	var buf bytes.Buffer
	RenderError(&buf, render.FormatTable, spot.ErrUnauthenticated)

	got := buf.String()
	assert.Contains(t, got, "not signed in", "table mode should hint at sign-in")
	assert.Contains(t, got, "spot auth login", "table mode should point at the login command")
	assert.NotContains(t, got, "{", "table mode should not emit JSON")
}

func TestRenderError_Table_FriendlyAuthExpired(t *testing.T) {
	var buf bytes.Buffer
	RenderError(&buf, render.FormatTable, spot.ErrAuthExpired)

	got := buf.String()
	assert.Contains(t, got, "session expired")
	assert.Contains(t, got, "spot auth login")
}

func TestRenderError_Table_GenericErrorFallback(t *testing.T) {
	var buf bytes.Buffer
	RenderError(&buf, render.FormatTable, errors.New("boom"))

	got := buf.String()
	assert.Contains(t, got, "error:")
	assert.Contains(t, got, "boom")
}

func TestRenderError_Table_FriendlyPlatformNotConnected(t *testing.T) {
	err := &spot.Error{
		Code:       spot.ErrPlatformNotConnected.Code,
		Message:    "platform not connected",
		HTTPStatus: 412,
		Platform:   "resy",
	}

	var buf bytes.Buffer
	RenderError(&buf, render.FormatTable, err)

	got := buf.String()
	assert.Contains(t, got, "Resy account isn't connected")
	assert.Contains(t, got, "Spot mobile app")
}

func TestRenderError_Table_FriendlySlotExpired(t *testing.T) {
	var buf bytes.Buffer
	RenderError(&buf, render.FormatTable, spot.ErrSlotExpired)

	got := buf.String()
	assert.Contains(t, got, "no longer available")
	assert.Contains(t, got, "spot reservations search")
}

func TestPlatformDisplayName(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"resy", "Resy"},
		{"opentable", "OpenTable"},
		{"sevenrooms", "SevenRooms"},
		{"doordash", "DoorDash"},
		{"unknown-platform", "unknown-platform"},
		{"", ""},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, tc.want, platformDisplayName(tc.in))
		})
	}
}
