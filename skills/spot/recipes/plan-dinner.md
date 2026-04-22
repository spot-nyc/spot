# Recipe: plan dinner

**Use when:** the user wants help picking a restaurant, not a specific one. "Plan a dinner", "find me something good", "surprise me" all match.

## Flow

### 1. Gather constraints

Ask (or infer from context) the minimum set:
- Date — confirm absolute date.
- Time window — propose 6:00–9:00 PM if unspecified.
- Party size.
- Neighborhood and/or cuisine preference.
- Any hard filters: vegetarian, kid-friendly, on a particular platform, etc.

### 2. Discover candidates

Search by the strongest signal:

```
spot restaurants search "flatiron" --json
```

or

```
spot restaurants search "italian" --json
```

You'll often need 2–3 queries to get a reasonable candidate set. Merge results in memory; dedupe by ID.

### 3. Shortlist to 3–5 candidates

From the candidate set, pick the top 3–5 based on:
- Name / reputation match to user's preference.
- Platform fit (user has that platform connected).
- Party-size fit (check `minimumPartySize` / `maximumPartySize` from `restaurants get`).

Optionally call `spot restaurants get <id>` on each top candidate to see cuisine, hours, address if that helps differentiate.

### 4. Search availability across the shortlist in one call

```
spot reservations search \
  --restaurant rst_a,rst_b,rst_c,rst_d,rst_e \
  --date 2026-05-15 \
  --start-time 18:00 \
  --end-time 21:00 \
  --party 4 \
  --json
```

### 5. Present ranked options

Narrate 2–4 options to the user. For each, call out:
- Restaurant name + cuisine + neighborhood.
- Time slot (closer to user's stated preference ranks higher).
- Seating (Dining Room beats Bar for most users, but check context).
- Platform (if the user prefers one).

Example:

> "Here are three options for Saturday night:
> - **Lodi** (Italian, Rockefeller Plaza) — 7:00 PM, Dining Room, Resy.
> - **Via Carota** (Italian, West Village) — 6:30 PM, Bar, OpenTable.
> - **Raoul's** (French, SoHo) — 8:15 PM, Dining Room, Resy.
>
> Lodi's time is the best fit; Raoul's is later but a quieter vibe. Want Lodi, or one of the others?"

### 6. Book on confirmation

Once the user picks, book per the find-and-book flow (step 4a).

## Edge cases

- **All candidates have zero availability:** pivot to `recipes/monitor-drop.md` — offer to set an autobook search covering the 5 best shortlisted restaurants.
- **User has an existing search covering the same window:** flag it — *"You already have a search watching Don Angie, Via Carota, and L'Artusi for this same window. Should I pick from those first, or add Lodi to that search?"*
- **Shortlist too narrow:** if `restaurants search` returns fewer than 3 results for all queries, broaden ("italian" → "pasta" → "european"). Ask the user for more signals if still stuck.
