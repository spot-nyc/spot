package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/spot-nyc/spot"
	"github.com/spot-nyc/spot/internal/render"
)

func newReservationsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reservations",
		Short: "Manage booked reservations",
	}

	cmd.AddCommand(newReservationsListCmd(flags))
	cmd.AddCommand(newReservationsCancelCmd(flags))
	cmd.AddCommand(newReservationsSearchCmd(flags))
	cmd.AddCommand(newReservationsBookCmd(flags))

	return cmd
}

func newReservationsListCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List your upcoming reservations",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			reservations, err := client.Reservations.List(cmd.Context())
			if err != nil {
				return err
			}

			format := flags.resolveFormat(cmd.OutOrStdout())
			if format == render.FormatJSON {
				return render.JSON(cmd.OutOrStdout(), reservations)
			}

			if len(reservations) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No upcoming reservations.")
				return nil
			}

			tw := render.Table(cmd.OutOrStdout())
			_, _ = fmt.Fprintln(tw, "ID\tRESTAURANT\tDATE\tTIME\tPARTY\tSEATING")
			for _, r := range reservations {
				restaurantName := "—"
				if r.Table.Restaurant != nil && r.Table.Restaurant.Name != "" {
					restaurantName = r.Table.Restaurant.Name
				}
				seating := formatSeating(r.Table.Seating)
				if seating == "" {
					seating = "—"
				}
				_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\t%s\n",
					r.ID,
					restaurantName,
					formatDate(r.Table.Date),
					formatTime(r.Table.Time),
					r.Table.Party,
					seating,
				)
			}
			return tw.Flush()
		},
	}
}

func newReservationsCancelCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "cancel <id>",
		Short: "Cancel an upcoming reservation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			client, err := newClient()
			if err != nil {
				return err
			}

			if err := client.Reservations.Cancel(cmd.Context(), id); err != nil {
				return err
			}

			format := flags.resolveFormat(cmd.OutOrStdout())
			if format == render.FormatJSON {
				return render.JSON(cmd.OutOrStdout(), map[string]any{
					"cancelled": true,
					"id":        id,
				})
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Reservation %s cancelled.\n", id)
			return err
		},
	}
}

func newReservationsSearchCmd(flags *rootFlags) *cobra.Command {
	var (
		restaurants []string
		date        string
		startTime   string
		endTime     string
		party       int
	)

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Find reservations available to book right now",
		Long: "Searches for available tables at the specified restaurants within\n" +
			"the given date and time window. Each result includes a slot ID you\n" +
			"can pass to 'spot reservations book <id>'. Slots have a short TTL —\n" +
			"re-run search if the result is stale.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			params := &spot.SearchReservationsParams{
				RestaurantIDs: restaurants,
				Date:          date,
				StartTime:     expandTime(startTime),
				EndTime:       expandTime(endTime),
				Party:         party,
			}

			slots, err := client.Reservations.Search(cmd.Context(), params)
			if err != nil {
				return err
			}

			format := flags.resolveFormat(cmd.OutOrStdout())
			if format == render.FormatJSON {
				return render.JSON(cmd.OutOrStdout(), slots)
			}

			if len(slots) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No tables available in this window. Try a wider time range or other restaurants.")
				return nil
			}

			tw := render.Table(cmd.OutOrStdout())
			_, _ = fmt.Fprintln(tw, "ID\tRESTAURANT\tDATE\tTIME\tSEATING\tPLATFORM\tPARTY")
			for _, slot := range slots {
				restaurantName := "—"
				if slot.Restaurant != nil && slot.Restaurant.Name != "" {
					restaurantName = slot.Restaurant.Name
				}
				seating := formatSeating(slot.Seating)
				if seating == "" {
					seating = "—"
				}
				_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%d\n",
					slot.ID,
					restaurantName,
					formatDate(slot.Date),
					formatTime(slot.Time),
					seating,
					platformDisplayName(slot.Platform),
					slot.Party,
				)
			}
			return tw.Flush()
		},
	}

	cmd.Flags().StringSliceVar(&restaurants, "restaurant", nil, "restaurant ID (repeatable or comma-separated; at least one required)")
	cmd.Flags().StringVar(&date, "date", "", "search date YYYY-MM-DD (required)")
	cmd.Flags().StringVar(&startTime, "start-time", "", "earliest time HH:MM or HH:MM:SS (required)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "latest time HH:MM or HH:MM:SS (required)")
	cmd.Flags().IntVar(&party, "party", 0, "party size (required)")
	_ = cmd.MarkFlagRequired("restaurant")
	_ = cmd.MarkFlagRequired("date")
	_ = cmd.MarkFlagRequired("start-time")
	_ = cmd.MarkFlagRequired("end-time")
	_ = cmd.MarkFlagRequired("party")

	return cmd
}

func newReservationsBookCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "book <slotId>",
		Short: "Book a reservation slot by slot ID",
		Long: "Books a specific slot returned by 'spot reservations search'.\n" +
			"Slots expire quickly — if you see a 'slot no longer available'\n" +
			"error, re-run search and try again with the new ID.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			reservation, err := client.Reservations.Book(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			format := flags.resolveFormat(cmd.OutOrStdout())
			if format == render.FormatJSON {
				return render.JSON(cmd.OutOrStdout(), reservation)
			}

			restaurantName := "the restaurant"
			if reservation.Table.Restaurant != nil && reservation.Table.Restaurant.Name != "" {
				restaurantName = reservation.Table.Restaurant.Name
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(),
				"Booked %s at %s on %s at %s for %d.\n",
				reservation.ID,
				restaurantName,
				formatDate(reservation.Table.Date),
				formatTime(reservation.Table.Time),
				reservation.Table.Party,
			)
			return err
		},
	}
}
