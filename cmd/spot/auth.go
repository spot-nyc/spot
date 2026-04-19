package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/spot-nyc/spot"
	"github.com/spot-nyc/spot/auth"
	"github.com/spot-nyc/spot/internal/render"
)

func newAuthCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Sign in to Spot, manage and inspect your session",
	}

	cmd.AddCommand(newAuthLoginCmd(flags))

	return cmd
}

func newAuthLoginCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Sign in to Spot via your browser",
		Long: "Runs the PKCE authorization code flow against Supabase. Opens your\n" +
			"default browser for phone OTP sign-in and consent. On success, stores\n" +
			"credentials in the OS keychain (falling back to an XDG-path file).",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			fmt.Fprintln(os.Stderr, "Opening browser for sign-in…")

			creds, err := auth.DefaultLogin(ctx, auth.LoginOptions{})
			if err != nil {
				return err
			}

			if err := auth.DefaultStore().Save(creds); err != nil {
				return fmt.Errorf("save credentials: %w", err)
			}

			client, err := spot.NewClient(spot.WithTokenSource(auth.DefaultTokenSource()))
			if err != nil {
				return err
			}
			user, err := client.Users.Me(ctx)
			if err != nil {
				// Credentials saved, but profile lookup failed.
				// Still a partial success — the token works, we just can't greet.
				fmt.Fprintln(os.Stderr, "signed in (profile lookup failed:", err.Error()+")")
				return nil
			}

			format := flags.resolveFormat()
			if format == render.FormatJSON {
				return render.JSON(cmd.OutOrStdout(), map[string]any{
					"signedIn": true,
					"user":     user,
				})
			}

			name := user.Name
			if name == "" {
				name = user.Phone
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Signed in as %s\n", name)
			return err
		},
	}
}
