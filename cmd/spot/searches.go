package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/spot-nyc/spot"
	"github.com/spot-nyc/spot/internal/render"
)

func newSearchesCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "searches",
		Short: "Manage reservation searches",
	}

	cmd.AddCommand(newSearchesListCmd(flags))

	return cmd
}

func newSearchesListCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List your active searches",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			searches, err := client.Searches.List(cmd.Context())
			if err != nil {
				return err
			}

			format := flags.resolveFormat(cmd.OutOrStdout())
			if format == render.FormatJSON {
				return render.JSON(cmd.OutOrStdout(), searches)
			}

			if len(searches) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No active searches.")
				return nil
			}

			tw := render.Table(cmd.OutOrStdout())
			_, _ = fmt.Fprintln(tw, "ID\tPARTY\tDATE\tTIME\tRESTAURANTS")
			for _, s := range searches {
				timeRange := formatTime(s.StartTime) + "–" + formatTime(s.EndTime)
				_, _ = fmt.Fprintf(tw, "%s\t%d\t%s\t%s\t%s\n",
					shortID(s.ID), s.Party, formatDate(s.StartDate), timeRange, joinRestaurantNames(s.SearchTargets))
			}
			return tw.Flush()
		},
	}
}

// shortID returns a display-friendly abbreviation of an opaque ID. We show
// the first 8 characters and append an ellipsis when the original is longer.
// Raw IDs remain available via --json or `spot searches get <id>`.
const shortIDLen = 8

func shortID(id string) string {
	if len(id) <= shortIDLen {
		return id
	}
	return id[:shortIDLen] + "…"
}

// formatTime renders an "HH:MM:SS" or "HH:MM" string in 12-hour clock form
// with an "AM"/"PM" suffix (e.g. "18:00:00" → "6:00 PM"). Unrecognized inputs
// pass through unchanged so tables never render as empty.
func formatTime(t string) string {
	for _, layout := range []string{"15:04:05", "15:04"} {
		if parsed, err := time.Parse(layout, t); err == nil {
			return parsed.Format("3:04 PM")
		}
	}
	return t
}

// formatDate renders a "YYYY-MM-DD" string as "Jan 2, 2006" (e.g.
// "2026-05-01" → "May 1, 2026"). Unrecognized inputs pass through unchanged.
func formatDate(d string) string {
	if parsed, err := time.Parse("2006-01-02", d); err == nil {
		return parsed.Format("Jan 2, 2006")
	}
	return d
}

// joinRestaurantNames flattens a search's targets into a comma-separated
// string of restaurant names. Targets without a populated Restaurant (possible
// if the API ever omits the join) are skipped silently.
func joinRestaurantNames(targets []spot.SearchTarget) string {
	if len(targets) == 0 {
		return "—"
	}
	names := make([]string, 0, len(targets))
	for _, t := range targets {
		if t.Restaurant == nil || t.Restaurant.Name == "" {
			continue
		}
		names = append(names, t.Restaurant.Name)
	}
	if len(names) == 0 {
		return "—"
	}
	return strings.Join(names, ", ")
}
