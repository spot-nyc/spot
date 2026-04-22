# Recipe: check state

**Use when:** the user asks "what do I have", "what's on my plate", "am I signed in", or any variant asking for a state summary.

## Flow

If you already ran the opening routine this session, replay the cached results. Otherwise, run in parallel:

```
spot auth whoami --json
spot reservations list --json
spot searches list --json
```

## Narrate the snapshot

Present the three buckets in plain language. Be terse — users who ask "what do I have?" want a scan, not a report.

Example output:

> **Signed in as** Brian (+1 555-555-5555) · Resy, OpenTable linked
>
> **Upcoming reservations**
> - Gramercy Tavern — Sat May 17, 7:00 PM, party of 2
> - Shuko — Thu May 22, 8:30 PM, party of 4
>
> **Active searches**
> - Don Angie + Via Carota + L'Artusi — Fri May 16, 6–9 PM, party of 2

If a bucket is empty, say so briefly: *"No upcoming reservations."* or *"No active searches."*

## When to mention history

Only mention `reservations history` if the user specifically asks about past dining or the conversation context makes it relevant (e.g. "have I been to Tatiana recently?"). Otherwise, history is noise in a state summary.

## Edge cases

- **`whoami` returns unauthenticated (exit 3):** stop the snapshot, tell the user to run `spot auth login`.
- **Zero reservations + zero searches:** *"You're not actively watching anything right now. Want to set up a search or find a table?"* — offer a next step.
- **One of the reservations is on an external platform (no search linkage):** still include it; the table column is the same shape.
