# Recipe: monitor for a drop

**Use when:** no slots are currently available, or the user explicitly asks to "watch for", "grab the first opening at", etc.

## Flow

### 1. Explain autobook if needed

Many users haven't encountered this pattern. Briefly explain:

> "Spot can watch up to 5 restaurants and automatically book the first matching slot that drops. You'll get a notification in the mobile app when it lands. Want me to set that up?"

Skip the explanation if the user clearly already knows the model (e.g. they said "set a search" or "watch for a table").

### 2. Build the 5-restaurant set

Quality of the autobook outcome depends on the candidates. Help the user assemble a set of up to 5 restaurants that they'd genuinely be happy with:

- Start from their stated preference (one restaurant → add 4 similar).
- Or use the plan-dinner flow's shortlisting to pick 5 from a broader search.
- **Draw on history as a source of likely-fit fallbacks.** If the stated restaurant is Italian in Flatiron, suggest adding cuisines / neighborhoods the user has booked before in the same vibe. Mention why: *"I'm adding Via Carota and Don Angie since you've booked both in the past — fair game as backups?"*
- Cap at 5 — the API rejects more.

For each candidate, confirm the user would actually accept a booking there. Avoid padding the list for the sake of having 5 entries.

### 3. Confirm the date / time window / party

- Date: a single day. If the user wants a range, ask which days to try and create one search per day (each its own call).
- Time window: be generous if the user is flexible ("any time" → 17:00–22:00). Narrow windows catch fewer slots.
- Party: hard constraint.

### 4. Check for existing overlapping searches

Consult the cached `searches list` snapshot. If an existing search already covers this date/window with overlapping restaurants, offer:

> "You already have a search running for Don Angie + Via Carota on May 15. Want me to add Lodi to that existing search, or set a separate one?"

Use `spot searches update --restaurant ...` for the in-place add (full replacement of the target list — include the existing restaurants in the new list).

### 5. Create the search

```
spot searches create \
  --restaurant rst_a,rst_b,rst_c,rst_d,rst_e \
  --date 2026-05-15 \
  --start-time 17:00 \
  --end-time 22:00 \
  --party 4 \
  --json
```

### 6. Set expectations

After creation, narrate what happens next:

> "Search set. Spot will watch for openings through May 15 at 10 PM. When something hits, you'll get a notification in the mobile app and the booking will be complete. You can check status any time with `spot searches list`."

## Edge cases

- **User wants a range of dates:** create one search per date. Confirm each one before firing the commands.
- **User has 5+ candidates already:** narrow with the user ("these two are must-haves; which 3 of the rest are highest priority?").
- **One of the candidates has a platform the user hasn't connected:** flag it — Spot can still watch the restaurant, but the booking will fail. Offer to exclude that restaurant or ask the user to link the platform.
