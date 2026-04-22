# Spot skill for Claude Code

Lets Claude Code drive the Spot CLI to find and book NYC restaurant reservations via conversation.

## Install

1. Install the Spot CLI. See the top-level repo README for options (Homebrew, Scoop, `go install`, `install.sh`).
2. Sign in once:
   ```bash
   spot auth login
   ```
3. Copy this skill into your Claude Code skills directory:
   ```bash
   cp -R skills/spot ~/.claude/skills/
   ```
4. Restart Claude Code. The skill auto-activates when you ask about reservations, restaurants, or dining plans.

## What it does

- **Finds tables.** Ask for a specific restaurant and date, or describe the kind of dinner you want ("Italian, Flatiron, Saturday, 4 people"). The skill narrows candidates, searches for availability, and narrates options.
- **Books on confirmation.** Always asks before locking in a reservation.
- **Sets autobook searches.** If nothing is available now, Spot can watch up to 5 restaurants and grab the first matching slot automatically.
- **State-aware.** Loads your upcoming reservations, active searches, and connected platforms at the start of a session so suggestions are informed.

## Example prompts

- *"Book Gramercy Tavern for 2 at 7 next Friday."*
- *"I need a 4-top in the West Village on Saturday — Italian if possible."*
- *"Nothing open at Tatiana for this Saturday? Watch it for me."*
- *"What do I have coming up?"*
- *"Cancel my Saturday reservation and find something for Sunday."*

## Caveats

- Booking requires you to have linked the relevant platform (Resy, OpenTable, SevenRooms, DoorDash) via the Spot **mobile app**. The CLI can't link platforms — if the skill tells you to open the mobile app, that's why.
- All times and dates are `America/New_York`.
- Slots have a ~5-minute TTL. If the skill tells you a slot expired, it'll re-search and try again.

## Where to file issues

Skill or CLI issues: [spot-nyc/spot](https://github.com/spot-nyc/spot/issues).
