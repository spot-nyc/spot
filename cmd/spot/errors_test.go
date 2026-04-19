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

func TestRenderError_Table_PlainText(t *testing.T) {
	var buf bytes.Buffer
	RenderError(&buf, render.FormatTable, spot.ErrUnauthenticated)

	got := buf.String()
	assert.Contains(t, got, "unauthenticated")
	assert.NotContains(t, got, "{", "table mode should not emit JSON")
}
