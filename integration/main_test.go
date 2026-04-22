//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/spot-nyc/spot"
)

// TestMain runs before any test in the integration package.
//
// Responsibilities:
//  1. If secrets are missing, skip cleanly (exit 0) so a CI misconfiguration
//     or a local `-tags=integration` invocation doesn't emit false positives.
//  2. Delete any searches left behind by a previous failed run — each test
//     creates disposable searches and cleans them up via t.Cleanup, but a
//     panicked test or a kill -9 can leave stragglers. We wipe the slate so
//     every run starts from a clean state on the dedicated CI test user.
//  3. Run the suite.
func TestMain(m *testing.M) {
	if !integrationEnabled() {
		logSkip(fmt.Sprintf("%s not set", envAccessToken))
		os.Exit(0)
	}

	if err := cleanStragglers(); err != nil {
		fmt.Fprintf(os.Stderr, "[integration] straggler cleanup failed (continuing): %v\n", err)
	}

	os.Exit(m.Run())
}

func cleanStragglers() error {
	token := os.Getenv(envAccessToken)
	opts := []spot.Option{spot.WithToken(token)}
	if base := os.Getenv(envBaseURL); base != "" {
		opts = append(opts, spot.WithBaseURL(base))
	}
	client, err := spot.NewClient(opts...)
	if err != nil {
		return err
	}

	ctx := context.Background()
	searches, err := client.Searches.List(ctx)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "[integration] deleting %d straggler searches\n", len(searches))
	for _, s := range searches {
		if err := client.Searches.Delete(ctx, s.ID); err != nil {
			fmt.Fprintf(os.Stderr, "[integration]   failed to delete %s: %v\n", s.ID, err)
		}
	}
	return nil
}
