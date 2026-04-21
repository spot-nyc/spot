# Changelog

All notable changes to the Spot SDK will be documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-04-21

First tagged release of the Spot SDK.

### Added

#### Authentication
- `spot auth login` — browser-based PKCE OAuth sign-in. Persists credentials to the OS keyring (macOS Keychain, libsecret on Linux, Windows Credential Manager) with a file fallback at `$XDG_DATA_HOME/spot/credentials.json`.
- `spot auth logout` — clears stored credentials (idempotent).
- `spot auth whoami` — prints the currently-authenticated user profile.
- Automatic access-token refresh via `auth.NewRefreshingTokenSource`.

#### Searches, reservations, and restaurants
- `spot searches list|get|create|delete` — manage reservation searches. Times accept both `HH:MM` and `HH:MM:SS`. Full IDs shown in the list for copy-paste into `get`/`delete`.
- `spot reservations list|cancel` — view and cancel booked reservations. Seating renders as "Dining Room" (OpenTable's default), "Bar", etc., mirroring the mobile client.
- `spot restaurants search <query>` — look up restaurants by name. Columns: `ID`, `NAME`, `CUISINE`, `NEIGHBORHOOD`, `PLATFORMS` (derived from the real `resyActive` / `openTableActive` / `sevenRoomsActive` / `doorDashActive` flags).

#### Output and exit codes
- Global `--json` / `-j` flag on every command; stdout format auto-detects based on TTY.
- Stable, documented exit codes mapped from library error sentinels: 3 unauth, 4 expired, 5 not-found, 6 conflict, 7 validation, 8 rate-limit, 9 server.
- Friendlier table-mode messages for `ErrUnauthenticated` and `ErrAuthExpired` (suggests running `spot auth login`).

#### Update check
- Automatic non-blocking update check on every command exit. Hits `api.github.com/repos/spot-nyc/spot/releases/latest` at most once every 24 hours (cached at `$XDG_CACHE_HOME/spot/update.json`). Prints a one-liner to stderr when a newer release is available.
- Opt out via `SPOT_NO_UPDATE_CHECK=1`. Automatically skipped in CI (`CI`, `GITHUB_ACTIONS`, `BUILDKITE`) and for `dev` builds.

#### Go library (`github.com/spot-nyc/spot`)
- Typed `*spot.Error` with 10 sentinel values (`ErrUnauthenticated`, `ErrAuthExpired`, `ErrNotFound`, `ErrConflict`, `ErrValidation`, `ErrRateLimit`, `ErrServer`, `ErrNetwork`, `ErrTimeout`, `ErrClient`).
- HTTP client with functional options, 30s default timeout, JSON encode/decode, error response mapping.
- Services: `UsersService`, `SearchesService`, `ReservationsService`, `RestaurantsService`.
- `auth` package helpers: `Store` interface, `EnvStore`, `FileStore` (XDG path, 0600 perms), `KeyringStore`, `ChainedStore`, `DefaultStore()`, `Login()` (full PKCE flow with loopback server and browser opener), `NewRefreshingTokenSource`.

### Infrastructure

- Per-PR CI: `go test -race`, `go vet`, `golangci-lint run`, `goreleaser release --snapshot`.
- Tag-triggered release workflow publishes cross-platform binaries (`darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`, `windows/amd64`) to GitHub Releases with SHA256 checksums, an auto-generated formula in `spot-nyc/homebrew-tap`, and an auto-generated manifest in `spot-nyc/scoop-bucket`.

### Install

- `go install github.com/spot-nyc/spot/cmd/spot@latest`
- `brew install spot-nyc/tap/spot`
- `scoop bucket add spot-nyc https://github.com/spot-nyc/scoop-bucket && scoop install spot`
- `curl -fsSL https://raw.githubusercontent.com/spot-nyc/spot/main/install.sh | sh`

[Unreleased]: https://github.com/spot-nyc/spot/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/spot-nyc/spot/releases/tag/v0.1.0
