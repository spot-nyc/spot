package main

import (
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

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
	cmd.AddCommand(newSearchesGetCmd(flags))
	cmd.AddCommand(newSearchesDeleteCmd(flags))
	cmd.AddCommand(newSearchesCreateCmd(flags))
	cmd.AddCommand(newSearchesUpdateCmd(flags))

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
					s.ID, s.Party, formatDate(s.StartDate), timeRange, joinRestaurantNames(s.SearchTargets))
			}
			return tw.Flush()
		},
	}
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

// formatSeating normalizes the raw seating string the Spot API returns for display.
// OpenTable stores its default dining-area slots as "default"; we surface
// that as "Dining Room" to match the mobile client. Other values are
// title-cased (e.g. "bar" → "Bar"). Empty input returns empty so callers
// can apply their own placeholder.
func formatSeating(s string) string {
	if s == "" {
		return ""
	}
	if s == "default" {
		return "Dining Room"
	}
	return titlecase(s)
}

// titlecase uppercases the first rune of each space-separated word, leaving
// the rest untouched. Mirrors the mobile client's `titlecase` so seating
// types render identically on CLI and mobile.
func titlecase(s string) string {
	words := strings.Split(s, " ")
	for i, word := range words {
		if word == "" {
			continue
		}
		r, size := utf8.DecodeRuneInString(word)
		words[i] = string(unicode.ToUpper(r)) + word[size:]
	}
	return strings.Join(words, " ")
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

func newSearchesGetCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Show details for one of your searches",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			search, err := client.Searches.Get(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			format := flags.resolveFormat(cmd.OutOrStdout())
			if format == render.FormatJSON {
				return render.JSON(cmd.OutOrStdout(), search)
			}

			tw := render.Table(cmd.OutOrStdout())
			_, _ = fmt.Fprintf(tw, "ID\t%s\n", search.ID)
			_, _ = fmt.Fprintf(tw, "Party\t%d\n", search.Party)
			_, _ = fmt.Fprintf(tw, "Date\t%s\n", formatDate(search.StartDate))
			_, _ = fmt.Fprintf(tw, "Time\t%s – %s\n", formatTime(search.StartTime), formatTime(search.EndTime))
			_, _ = fmt.Fprintf(tw, "Restaurants\t%s\n", joinRestaurantNames(search.SearchTargets))
			return tw.Flush()
		},
	}
}

func newSearchesDeleteCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete one of your searches",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			client, err := newClient()
			if err != nil {
				return err
			}

			if err := client.Searches.Delete(cmd.Context(), id); err != nil {
				return err
			}

			format := flags.resolveFormat(cmd.OutOrStdout())
			if format == render.FormatJSON {
				return render.JSON(cmd.OutOrStdout(), map[string]any{
					"deleted": true,
					"id":      id,
				})
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Search %s deleted.\n", id)
			return err
		},
	}
}

func newSearchesCreateCmd(flags *rootFlags) *cobra.Command {
	var (
		party       int
		date        string
		startTime   string
		endTime     string
		restaurants []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new reservation search",
		Long: "Creates a new reservation search with the given party size, date, time\n" +
			"window, and restaurant IDs. Use 'spot restaurants search <query>' to find\n" +
			"restaurant IDs.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			params := &spot.CreateSearchParams{
				Party:         party,
				StartDate:     date,
				EndDate:       date,
				StartTime:     expandTime(startTime),
				EndTime:       expandTime(endTime),
				RestaurantIDs: restaurants,
			}

			created, err := client.Searches.Create(cmd.Context(), params)
			if err != nil {
				return err
			}

			format := flags.resolveFormat(cmd.OutOrStdout())
			if format == render.FormatJSON {
				return render.JSON(cmd.OutOrStdout(), created)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Search created: %s\n", created.ID)
			return err
		},
	}

	cmd.Flags().IntVar(&party, "party", 0, "party size (required)")
	cmd.Flags().StringVar(&date, "date", "", "search date YYYY-MM-DD (required)")
	cmd.Flags().StringVar(&startTime, "start-time", "", "earliest time HH:MM or HH:MM:SS (required)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "latest time HH:MM or HH:MM:SS (required)")
	cmd.Flags().StringSliceVar(&restaurants, "restaurant", nil, "restaurant ID (repeatable; at least one required)")
	_ = cmd.MarkFlagRequired("party")
	_ = cmd.MarkFlagRequired("date")
	_ = cmd.MarkFlagRequired("start-time")
	_ = cmd.MarkFlagRequired("end-time")
	_ = cmd.MarkFlagRequired("restaurant")

	return cmd
}

// expandTime accepts "HH:MM" and normalizes to "HH:MM:SS" by appending ":00".
// Strings of any other length (including already-normalized "HH:MM:SS") pass
// through unchanged; the server validates the final format.
func expandTime(t string) string {
	if len(t) == 5 && t[2] == ':' {
		return t + ":00"
	}
	return t
}

func newSearchesUpdateCmd(flags *rootFlags) *cobra.Command {
	var (
		party       int
		date        string
		startTime   string
		endTime     string
		restaurants []string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Modify an existing search",
		Long: "Updates fields on an existing search. Only the flags you set are\n" +
			"sent — unset flags leave their server-side values alone. --restaurant\n" +
			"is a full replacement of the search's restaurant list.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			flagSet := cmd.Flags()
			params := &spot.UpdateSearchParams{}
			changed := false
			if flagSet.Changed("party") {
				p := party
				params.Party = &p
				changed = true
			}
			if flagSet.Changed("date") {
				d := date
				params.StartDate = &d
				params.EndDate = &d
				changed = true
			}
			if flagSet.Changed("start-time") {
				t := expandTime(startTime)
				params.StartTime = &t
				changed = true
			}
			if flagSet.Changed("end-time") {
				t := expandTime(endTime)
				params.EndTime = &t
				changed = true
			}
			if flagSet.Changed("restaurant") {
				params.RestaurantIDs = restaurants
				changed = true
			}

			if !changed {
				return fmt.Errorf("nothing to update; provide at least one flag")
			}

			updated, err := client.Searches.Update(cmd.Context(), args[0], params)
			if err != nil {
				return err
			}

			format := flags.resolveFormat(cmd.OutOrStdout())
			if format == render.FormatJSON {
				return render.JSON(cmd.OutOrStdout(), updated)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Search %s updated.\n", updated.ID)
			return err
		},
	}

	cmd.Flags().IntVar(&party, "party", 0, "party size")
	cmd.Flags().StringVar(&date, "date", "", "search date YYYY-MM-DD (sets both start and end)")
	cmd.Flags().StringVar(&startTime, "start-time", "", "earliest time HH:MM or HH:MM:SS")
	cmd.Flags().StringVar(&endTime, "end-time", "", "latest time HH:MM or HH:MM:SS")
	cmd.Flags().StringSliceVar(&restaurants, "restaurant", nil, "restaurant ID (repeatable or comma-separated; full replacement)")

	return cmd
}
