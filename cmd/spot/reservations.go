package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/spot-nyc/spot/internal/render"
)

func newReservationsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reservations",
		Short: "Manage booked reservations",
	}

	cmd.AddCommand(newReservationsListCmd(flags))
	cmd.AddCommand(newReservationsCancelCmd(flags))

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
