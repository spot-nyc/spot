package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/spot-nyc/spot"
	"github.com/spot-nyc/spot/internal/render"
)

// ExitCodeFor maps a library error to a stable CLI exit code. These codes
// are treated as API — scripts and LLM agents branch on them. See the
// design spec's "Exit codes" table.
func ExitCodeFor(err error) int {
	if err == nil {
		return 0
	}
	switch {
	case errors.Is(err, spot.ErrUnauthenticated):
		return 3
	case errors.Is(err, spot.ErrAuthExpired):
		return 4
	case errors.Is(err, spot.ErrRestaurantNotFound),
		errors.Is(err, spot.ErrSearchNotFound),
		errors.Is(err, spot.ErrReservationNotFound),
		errors.Is(err, spot.ErrNoAvailability):
		return 5
	case errors.Is(err, spot.ErrConflict):
		return 6
	case errors.Is(err, spot.ErrValidation):
		return 7
	case errors.Is(err, spot.ErrRateLimited):
		return 8
	case errors.Is(err, spot.ErrServer):
		return 9
	}

	// Generic *spot.Error not matching a sentinel: use its Code to pick.
	var spotErr *spot.Error
	if errors.As(err, &spotErr) {
		switch spotErr.Code {
		case "not_found":
			return 5
		}
	}

	return 1
}

// jsonErrorBody is the wire shape for --json error output.
type jsonErrorBody struct {
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	HTTPStatus int            `json:"httpStatus,omitempty"`
	Details    map[string]any `json:"details,omitempty"`
}

type jsonErrorEnvelope struct {
	Error jsonErrorBody `json:"error"`
}

// RenderError writes err in the requested format. Callers pair this with
// ExitCodeFor(err) and os.Exit.
func RenderError(w io.Writer, format render.Format, err error) {
	if err == nil {
		return
	}

	var body jsonErrorBody
	var spotErr *spot.Error
	if errors.As(err, &spotErr) {
		body = jsonErrorBody{
			Code:       spotErr.Code,
			Message:    spotErr.Message,
			HTTPStatus: spotErr.HTTPStatus,
			Details:    spotErr.Details,
		}
	} else {
		body = jsonErrorBody{Code: "error", Message: err.Error()}
	}

	if format == render.FormatJSON {
		_ = render.JSON(w, jsonErrorEnvelope{Error: body})
		return
	}

	// Table mode — friendlier messages for common cases, plain "error: <msg>"
	// as the fallback. JSON consumers get the structured envelope above;
	// humans get hand-holding here.
	switch {
	case errors.Is(err, spot.ErrUnauthenticated):
		_, _ = fmt.Fprintln(w, `You are not signed in. Run "spot auth login" to sign in.`)
	case errors.Is(err, spot.ErrAuthExpired):
		_, _ = fmt.Fprintln(w, `Your session expired. Run "spot auth login" to sign in again.`)
	default:
		_, _ = fmt.Fprintln(w, "error:", body.Message)
	}
}
