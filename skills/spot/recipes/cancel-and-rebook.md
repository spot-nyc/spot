# Recipe: cancel and rebook

**Use when:** the user wants to modify an existing reservation — different time, different restaurant, different party size, or straight cancellation with a replacement.

## Flow

### 1. Identify the reservation

If the user doesn't say which reservation, show their upcoming list (from opening-routine cache or fresh `spot reservations list --json`) and ask which one.

### 2. Understand the replacement intent

Ask before cancelling:

- *"Do you want me to just cancel, or cancel and book something else?"*
- If "and book something else", capture the new intent (restaurant / date / time / party) before touching the existing reservation. Failing to book the replacement after cancelling leaves the user with nothing.

### 3. Confirm cancellation

Full canonical phrasing:

> "About to cancel your reservation at Gramercy Tavern on May 15 at 7:00 PM for 2 people (Resy). Proceed?"

Include the platform — some platforms charge cancellation fees, and users may want to reconsider.

### 4. If the user wants a replacement, test availability FIRST

Before cancelling, run `spot reservations search` for the new intent. If slots exist, note the candidate but don't book yet. If no slots, pivot to offering an autobook search for the new intent.

If the replacement intent is vague ("find me something for Sunday"), fall back to the plan-dinner flow and use reservation history as a preference prior (see SKILL.md § *Using history as soft preference signal*). Deprioritize the restaurant they just cancelled — if they're pivoting away from it, they probably want something different rather than a shifted time at the same spot, unless they say otherwise.

### 5. Cancel, then book

Only after confirming the replacement is feasible:

```
spot reservations cancel rsv_abc --json
```

Then either `spot reservations book <slotId>` for the candidate slot, or `spot searches create ...` for the autobook path.

### 6. Recap

Tell the user exactly what happened:

> "Cancelled Gramercy Tavern (May 15, 7 PM). Booked Shuko (May 15, 8:30 PM, Dining Room). You're set."

Or, if only cancelling:

> "Cancelled Gramercy Tavern (May 15, 7 PM). No replacement requested."

## Edge cases

- **Cancellation fails:** unusual. Surface the server message. Do NOT book a replacement until the cancel succeeds; the user would end up double-booked.
- **Replacement search finds nothing:** ask whether to cancel anyway (if the old slot is wrong enough) or keep the existing booking while setting an autobook search for a better option.
- **User changes mind mid-flow:** always possible to back out before the cancel command runs. Confirm again if any input has changed since the initial ask.
