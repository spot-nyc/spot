package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/spot-nyc/spot/auth"
	"github.com/spot-nyc/spot/internal/render"
)

func newAuthCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Sign in to Spot, manage and inspect your session",
	}

	cmd.AddCommand(newAuthLoginCmd(flags))
	cmd.AddCommand(newAuthLogoutCmd(flags))
	cmd.AddCommand(newAuthWhoamiCmd(flags))

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

			client, err := newClient()
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

			format := flags.resolveFormat(cmd.OutOrStdout())
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

func newAuthLogoutCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Sign this session out of Spot",
		Long: "Removes locally stored credentials. Your other devices (mobile,\n" +
			"web) stay signed in. To revoke access everywhere, visit your Spot\n" +
			"account settings.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := auth.DefaultStore().Delete(); err != nil {
				return err
			}

			format := flags.resolveFormat(cmd.OutOrStdout())
			if format == render.FormatJSON {
				return render.JSON(cmd.OutOrStdout(), map[string]any{"signedOut": true})
			}
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "Signed out.\n")
			return err
		},
	}
}

func newAuthWhoamiCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show the currently-authenticated user",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := newClient()
			if err != nil {
				return err
			}

			user, err := client.Users.Me(cmd.Context())
			if err != nil {
				return err
			}

			format := flags.resolveFormat(cmd.OutOrStdout())
			if format == render.FormatJSON {
				return render.JSON(cmd.OutOrStdout(), user)
			}

			tw := render.Table(cmd.OutOrStdout())
			_, _ = fmt.Fprintf(tw, "ID\t%s\n", user.ID)
			if user.Name != "" {
				_, _ = fmt.Fprintf(tw, "Name\t%s\n", user.Name)
			}
			if user.Phone != "" {
				_, _ = fmt.Fprintf(tw, "Phone\t%s\n", user.Phone)
			}
			return tw.Flush()
		},
	}
}
