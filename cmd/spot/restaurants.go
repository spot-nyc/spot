package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/spot-nyc/spot/internal/render"
)

func newRestaurantsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restaurants",
		Short: "Look up restaurants",
	}

	cmd.AddCommand(newRestaurantsSearchCmd(flags))
	cmd.AddCommand(newRestaurantsGetCmd(flags))

	return cmd
}

func newRestaurantsSearchCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search restaurants by name",
		Long: "Returns restaurants matching the query string. Use the returned IDs\n" +
			"with 'spot searches create --restaurant <id>'.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			results, err := client.Restaurants.Search(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			format := flags.resolveFormat(cmd.OutOrStdout())
			if format == render.FormatJSON {
				return render.JSON(cmd.OutOrStdout(), results)
			}

			if len(results) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No restaurants matched.")
				return nil
			}

			tw := render.Table(cmd.OutOrStdout())
			_, _ = fmt.Fprintln(tw, "ID\tNAME\tCUISINE\tNEIGHBORHOOD\tPLATFORMS")
			for _, r := range results {
				cuisine := r.Cuisine
				if cuisine == "" {
					cuisine = "—"
				}
				neighborhood := r.Neighborhood
				if neighborhood == "" {
					neighborhood = "—"
				}
				platforms := strings.Join(r.Platforms(), ", ")
				if platforms == "" {
					platforms = "—"
				}
				_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", r.ID, r.Name, cuisine, neighborhood, platforms)
			}
			return tw.Flush()
		},
	}
}

func newRestaurantsGetCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Show restaurant details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			restaurant, err := client.Restaurants.Get(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			format := flags.resolveFormat(cmd.OutOrStdout())
			if format == render.FormatJSON {
				return render.JSON(cmd.OutOrStdout(), restaurant)
			}

			tw := render.Table(cmd.OutOrStdout())
			_, _ = fmt.Fprintf(tw, "ID\t%s\n", restaurant.ID)
			_, _ = fmt.Fprintf(tw, "Name\t%s\n", restaurant.Name)
			if restaurant.Cuisine != "" {
				_, _ = fmt.Fprintf(tw, "Cuisine\t%s\n", restaurant.Cuisine)
			}
			if restaurant.Neighborhood != "" {
				_, _ = fmt.Fprintf(tw, "Neighborhood\t%s\n", restaurant.Neighborhood)
			}
			if restaurant.Address != "" {
				_, _ = fmt.Fprintf(tw, "Address\t%s\n", restaurant.Address)
			}
			if restaurant.Phone != "" {
				_, _ = fmt.Fprintf(tw, "Phone\t%s\n", restaurant.Phone)
			}
			if restaurant.Website != "" {
				_, _ = fmt.Fprintf(tw, "Website\t%s\n", restaurant.Website)
			}
			platforms := strings.Join(restaurant.Platforms(), ", ")
			if platforms == "" {
				platforms = "—"
			}
			_, _ = fmt.Fprintf(tw, "Platforms\t%s\n", platforms)
			if restaurant.MinimumPartySize > 0 || restaurant.MaximumPartySize > 0 {
				_, _ = fmt.Fprintf(tw, "Party Limits\t%d-%d\n", restaurant.MinimumPartySize, restaurant.MaximumPartySize)
			}
			if restaurant.BookingDifficulty > 0 {
				_, _ = fmt.Fprintf(tw, "Booking Difficulty\t%d / 10\n", restaurant.BookingDifficulty)
			}
			return tw.Flush()
		},
	}
}
