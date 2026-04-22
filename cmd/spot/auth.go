package main

import (
	"context"
	"fmt"
	"os"
	"strings"

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
	var all bool

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Clear local credentials and revoke the refresh token server-side",
		Long: "Revokes the refresh token on the Spot API and clears credentials\n" +
			"from the OS keyring / credential file. Use --all to revoke every\n" +
			"active session across all devices (useful if a token may have\n" +
			"leaked). Server-side revocation is best-effort — local credentials\n" +
			"are always cleared even if the server call fails.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			scope := ""
			if all {
				scope = "global"
			}

			// Best-effort server-side revocation. Failures are logged to
			// stderr but do not block local cleanup — a user who asks to
			// "log out" always gets their local session cleared.
			if client, clientErr := newClient(); clientErr == nil {
				if err := client.Users.Logout(cmd.Context(), scope); err != nil {
					_, _ = fmt.Fprintf(cmd.ErrOrStderr(),
						"warning: server-side revocation did not complete (%s). Your local credentials were cleared, but the refresh token may still be valid on the server. Run 'spot auth logout --all' from an authenticated session to revoke it.\n",
						err.Error())
				}
			}

			// Clear local credentials regardless of revocation outcome.
			if err := auth.DefaultStore().Delete(); err != nil {
				return err
			}

			format := flags.resolveFormat(cmd.OutOrStdout())
			if format == render.FormatJSON {
				return render.JSON(cmd.OutOrStdout(), map[string]any{"signedOut": true})
			}
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "Signed out.")
			return err
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "revoke the refresh token on all devices (global logout)")
	return cmd
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
			platforms := strings.Join(user.ConnectedPlatforms(), ", ")
			if platforms == "" {
				platforms = "None — open the Spot mobile app to link a booking platform."
			}
			_, _ = fmt.Fprintf(tw, "Connected\t%s\n", platforms)
			return tw.Flush()
		},
	}
}
