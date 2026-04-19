package main

import "github.com/spf13/cobra"

func newAuthCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "auth",
		Short: "Manage Spot authentication",
	}
}
