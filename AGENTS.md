# Repository Guidelines

## Project Overview

This repository contains the Spot Go SDK and `spot` CLI for reservations,
searches, restaurant lookup, and auth workflows on the Spot platform.

Primary areas:

- Root package `github.com/spot-nyc/spot`: SDK client, service types, errors,
  options, and tests.
- `cmd/spot`: Cobra CLI commands and CLI-specific tests.
- `internal/`: private helpers such as rendering, TTY handling, and update
  checks.
- `integration/`: live API integration tests gated behind the `integration`
  build tag.
- `skills/spot`: Claude Code skill shipped with the repo.
- `docs/`: integration testing notes and dogfood logs.

## Common Commands

- `go test ./...`: run default unit tests.
- `make test`: run unit tests with race detector and write `coverage.out`.
- `make lint`: run `golangci-lint run ./...`.
- `make build`: build the CLI to `dist/spot`.
- `make all`: run test, lint, then build.
- `go test -tags=integration -v ./integration/...`: run live integration
  tests when credentials are available.

## Development Notes

- Keep SDK behavior in the root package and CLI presentation/argument parsing
  in `cmd/spot`.
- Prefer table-driven tests and `httptest` for HTTP behavior. Most tests should
  stay offline.
- Run `gofmt` on touched Go files before finishing.
- Keep public SDK API changes intentional and document-facing. Root package
  exported names are part of the SDK surface.
- Avoid broad refactors unless they are necessary for the requested change.

## Integration Test Safety

The integration suite talks to a real Spot API and requires
`SPOT_TEST_ACCESS_TOKEN`. Optional variables are `SPOT_TEST_BASE_URL` and
`SPOT_TEST_USER_ID`; see `docs/integration-testing.md`.

Important cautions:

- Do not invent, print, or commit tokens or refresh tokens.
- Do not run live integration tests unless explicitly needed and credentials
  are present.
- `integration/TestMain` cleans up existing searches for the dedicated test
  user before running. Treat that suite as stateful and live-api-affecting.
- Booking and cancellation against real tables are intentionally not covered by
  integration tests.

## Generated And Local Artifacts

- `coverage.out` is produced by `make test`.
- `dist/` is produced by `make build` and release tooling.
- Do not remove or overwrite unrelated artifacts or user changes unless the
  user explicitly asks.
