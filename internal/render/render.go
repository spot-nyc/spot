// Package render provides CLI output helpers: format resolution (JSON vs
// table), pretty JSON encoding, and aligned table writing. Used only by
// cmd/spot; the library (spot package) never imports this.
package render

import (
	"encoding/json"
	"io"
	"os"
	"text/tabwriter"

	"github.com/spot-nyc/spot/internal/tty"
)

// Format selects the output representation.
type Format int

const (
	// FormatJSON emits JSON to the writer.
	FormatJSON Format = iota
	// FormatTable emits aligned, human-readable text.
	FormatTable
)

// Resolve decides the effective format for an *os.File destination:
//
//  1. If forceJSON is true, return FormatJSON.
//  2. Else if out is a TTY, return FormatTable.
//  3. Else return FormatJSON (piped / redirected).
func Resolve(forceJSON bool, out *os.File) Format {
	if forceJSON {
		return FormatJSON
	}
	if out != nil && tty.IsTerminal(out.Fd()) {
		return FormatTable
	}
	return FormatJSON
}

// ResolveWriter is the io.Writer variant. Non-file writers can't be probed
// for TTY-ness, so it falls back to JSON unless forceJSON is also false
// (in which case it returns JSON too — there is no TTY path for arbitrary
// writers). Useful for tests writing to bytes.Buffer.
func ResolveWriter(forceJSON bool, out io.Writer) Format {
	if f, ok := out.(*os.File); ok {
		return Resolve(forceJSON, f)
	}
	return FormatJSON
}

// JSON writes v as indented JSON (2-space) to w, terminated by a newline.
func JSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// Table returns a tabwriter configured for Spot's default CLI style:
// two-space inter-column gap, left-aligned, space-padded.
//
// Callers write tab-separated rows via tw.Write, then call tw.Flush().
func Table(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
}
