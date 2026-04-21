// Package updatecheck checks whether a newer release of spot is available on
// GitHub and caches the result locally so the network request only runs once
// per day. All functions are silent on errors: update checks must never block
// or noisily fail.
package updatecheck

import "golang.org/x/mod/semver"

// isNewer reports whether latest is a strictly higher semver than current.
// Returns false when either version fails to parse as semver, and also when
// current is the sentinel "dev" (development builds installed via `go install`
// or unflagged `go build`) — those users don't want upgrade nags.
func isNewer(current, latest string) bool {
	if current == "" || current == "dev" {
		return false
	}
	c := ensurePrefix(current)
	l := ensurePrefix(latest)
	if !semver.IsValid(c) || !semver.IsValid(l) {
		return false
	}
	return semver.Compare(l, c) > 0
}

// ensurePrefix adds a leading "v" when missing. golang.org/x/mod/semver
// requires the "v" prefix on inputs.
func ensurePrefix(s string) string {
	if s == "" || s[0] == 'v' {
		return s
	}
	return "v" + s
}
