package main

import "github.com/spf13/cobra"

func newSearchesCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "searches",
		Short: "Manage Spot searches",
	}
}
