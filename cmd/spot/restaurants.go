package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/spot-nyc/spot"
	"github.com/spot-nyc/spot/internal/render"
)

func newRestaurantsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restaurants",
		Short: "Look up restaurants",
	}

	cmd.AddCommand(newRestaurantsSearchCmd(flags))
	cmd.AddCommand(newRestaurantsDiscoverCmd(flags))
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

func newRestaurantsDiscoverCmd(flags *rootFlags) *cobra.Command {
	var (
		query        string
		cuisine      string
		neighborhood string
		market       string
		scope        string
		sort         string
		limit        int
		lat          float64
		lon          float64
		radiusMeters int
	)

	cmd := &cobra.Command{
		Use:   "discover [query]",
		Short: "Discover restaurants with filters",
		Long: "Searches the restaurant discovery endpoint by text, cuisine,\n" +
			"neighborhood, market, or geo filters.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				if strings.TrimSpace(query) != "" {
					return fmt.Errorf("provide search text either as [query] or --q, not both")
				}
				query = args[0]
			}
			if err := validateRestaurantSearchFlags(cmd, scope, sort, limit, lat, lon, radiusMeters); err != nil {
				return err
			}

			client, err := newClient()
			if err != nil {
				return err
			}

			params := &spot.RestaurantSearchParams{
				Q:            query,
				Cuisine:      cuisine,
				Neighborhood: neighborhood,
				Market:       market,
				Scope:        spot.RestaurantSearchScope(scope),
				Sort:         spot.RestaurantSearchSort(sort),
			}
			flagSet := cmd.Flags()
			if flagSet.Changed("limit") {
				params.Limit = limit
			}
			if flagSet.Changed("lat") {
				params.Lat = &lat
			}
			if flagSet.Changed("lon") {
				params.Lon = &lon
			}
			if flagSet.Changed("radius-meters") {
				params.RadiusMeters = radiusMeters
			}

			response, err := client.Restaurants.Discover(cmd.Context(), params)
			if err != nil {
				return err
			}

			format := flags.resolveFormat(cmd.OutOrStdout())
			if format == render.FormatJSON {
				return render.JSON(cmd.OutOrStdout(), response)
			}

			if len(response.Results) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No restaurants matched.")
				return nil
			}

			tw := render.Table(cmd.OutOrStdout())
			_, _ = fmt.Fprintln(tw, "ID\tNAME\tCUISINE\tNEIGHBORHOOD\tPLATFORMS\tSCORE")
			for _, result := range response.Results {
				r := result.Restaurant
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
				id := r.ID
				if id == "" {
					id = "—"
				}
				name := r.Name
				if r.Name == "" {
					name = "—"
				}
				_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%.3f\n",
					id, name, cuisine, neighborhood, platforms, result.Score)
			}
			if err := tw.Flush(); err != nil {
				return err
			}
			if response.NextCursor != nil && *response.NextCursor != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nNext cursor: %s\n", *response.NextCursor)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&query, "q", "", "text query")
	cmd.Flags().StringVar(&cuisine, "cuisine", "", "filter by cuisine")
	cmd.Flags().StringVar(&neighborhood, "neighborhood", "", "filter by neighborhood")
	cmd.Flags().StringVar(&market, "market", "", "filter by market (for NYC, use new-york)")
	cmd.Flags().StringVar(&scope, "scope", "", "scope: all, canonical")
	cmd.Flags().StringVar(&sort, "sort", "", "sort: relevance, distance, recent")
	cmd.Flags().IntVar(&limit, "limit", 0, "max results, 1-50 (server default 20)")
	cmd.Flags().Float64Var(&lat, "lat", 0, "latitude for geo search")
	cmd.Flags().Float64Var(&lon, "lon", 0, "longitude for geo search")
	cmd.Flags().IntVar(&radiusMeters, "radius-meters", 0, "positive search radius in meters")

	return cmd
}

func validateRestaurantSearchFlags(cmd *cobra.Command, scope, sort string, limit int, lat, lon float64, radiusMeters int) error {
	flagSet := cmd.Flags()
	latSet := flagSet.Changed("lat")
	lonSet := flagSet.Changed("lon")
	radiusSet := flagSet.Changed("radius-meters")
	if scope != "" && !oneOf(scope, "all", "canonical") {
		return fmt.Errorf("invalid --scope %q; expected all or canonical", scope)
	}
	if sort != "" && !oneOf(sort, "relevance", "distance", "recent") {
		return fmt.Errorf("invalid --sort %q; expected relevance, distance, or recent", sort)
	}
	if sort == "distance" && (!latSet || !lonSet) {
		return fmt.Errorf("--sort distance requires both --lat and --lon")
	}
	if flagSet.Changed("limit") && (limit < 1 || limit > 50) {
		return fmt.Errorf("invalid --limit %d; expected 1-50", limit)
	}
	if latSet != lonSet {
		return fmt.Errorf("--lat and --lon must be provided together")
	}
	if latSet && (lat < -90 || lat > 90) {
		return fmt.Errorf("invalid --lat %v; expected -90 to 90", lat)
	}
	if lonSet && (lon < -180 || lon > 180) {
		return fmt.Errorf("invalid --lon %v; expected -180 to 180", lon)
	}
	if radiusSet && (!latSet || !lonSet) {
		return fmt.Errorf("--radius-meters requires both --lat and --lon")
	}
	if radiusSet && radiusMeters <= 0 {
		return fmt.Errorf("invalid --radius-meters %d; expected a positive integer", radiusMeters)
	}
	return nil
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
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
			return tw.Flush()
		},
	}
}
