package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spot-nyc/spot/internal/updatecheck"
)

// updateCheckOptions is a test seam for runUpdateCheck. Production callers
// use runUpdateCheck, which wires in the real stderr, cache dir, and network
// endpoint.
type updateCheckOptions struct {
	currentVersion string
	cacheDir       string
	apiBaseURL     string // empty => use updatecheck package default
	grace          time.Duration
	stderr         io.Writer
	// done, if non-nil, is closed when the background goroutine finishes.
	// Tests that use a scratch cacheDir set this and wait on it before
	// returning so t.TempDir cleanup doesn't race writeCache. Production
	// callers leave it nil — whatever the goroutine hasn't finished by
	// os.Exit is discarded.
	done chan<- struct{}
}

// runUpdateCheck fires an update check in a background goroutine and prints
// a one-liner to stderr when a newer release is available. It waits at most
// `grace` for the check to complete before returning — if the check is slow,
// we simply return without a message (the user will see it on a later run
// when the cache is warm).
//
// Errors are swallowed: update checks must never block or disturb the user.
func runUpdateCheck(currentVersion string) {
	if updatecheck.ShouldSkip(currentVersion) {
		return
	}
	cacheDir, err := updateCacheDir()
	if err != nil {
		return
	}
	runUpdateCheckWithOptions(context.Background(), updateCheckOptions{
		currentVersion: currentVersion,
		cacheDir:       cacheDir,
		grace:          500 * time.Millisecond,
		stderr:         os.Stderr,
	})
}

func runUpdateCheckWithOptions(ctx context.Context, opts updateCheckOptions) {
	type result struct {
		latest    string
		available bool
	}

	ch := make(chan result, 1)
	go func() {
		if opts.done != nil {
			defer close(opts.done)
		}
		checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		latest, available := checkCachedResolved(checkCtx, opts)
		ch <- result{latest: latest, available: available}
	}()

	select {
	case r := <-ch:
		if r.available {
			_, _ = fmt.Fprintf(opts.stderr, "A new version %s is available — run `brew upgrade spot` (or your package manager of choice)\n", r.latest)
		}
	case <-time.After(opts.grace):
		// Exceeded grace period; skip the message silently.
	}
}

// checkCachedResolved picks between the real updatecheck.CheckCached and the
// test-only baseURL-overriding variant. We can't reach package-internal
// functions from main, so the two variants are exposed explicitly.
func checkCachedResolved(ctx context.Context, opts updateCheckOptions) (string, bool) {
	if opts.apiBaseURL != "" {
		return updatecheck.CheckCachedWithBaseURL(ctx, opts.currentVersion, opts.cacheDir, 24*time.Hour, opts.apiBaseURL)
	}
	latest, available, err := updatecheck.CheckCached(ctx, opts.currentVersion, opts.cacheDir, 24*time.Hour)
	if err != nil {
		return "", false
	}
	return latest, available
}

// updateCacheDir returns $XDG_CACHE_HOME/spot (or $HOME/.cache/spot as a
// fallback per XDG spec).
func updateCacheDir() (string, error) {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "spot"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cache", "spot"), nil
}
