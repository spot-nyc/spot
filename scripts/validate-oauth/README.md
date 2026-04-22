# validate-oauth

Manual end-to-end validation of the Spot SDK's OAuth 2.1 flow.

## Usage

```bash
cd ~/Desktop/dev/aui/sdk
go run ./scripts/validate-oauth
```

The program:

1. Runs `auth.DefaultLogin` — opens your browser to Supabase's authorize URL.
2. You complete phone OTP sign-in + consent on Spot Pro.
3. Program receives the callback, exchanges code for tokens.
4. Program calls `GET /users/me` on the Spot API with the issued access token.
5. Prints the user profile on success.

## Prerequisites

- Spot Pro running at the configured Supabase "Site URL" (for local dev: `http://localhost:3000` via `pnpm dev`).
- Supabase "Site URL" matches where Pro is running.
- The Spot CLI OAuth client is registered with all 20 loopback redirect URIs (`127.0.0.1` and `localhost` × ports 52853–52862).

## Environment variables

- `SPOT_BASE_URL` — override the Spot API base URL. Default: `https://api.spot.nyc`.

## Expected output

```
Opening browser for OAuth sign-in …
Complete the flow in your browser; this program will continue automatically.

✓ Login complete.
  access_token (first 20 chars): eyJhbGciOiJIUzI1NiIs…
  has refresh_token: true
  expiry: 2026-04-18 15:42:13 +0000 UTC

Calling GET /users/me …
✓ /users/me returned:
  {
    "id": "user-uuid-here",
    "phone": "+15555551234",
    "name": "..."
  }

✓ Validation successful. The full OAuth stack works end-to-end.
```

## Failure modes

- **"open browser" error** — your environment has no default browser. Run on a machine with a GUI.
- **"all pre-registered loopback ports are in use"** — another process is holding one of ports 52853–52862. Close it and retry.
- **"state mismatch"** — the browser somehow sent a different state than the CLI generated. Rerun.
- **"users.me: unauthenticated"** — the Spot API rejected the OAuth-issued access token. See M1c plan Task 9 for API middleware diagnosis steps.
