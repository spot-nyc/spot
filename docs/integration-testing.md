# Integration Testing

The integration suite hits a real Spot API with a dedicated test user to catch
contract drift between SDK and server that unit tests cannot. Gated behind a
Go build tag so the default test run stays offline.

## Running locally

Prereqs: valid credentials for the dedicated CI test user.

```bash
SPOT_TEST_ACCESS_TOKEN=<user access token> \
go test -tags=integration -v ./integration/...
```

Optional env:
- `SPOT_TEST_BASE_URL` — point at a non-prod instance (default: `https://api.spot.nyc`).
- `SPOT_TEST_USER_ID` — if set, the whoami test asserts the returned profile matches.

Unset `SPOT_TEST_ACCESS_TOKEN` → TestMain skips cleanly (exit 0).

## CI setup

Runs in the Release workflow before goreleaser, plus on `workflow_dispatch`
for manual invocations between releases. Auth uses a long-lived refresh
token stashed as a GitHub secret; the workflow rotates it on every run.

### Required GitHub Secrets

| Secret | What it is | Who sets it |
|---|---|---|
| `SPOT_TEST_REFRESH_TOKEN` | Supabase refresh token for the dedicated test user. Self-rotated by the workflow. | One-time human setup; workflow thereafter. |
| `SPOT_TEST_USER_ID` | UUID of the test user, for whoami assertions. | Human, permanent. |
| `SUPABASE_URL` | Supabase project URL (e.g. `https://xyz.supabase.co`). | Human, permanent. |
| `SUPABASE_ANON_KEY` | Supabase anon/publishable key, sent as `apikey` header on the refresh call. | Human, permanent. |
| `SPOT_CI_PAT` | GitHub PAT with `actions:write` on this repo, used by the workflow to overwrite `SPOT_TEST_REFRESH_TOKEN`. | Human, rotate periodically. |

### Workflow behavior

On every integration-test run:

1. Curl `${SUPABASE_URL}/auth/v1/token?grant_type=refresh_token` with the
   stashed refresh token → new `access_token` + `refresh_token`.
2. `gh secret set SPOT_TEST_REFRESH_TOKEN` with the new refresh token. This
   persists *before* the tests run so a test failure doesn't lose the new
   token (old one is already invalidated by Supabase).
3. Run `go test -tags=integration ./integration/...` with the new access
   token in `SPOT_TEST_ACCESS_TOKEN`.

### One-time setup

1. Have the test user (e.g. `spot-ci@anthropic.com`) go through the OAuth
   flow from the Spot mobile app or CLI once.
2. Extract their refresh token from the keyring or credential file.
3. `gh secret set SPOT_TEST_REFRESH_TOKEN --body "<refresh token>"` in the
   SDK repo.
4. Set the other permanent secrets listed above.
5. Create a fine-grained PAT with `Actions:write` permission on this repo
   only, save it as `SPOT_CI_PAT`.

If the workflow ever loses the rotation (e.g. a failed `gh secret set`
between the refresh and the test run), integration tests will start failing
with auth errors. Recovery: repeat the one-time setup with a fresh login.

## Coverage

Covered:
- Auth: whoami roundtrip.
- Searches: full CRUD lifecycle.
- Restaurants: search + get.
- Reservations: list, history, search (read-only).

**Not covered (out of scope for v0.4.0):**
- `reservations book` / `cancel` — would hold or release real tables on
  booking platforms. Unit tests + manual smoke cover these until a
  server-side test-mode booking mock exists.
- `auth logout` server-side revocation — logging out the test session's
  access token breaks subsequent tests. Covered by unit tests + manual
  smoke; integration coverage is tracked as future work.

## Known gaps

- Token rotation is brittle if Supabase's rotation semantics change or if
  two concurrent workflow runs invalidate each other's token. Bounded in
  practice because the suite only runs on release tags and manual dispatch.
- Restaurant IDs used for search-based tests are fetched at runtime from
  `restaurants search "gramercy"`. If that query ever returns zero results
  (unlikely), broaden the `searchProbe` constant in `helpers_test.go`.
