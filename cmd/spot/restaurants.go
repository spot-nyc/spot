package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/spot-nyc/spot/internal/render"
)

func newRestaurantsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restaurants",
		Short: "Look up restaurants",
	}

	cmd.AddCommand(newRestaurantsSearchCmd(flags))

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
			_, _ = fmt.Fprintln(tw, "ID\tNAME\tPLATFORM\tZONE")
			for _, r := range results {
				platform := r.Platform
				if platform == "" {
					platform = "—"
				}
				zone := r.Zone
				if zone == "" {
					zone = "—"
				}
				_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", r.ID, r.Name, platform, zone)
			}
			return tw.Flush()
		},
	}
}
