# Changelog

All notable changes to the Spot SDK will be documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial repo scaffolding (M0).
- Cobra CLI root with `--version` and `--help`.
- CI pipeline (test, lint, goreleaser snapshot).
- Typed `*spot.Error` with 10 sentinel values (M1a).
- HTTP `Client` with functional options, 30s default timeout, JSON encode/decode, error response mapping (M1a).
- `auth.Store` interface + `EnvStore` backend reading `SPOT_TOKEN` (M1a).
- `UsersService.Me` — first concrete service method (M1a).
- `auth.FileStore` (XDG path, 0600 perms) and `auth.KeyringStore` (go-keyring) (M1b).
- `auth.ChainedStore` + `auth.DefaultStore()` resolving env → keyring → file (M1b).
- `auth.Login(ctx, cfg, opts)` — full PKCE authorization code flow with loopback server and browser opener (M1b).
- `auth.NewRefreshingTokenSource` — auto-refresh via `oauth2.Config.TokenSource`, persists rotated tokens (M1b).
- Real `SupabaseProjectRef` and `ClientID` in `auth/constants.go` pinned (M1c).
- `scripts/validate-oauth/` — manual end-to-end harness that exercises the full OAuth stack against real infra (M1c).

### Validated (M1c)
Full OAuth flow works end-to-end against real Supabase + real morty:
- Morty's existing HS256 JWT middleware accepts Supabase OAuth-issued access tokens unchanged — no morty changes needed.
- `GET /users/me` shape matches `spot.User` (`id`, `phone`, `name`).
- Supabase honors the 10 pre-registered loopback redirect URIs (`127.0.0.1:52853–52862`).
- PKCE S256 challenge/verifier round-trips correctly.
- Refresh token included in response (rotation behavior to validate once tokens age past 1h).

### Infrastructure (M1c)
- "Spot CLI" public OAuth 2.1 client registered on Supabase (client type: public, `token_endpoint_auth_method: none`).
- Supabase project's `authorization_url_path` set to `/oauth/consent`.
- Spot Pro ships consent UI at `/oauth/consent` and decision handler at `/api/oauth/decision`.
- Spot Pro sign-in preserves `?redirect=` query param through phone + OTP flow.

### Added (M2 — first user-visible commands)
- `spot auth login` — browser-based PKCE sign-in, persists credentials to keyring/file.
- `spot auth logout` — deletes locally-stored credentials (idempotent).
- `spot auth whoami` — prints the currently-authenticated user profile.
- `spot searches list` — lists active reservation searches with human-friendly formatting (truncated IDs, `May 1, 2026` dates, `6:00 PM–9:00 PM` times, restaurant names).
- Global `--json` / `-j` flag on every command; stdout format auto-detects based on TTY.
- Stable, documented exit codes mapped from library error sentinels (3 unauth, 4 expired, 5 not-found, 6 conflict, 7 validation, 8 rate-limit, 9 server).
- Friendlier table-mode messages for `ErrUnauthenticated` and `ErrAuthExpired` that tell users to run `spot auth login`.
- `internal/tty` and `internal/render` helpers (TTY detection + JSON/table writing via text/tabwriter).
- `SearchesService.List` + `Search` / `SearchTarget` / `Restaurant` types in the library.
- `SPOT_BASE_URL` env var for overriding the API base URL (primarily for testing).
- End-to-end CLI integration tests covering the full `cobra → Client → httptest → render` pipeline.
