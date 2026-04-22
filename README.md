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

Every command supports `--json` for machine-readable output. Run `spot <command> --help` for flags and details.

## Using Spot with Claude Code

The repo ships a Claude Code skill at [`skills/spot/`](skills/spot/README.md) that lets Claude Code drive the CLI conversationally — find tables, plan dinners, set autobook searches, check state, handle cancellations.

```bash
cp -R skills/spot ~/.claude/skills/
```

Restart Claude Code. The skill auto-activates when you ask about reservations, restaurants, or dining plans. See [`skills/spot/README.md`](skills/spot/README.md) for example prompts.

## License

MIT — see [LICENSE](LICENSE).
