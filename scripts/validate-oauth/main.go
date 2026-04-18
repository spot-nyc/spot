// Command validate-oauth exercises the full OAuth login flow end-to-end.
//
// It opens a browser, guides the user through the Supabase consent flow on
// Spot Pro, receives the callback on a local loopback port, exchanges the
// code for tokens, and finally calls morty's /users/me endpoint to confirm
// the access token is accepted by morty's JWT middleware.
//
// Manual run — requires a real browser and a human to complete OTP sign-in.
//
//	cd ~/Desktop/dev/aui/sdk
//	go run ./scripts/validate-oauth
//
// Environment overrides (optional):
//
//	SPOT_BASE_URL — override morty API base URL (default: https://api.spot.nyc)
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spot-nyc/spot"
	"github.com/spot-nyc/spot/auth"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "FAIL:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	fmt.Println("Opening browser for OAuth sign-in …")
	fmt.Println("Complete the flow in your browser; this program will continue automatically.")
	fmt.Println()

	creds, err := auth.DefaultLogin(ctx, auth.LoginOptions{
		Timeout: 5 * time.Minute,
	})
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}

	fmt.Println("✓ Login complete.")
	fmt.Printf("  access_token (first 20 chars): %s…\n", truncate(creds.AccessToken, 20))
	fmt.Printf("  has refresh_token: %v\n", creds.RefreshToken != "")
	fmt.Printf("  expiry: %v\n", creds.Expiry)
	fmt.Println()

	baseURL := os.Getenv("SPOT_BASE_URL")
	opts := []spot.Option{spot.WithToken(creds.AccessToken)}
	if baseURL != "" {
		opts = append(opts, spot.WithBaseURL(baseURL))
	}

	client, err := spot.NewClient(opts...)
	if err != nil {
		return fmt.Errorf("new client: %w", err)
	}

	fmt.Println("Calling GET /users/me …")
	user, err := client.Users.Me(ctx)
	if err != nil {
		return fmt.Errorf("users.me: %w", err)
	}

	fmt.Println("✓ /users/me returned:")
	j, _ := json.MarshalIndent(user, "  ", "  ")
	fmt.Printf("  %s\n", j)
	fmt.Println()

	fmt.Println("✓ Validation successful. The full OAuth stack works end-to-end.")
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
