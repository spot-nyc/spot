# Dogfood log — Spot SDK M6 (v0.4.0)

Running journal of dogfood sessions during v0.4.0 development. Each entry captures a real-world use of the `skills/spot` skill against the production Spot API, and what it surfaced.

## Graduation criterion

Ship v0.4.0 when **5 consecutive** booking sessions across **separate days** succeed end-to-end without the user having to correct the agent. "Normal" intent (book-specific, or plan-and-book), typical NYC restaurants, not adversarial edge cases.

## Template

```
## YYYY-MM-DD — <one-line goal>
**Goal:** <what the user asked for>
**Flow:** <which recipe(s) the agent matched>
**Outcome:** <booked / set search / abandoned / etc.>
**Worked:** <what went right>
**Friction:** <what was awkward, confusing, or broken>
**Fix:** <commit sha, skill edit, or follow-up>
```

## Entries

<!-- Append new entries here. Most recent at the top. -->
