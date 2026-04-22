---
name: spot
description: Find and book NYC restaurant reservations via the Spot CLI. Use when the user asks about dining plans, reservations, restaurants, or what they currently have booked.
---

# Spot

## What Spot is

Spot is a CLI-first NYC restaurant reservation service. It hunts the big booking platforms — Resy, OpenTable, SevenRooms, and DoorDash — on the user's behalf. The model has two tracks:

1. **Book-now.** If a desirable slot exists this moment, book it.
2. **Autobook search.** If nothing is available, create a *search* targeting up to 5 restaurants over a date/time window; the server grabs the first matching slot as soon as one drops.

Both tracks are first-class. Know which one fits the user's situation before reaching for a command.

## Core concepts

- **Restaurant** — a venue in the Spot catalog. Discover with `spot restaurants search <query>`; get detail with `spot restaurants get <id>`.
- **Reservation** — a booked table the user holds. Upcoming via `spot reservations list`; full log (past + upcoming + external platforms) via `spot reservations history`.
- **Search** — the user's standing autobook request. Has a date/time window, party size, and **up to 5 restaurant targets**. The first target that drops a matching slot gets booked.
- **SearchTarget** — one restaurant attached to a search.
- **Slot** — an immediate-booking opportunity returned by `spot reservations search`. Each slot has a **~5-minute TTL** on the server. If the user hesitates, re-run search.
- **Platform connection** — booking requires the user to have linked their Resy / OpenTable / SevenRooms / DoorDash account via the Spot **mobile app**. Cannot be linked from the CLI. If a book call returns `ErrPlatformNotConnected` (exit 10), stop and tell the user to open the mobile app.
- **Time zone** — every date and time in the CLI is `America/New_York`. Dates are `YYYY-MM-DD`; times are `HH:MM` or `HH:MM:SS`.

## Opening routine

On the first relevant query in a session, run these three commands **in parallel** with `--json` and cache the results for the rest of the session:

```
spot auth whoami --json
spot reservations list --json
spot searches list --json
```

Use the snapshot to inform every suggestion:

- Don't propose booking a slot the user already holds.
- Don't create a new search if an existing one already covers the intent — offer to update it instead.
- If `auth whoami` fails, tell the user to run `spot auth login` and wait.
- Check `connectedPlatforms` on the whoami result — if the platform needed for a booking isn't connected, warn early instead of letting `book` fail.

## Decision tree

### Intent: "book <specific restaurant> on <date> at <time> for <N>"

1. Resolve the restaurant ID via `spot restaurants search <name>` if not already in cache.
2. `spot reservations search --restaurant <id> --date <YYYY-MM-DD> --start-time HH:MM --end-time HH:MM --party <N> --json`.
3. If slots exist: pick the closest match to the stated time. Narrate it in plain language and ask confirmation: *"About to book Gramercy Tavern on May 15 at 7:00 PM for 2 people on Resy. Proceed?"*. On confirmation, `spot reservations book <slotId>`.
4. If no slots: offer to create an autobook search. Suggest adding similar-vibe fallback restaurants (up to 5 total) so the user has a real chance of a match. Use `spot searches create --restaurant a,b,c --date ... --start-time ... --end-time ... --party ...`.

### Intent: "plan <something> in <neighborhood> on <date>" (research mode)

1. `spot restaurants search <query>` for candidates — query by neighborhood, cuisine, or vibe word.
2. Narrow to 3–5 top candidates. Optionally `spot restaurants get <id>` on each to read cuisine, party limits, hours, and active platforms.
3. `spot reservations search --restaurant a,b,c --date ... --start-time ... --end-time ... --party <N>` across candidates in one call.
4. Present ranked options: time fit, seating, platform, any other differentiators. Let the user pick.
5. Confirm + book the chosen slot, same as the book flow.

### Intent: "what do I have?"

Replay the opening-routine snapshot. Summarize upcoming reservations, active searches, and connected platforms in plain language. No commands needed beyond the snapshot.

### Platform-not-connected error (exit code 10)

Stop the booking flow. Tell the user explicitly which platform needs linking ("your Resy account isn't linked to Spot — open the Spot mobile app to connect it"). **Do not silently try other platforms.** The user may prefer that specific restaurant.

### Slot-expired error (exit code 11)

Recoverable. Re-run `spot reservations search` with the original params, pick the closest match to the original intent, confirm with the user, and try `book` again. If it expires twice, surface the issue — likely a very hot room.

## Autonomy rules

Different actions warrant different levels of ceremony.

**Confirm in plain language before:**
- `spot reservations book` — holds a real table.
- `spot reservations cancel` — releases a real table and may incur cancellation fees.
- `spot searches delete` — removes the autobook hunt.

Canonical phrasing: *"About to <verb> <entity> at <restaurant> on <date> at <time> for <party>. Proceed?"*

**Auto with a brief announcement (reversible, low-cost failure):**
- `spot searches create`
- `spot searches update`

Example: *"Setting up a search for Don Angie, Via Carota, and L'Artusi on May 15, 6–8 PM, party of 2."* Then run.

**Silent (no ceremony needed — read-only):**
- `spot auth whoami`
- `spot reservations list`, `history`, `search`
- `spot restaurants search`, `get`
- `spot searches list`, `get`

Never proactively run `spot auth logout` or `spot auth logout --all`. Those are user-initiated only.

## Errors and exit codes

The CLI uses stable exit codes. The most common ones:

| Code | Sentinel | Meaning | Action |
|---|---|---|---|
| 0 | — | Success | — |
| 3 | `ErrUnauthenticated` | Not logged in / session expired | Tell user to run `spot auth login` |
| 5 | `ErrRestaurantNotFound`, `ErrSearchNotFound`, `ErrReservationNotFound` | Entity not found | Name the specific missing entity |
| 7 | `ErrValidation` | Bad request | Show the server's message verbatim |
| 10 | `ErrPlatformNotConnected` | Booking platform not linked | Direct user to mobile app for that platform |
| 11 | `ErrSlotExpired` | Slot TTL expired or snapped up | Re-search transparently; retry once |

For any non-zero exit not listed: show the server's message. Don't fabricate explanations.

## Format conventions

- **Dates:** `YYYY-MM-DD`. Always resolve relative phrases ("this Friday", "next weekend") to an absolute date and **confirm with the user** before running commands.
- **Times:** `HH:MM` (CLI normalizes) or `HH:MM:SS`. All America/New_York.
- **Party size:** positive integer.
- **Restaurant flag:** `--restaurant` accepts a comma-separated list or can be repeated. Max 5 per search.
- **`--json` is authoritative.** Table mode is pretty for humans; use `--json` when parsing output.

## Recipes

Pattern library for common flows. Read the recipe that matches the user's intent before acting:

- `recipes/find-and-book.md` — book a specific restaurant / date, or fall back to an autobook search.
- `recipes/plan-dinner.md` — research + narrow + book across neighborhoods.
- `recipes/monitor-drop.md` — no availability? Set a search and explain autobook.
- `recipes/check-state.md` — snapshot of upcoming + searches + connected platforms.
- `recipes/cancel-and-rebook.md` — change plans cleanly.
