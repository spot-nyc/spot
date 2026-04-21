package updatecheck

import (
	"context"
	"time"
)

// CheckCached wraps Check with a file-based cache at cacheDir/update.json.
// If the cache is less than ttl old AND was recorded for currentVersion, the
// cached result is returned without a network call. Otherwise Check runs and
// the result is persisted.
//
// Errors are returned but callers should typically ignore them: update
// checks must never block output or noisily fail. A missing or corrupt cache
// is treated as a cache miss and silently refreshed on the next call.
func CheckCached(ctx context.Context, currentVersion, cacheDir string, ttl time.Duration) (latest string, available bool, err error) {
	return checkCachedWithBaseURL(ctx, currentVersion, cacheDir, ttl, githubReleasesAPI)
}

func checkCachedWithBaseURL(ctx context.Context, currentVersion, cacheDir string, ttl time.Duration, baseURL string) (string, bool, error) {
	if entry, err := readCache(cacheDir); err == nil && entry.CurrentVersion == currentVersion && time.Since(entry.CheckedAt) < ttl {
		return entry.Latest, isNewer(currentVersion, entry.Latest), nil
	}

	latest, available, err := checkWithBaseURL(ctx, currentVersion, baseURL)
	if err != nil {
		return "", false, err
	}

	_ = writeCache(cacheDir, cacheEntry{
		CheckedAt:      time.Now().UTC(),
		Latest:         latest,
		CurrentVersion: currentVersion,
	})

	return latest, available, nil
}

// CheckCachedWithBaseURL is CheckCached with an overridable base URL for
// tests. Production code uses CheckCached. Returns (latest, available) and
// swallows errors — callers that care about errors should use CheckCached.
func CheckCachedWithBaseURL(ctx context.Context, currentVersion, cacheDir string, ttl time.Duration, baseURL string) (string, bool) {
	latest, available, err := checkCachedWithBaseURL(ctx, currentVersion, cacheDir, ttl, baseURL)
	if err != nil {
		return "", false
	}
	return latest, available
}
