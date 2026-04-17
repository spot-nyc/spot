// Command spot is the Spot CLI — manage reservations, searches, and restaurant lookup.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/spot-nyc/spot"
)

func newRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "spot",
		Short: "Spot — manage reservations, searches, and restaurant lookup",
		Long: "The Spot CLI.\n\n" +
			"Monitor reservations, create searches, book tables, and query restaurants " +
			"from the command line.\n\n" +
			"Run `spot <command> --help` for detail on any command.",
		Version:       spot.Version,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(command *cobra.Command, _ []string) error {
			return command.Help()
		},
	}
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
