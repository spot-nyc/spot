package updatecheck

import "os"

// ShouldSkip reports whether the update check should be bypassed. Returns
// true for development builds (no useful version to compare), when the user
// has set SPOT_NO_UPDATE_CHECK=1, or when common CI environment variables
// are set (CI, GITHUB_ACTIONS, BUILDKITE) — CI runs shouldn't emit upgrade
// nags into logs.
//
// The "0" value for SPOT_NO_UPDATE_CHECK is intentionally not treated as
// opt-out so a user can override an inherited-from-parent-shell "1" with
// "SPOT_NO_UPDATE_CHECK=0 spot ...".
func ShouldSkip(currentVersion string) bool {
	if currentVersion == "" || currentVersion == "dev" {
		return true
	}
	if os.Getenv("SPOT_NO_UPDATE_CHECK") == "1" {
		return true
	}
	for _, key := range []string{"CI", "GITHUB_ACTIONS", "BUILDKITE"} {
		if os.Getenv(key) != "" {
			return true
		}
	}
	return false
}
