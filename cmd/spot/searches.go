package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/spot-nyc/spot"
	"github.com/spot-nyc/spot/auth"
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
			client, err := spot.NewClient(spot.WithTokenSource(auth.DefaultTokenSource()))
			if err != nil {
				return err
			}

			searches, err := client.Searches.List(cmd.Context())
			if err != nil {
				return err
			}

			format := flags.resolveFormat()
			if format == render.FormatJSON {
				return render.JSON(cmd.OutOrStdout(), searches)
			}

			if len(searches) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No active searches.")
				return nil
			}

			tw := render.Table(cmd.OutOrStdout())
			_, _ = fmt.Fprintln(tw, "ID\tPARTY\tDATES\tTIMES\tTARGETS")
			for _, s := range searches {
				dates := s.StartDate
				if s.EndDate != s.StartDate {
					dates = s.StartDate + " → " + s.EndDate
				}
				times := trimSeconds(s.StartTime) + "-" + trimSeconds(s.EndTime)
				_, _ = fmt.Fprintf(tw, "%s\t%d\t%s\t%s\t%d\n",
					s.ID, s.Party, dates, times, len(s.SearchTargets))
			}
			return tw.Flush()
		},
	}
}

// trimSeconds reduces "HH:MM:SS" → "HH:MM" for readable table output.
// Accepts any value it doesn't understand and returns it unchanged.
func trimSeconds(t string) string {
	if len(t) >= 5 && t[2] == ':' {
		return t[:5]
	}
	return t
}
