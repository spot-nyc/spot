# spot

The Spot SDK and CLI — manage reservations, searches, and restaurant lookup on the [Spot](https://spot.nyc) reservation platform.

## Install

### Homebrew (macOS / Linux)

```bash
brew install spot-nyc/tap/spot
```

### Scoop (Windows)

```bash
scoop bucket add spot-nyc https://github.com/spot-nyc/scoop-bucket
scoop install spot
```

### `go install`

```bash
go install github.com/spot-nyc/spot/cmd/spot@latest
```

### Shell installer

```bash
curl -fsSL https://raw.githubusercontent.com/spot-nyc/spot/main/install.sh | sh
```

Run `spot update` at any time to check how `spot` was installed and see the matching upgrade command.

## Usage

```bash
spot auth login                              # Sign in via browser
spot restaurants search "gramercy"           # Find restaurants
spot restaurants discover "italian" --sort rating --limit 5 # Ranked recommendations
spot reservations search \                   # Check availability
  --restaurant rst_abc \
  --date 2026-05-15 \
  --start-time 18:00 --end-time 21:00 \
  --party 2
spot reservations book <slotId>              # Book a slot
spot searches create ...                     # Autobook when availability drops
spot reservations list                       # Upcoming reservations
spot reservations history                    # Full reservation log
```

Every command supports `--json` for machine-readable output. Restaurant discovery sorts are `auto`, `relevance`, `rating`, and `distance`. Run `spot <command> --help` for flags and details.

## Using Spot with AI agents

The repo ships agent instructions at [`skills/spot/`](skills/spot/README.md) that let an AI agent drive the CLI conversationally — find tables, plan dinners, set autobook searches, check state, handle cancellations.

Agents need shell access to run `spot` commands and should parse CLI responses with `--json`. Claude Code can install the skill directly:

```bash
cp -R skills/spot ~/.claude/skills/
```

Other agents can load [`skills/spot/SKILL.md`](skills/spot/SKILL.md) plus the relevant recipe from [`skills/spot/recipes/`](skills/spot/recipes/). See [`skills/spot/README.md`](skills/spot/README.md) for the runtime contract and example prompts.

## License

MIT — see [LICENSE](LICENSE).
