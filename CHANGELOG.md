# Changelog

All notable changes to the Spot SDK will be documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-04-22

### Added
- `spot searches update <id>` — modify an existing search. `--party`, `--date`, `--start-time`, `--end-time`, `--restaurant`. Only explicitly set flags are sent; unset fields are left untouched on the server.
- `spot restaurants get <id>` — detail view for a single restaurant: name, cuisine, neighborhood, address, phone, website, platforms, party limits, booking difficulty, description.
- `spot reservations search` — find reservations available to book right now. Accepts one or more restaurant IDs (`--restaurant rst_a,rst_b` or repeated). Returns slots with a short TTL and IDs to pass to `book`.
- `spot reservations book <slotId>` — book a slot returned by `reservations search`. Returns the resulting `Reservation`.
- `spot update` — detects how spot was installed (Homebrew / Scoop / `go install` / curl installer) and prints the matching upgrade command. Never executes; copy-paste into your shell.
- Go library:
  - `SearchesService.Update`, `RestaurantsService.Get`, `ReservationsService.Search`, `ReservationsService.Book`.
  - `ReservationSlot` type and `SearchReservationsParams` / `UpdateSearchParams`.
  - `Restaurant` gains: `Description`, `Hours`, `Phone`, `Website`, `ResyURL`, `OpenTableURL`, `SevenRoomsURL`, `DoorDashURL`, `MinimumPartySize`, `MaximumPartySize`, `BookingDifficulty`, `BookingDifficultyDetails`.
  - New error sentinels: `ErrSlotExpired` (HTTP 410), `ErrPlatformNotConnected` (HTTP 412, carries a `Platform` field).
- New CLI exit codes: **10** (`ErrPlatformNotConnected`), **11** (`ErrSlotExpired`). Table-mode messages guide users to the mobile app for platform linking.

### Changed
- Reservation booking is now available to any authenticated Spot user. Morty grows user-auth endpoints for synchronous search and book at `/reservations/search` and `/reservations/book`.

### Fixed
- `spot searches create` / `SearchesService.Create` were shipping the restaurant list under the JSON key `restaurants`, which morty rejects — the endpoint expects `restaurantIds`. In v0.1.0 the command returned a 400 from the server; the SDK mock-tests matched the SDK's bad output rather than the real API. The Go field is now `CreateSearchParams.RestaurantIDs` (JSON tag `restaurantIds`), matching `UpdateSearchParams.RestaurantIDs`.
- `spot searches delete` / `SearchesService.Delete` was calling `DELETE /searches/:id`, an endpoint that does not exist. Morty soft-deletes via `POST /searches/:id` with a non-null `deletedAt` timestamp. The SDK now performs that POST with a fresh RFC3339 timestamp; the public Go API (`Delete(ctx, id)`) is unchanged.
- Error messages from Hono's `HTTPException` (e.g. "Slot is no longer available", "Restaurant not found") are now surfaced intact instead of being replaced by `http.StatusText(code)` ("Gone", "Not Found"). The SDK previously only parsed JSON error bodies; Hono emits `text/plain` by default.

### Removed (breaking)
- `Search.Upgrade` and `SearchTarget.Upgrade` fields are no longer parsed or exposed by the SDK. The upgrade feature remains a morty-internal concern; it will return to the SDK surface in a later release if a use case emerges. Consumers on v0.1.0 that read these fields should drop the references — JSON consumers won't crash, but the fields won't be populated.

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

[Unreleased]: https://github.com/spot-nyc/spot/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/spot-nyc/spot/releases/tag/v0.2.0
[0.1.0]: https://github.com/spot-nyc/spot/releases/tag/v0.1.0
