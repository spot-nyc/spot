# Recipe: find and book

**Use when:** the user wants a specific restaurant, or a specific date + party with flexibility on restaurant.

## Flow

### 1. Understand the ask

Resolve any ambiguity **before** running commands:
- "This Friday" → compute absolute date, confirm with user.
- "Dinner" → propose a default window (6:00 PM – 9:00 PM) and confirm.
- Party size must be explicit.

### 2. Resolve the restaurant ID

If the user named a restaurant, resolve the ID:

```
spot restaurants search "Gramercy Tavern" --json
```

Pick the top match by name. If the top result is a close-but-not-exact name match, call out the ambiguity and confirm with the user before proceeding.

### 3. Search for available slots

```
spot reservations search \
  --restaurant rst_abc \
  --date 2026-05-15 \
  --start-time 18:00 \
  --end-time 21:00 \
  --party 2 \
  --json
```

Parse the response. Each entry has `id`, `date`, `time`, `seating`, `platform`, `restaurant.name`.

### 4a. Slots exist → narrate + confirm + book

Pick the slot closest to the user's stated time preference. Present in plain language:

> "Gramercy Tavern has a Dining Room slot on Resy at 7:00 PM for 2 people — closest to your 7 PM target. Book it?"

On confirmation:

```
spot reservations book slot_xyz --json
```

Read the resulting reservation ID back to the user. If the command returns exit code 11 (slot expired), re-run step 3 and retry once. If exit code 10 (platform not connected), stop and direct the user to the mobile app for that specific platform.

### 4b. No slots → offer an autobook search

> "Nothing available at Gramercy Tavern in that window. I can set up a search that will book the first matching table the moment one drops. We can add up to 4 other restaurants as fallbacks — want to add any?"

Ask for fallbacks. Then:

```
spot searches create \
  --restaurant rst_abc,rst_def,rst_ghi \
  --date 2026-05-15 \
  --start-time 18:00 \
  --end-time 21:00 \
  --party 2 \
  --json
```

Announce: *"Search created. I'll autobook the first slot that opens at any of these spots on May 15 between 6 and 9 PM."* Then remind the user they'll be notified via the Spot mobile app when a booking lands.

## Edge cases

- **Already holding a reservation for that date/time range:** check the cached `reservations list` snapshot. If the user already has a booking that conflicts, ask whether they want to cancel it first or proceed in parallel.
- **Existing search already covers this intent:** if `searches list` shows a matching search, offer `spot searches update` instead of creating a duplicate.
- **Platform not connected for the target restaurant:** before searching, check the opening-routine `whoami.connectedPlatforms` against the restaurant's `platforms`. If the user has zero of the required platforms connected, stop early and direct them to the mobile app.
