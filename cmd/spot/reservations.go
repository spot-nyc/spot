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
			_, _ = fmt.Fprintln(tw, "ID\tRESTAURANT\tDATE\tTIME\tPARTY")
			for _, r := range reservations {
				restaurantName := "—"
				if r.Table.Restaurant != nil && r.Table.Restaurant.Name != "" {
					restaurantName = r.Table.Restaurant.Name
				}
				_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%d\n",
					shortID(r.ID),
					restaurantName,
					formatDate(r.Table.Date),
					formatTime(r.Table.Time),
					r.Table.Party,
				)
			}
			return tw.Flush()
		},
	}
}
