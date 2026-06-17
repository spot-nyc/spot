# Spot agent skill

Lets an AI agent drive the Spot CLI to find and book NYC restaurant reservations via conversation.

## Requirements

- The Spot CLI is installed. See the top-level repo README for Homebrew, Scoop, `go install`, and `install.sh` options.
- The user is signed in once with:
  ```bash
  spot auth login
  ```
- The agent can run shell commands and read stdout/stderr. Agents without command execution can use these instructions as planning guidance, but cannot search, book, cancel, or create searches directly.

## Claude Code install

Copy this skill into your Claude Code skills directory:

```bash
cp -R skills/spot ~/.claude/skills/
```

Restart Claude Code. The skill auto-activates when you ask about reservations, restaurants, or dining plans.

## Generic agent usage

Load [SKILL.md](SKILL.md) as the primary instructions. Load the recipe that matches the user's intent from [recipes/](recipes/) before acting. The agent should:

- Use `--json` for CLI output it needs to parse.
- Confirm before `spot reservations book`, `spot reservations cancel`, or `spot searches delete`.
- Resolve relative dates to absolute `YYYY-MM-DD` dates in `America/New_York`.
- Never ask for or print tokens; use `spot auth login` for authentication.
- Send platform-linking issues to the Spot mobile app.

## What it does

- **Finds tables.** Ask for a specific restaurant and date, or describe the kind of dinner you want ("Italian, Flatiron, Saturday, 4 people"). The skill uses restaurant discovery, ratings, editorial mentions, and dish notes to narrow candidates, searches for availability, and narrates options.
- **Books on confirmation.** Always asks before locking in a reservation.
- **Sets autobook searches.** If nothing is available now, Spot can watch up to 5 restaurants and grab the first matching slot automatically.
- **State-aware.** Loads your upcoming reservations, active searches, and connected platforms at the start of a session so suggestions are informed.

## Example prompts

- *"Book Gramercy Tavern for 2 at 7 next Friday."*
- *"I need a 4-top in the West Village on Saturday — Italian if possible."*
- *"Find the highest-rated Italian places downtown and check tables."*
- *"Where should I go for dumplings, and what should I order?"*
- *"Nothing open at Tatiana for this Saturday? Watch it for me."*
- *"What do I have coming up?"*
- *"Cancel my Saturday reservation and find something for Sunday."*

## Caveats

- Booking requires you to have linked the relevant platform (Resy, OpenTable, SevenRooms, DoorDash) via the Spot **mobile app**. The CLI can't link platforms — if the skill tells you to open the mobile app, that's why.
- All times and dates are `America/New_York`.
- Slots have a ~5-minute TTL. If the skill tells you a slot expired, it'll re-search and try again.

## Where to file issues

Skill or CLI issues: [spot-nyc/spot](https://github.com/spot-nyc/spot/issues).
