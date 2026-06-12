# Stan

`stan` is a local-first Go CLI for Google Calendar and Google Tasks.

It focuses on a small, scriptable command surface:

- OAuth login for a Google Desktop OAuth client
- Calendar event listing and creation
- Google Tasks list discovery, task listing, and task creation
- human-readable output by default
- JSON and quiet output for scripts

## Getting Started From Source

Clone and build from this repository:

```bash
git clone https://github.com/leanbusqts/stan.git
cd stan
go build -o stan .
./stan help
```

Create Google OAuth credentials:

1. Open Google Cloud Console.
2. Create or select a project.
3. Enable Google Calendar API and Google Tasks API.
4. Create OAuth client credentials of type `Desktop app`.
5. Download the JSON file and save it as Stan's local credentials:

```bash
mkdir -p ~/.config/stan
mv ~/Downloads/client_secret*.json ~/.config/stan/client_secret.json
```

Then log in:

```bash
./stan auth login
./stan auth status
```

Check local setup:

```bash
./stan version
./stan doctor
```

Stan also supports keeping credentials in the repository working directory for local experiments:

```text
./client_secret.json
```

Do not commit `client_secret.json`.

List tasks from the default Stan task list:

```bash
./stan tasks list
```

List upcoming calendar events:

```bash
./stan calendar list
```

## Quickstart For Existing Checkouts

If you already have the repo locally:

```bash
go build -o stan .
./stan help
./stan doctor
```

## Google OAuth Setup

1. Open Google Cloud Console.
2. Create or select a project.
3. Enable Google Calendar API and Google Tasks API.
4. Create OAuth client credentials of type `Desktop app`.
5. Download the JSON file and name it `client_secret.json`.

Stan requests these scopes:

- `https://www.googleapis.com/auth/calendar.events`
- `https://www.googleapis.com/auth/tasks`
- `openid`
- `email`

## Token Storage

Stan stores OAuth tokens in this priority order:

1. OS keychain
2. `~/.config/stan/token.json`

The file fallback uses atomic writes and `0600` permissions. Refresh tokens are checksummed on save/load. Stan never prints OAuth secrets.

`auth status` reports the current access-token expiry. Google access tokens commonly expire in about one hour; Stan uses the stored refresh token to renew them automatically.

## Public Commands

```bash
stan version
stan doctor
stan help

stan auth login
stan auth status
stan auth logout

stan calendar list
stan calendar list --days 7
stan calendar list --start 2026-03-20 --end 2026-03-25
stan calendar add --when "10:00" "Meeting"
stan calendar add --when "2026-03-20T10:00" --duration 30m "Meeting"
stan calendar add --when "2026-03-20T10:00" --end "2026-03-20T11:00" "Meeting"

stan tasks list
stan tasks list --verbose
stan tasks lists
stan tasks show "Agent47"
stan tasks show --list "Stan" "Agent47"
stan tasks add "Buy milk"
stan tasks add --list "Personal" "Buy milk"
stan tasks add --due "2026-03-25" "Buy milk"
stan tasks add --notes "semi skimmed" "Buy milk"
```

Global flags:

```bash
--json
-q
--no-color
```

`--json` and `-q` are mutually exclusive.

## Tasks Behavior

`stan tasks list` uses the Google Tasks list named `Stan` by default. The match is case-insensitive, so `Stan`, `stan`, and `STAN` all match.

If multiple task lists differ only by case, Stan returns an ambiguity error instead of choosing one at random.

`stan tasks list` shows only top-level tasks by default.

Use `stan tasks list --verbose` to include subtasks. Verbose task output preserves Google Tasks hierarchy and ordering:

- top-level tasks are sorted by Google Tasks `position`
- subtasks are printed directly below their parent
- subtasks are indented in human-readable output
- all API result pages are loaded before sorting

Use `stan tasks show <id-or-title>` to inspect one task together with its subtasks. Human-readable output includes notes when present. Tasks JSON output includes only `title`, `notes`, and nested `subtasks` for `show`.

## Calendar Behavior

`stan calendar list` reads events from the primary calendar.

Default range:

- start: now
- end: now plus 7 days

Supported date/time input:

- RFC3339 timestamps
- date-only values
- time-only values for today

## Repository Structure

```text
stan/
|
+-- auth/              OAuth login, status, logout, token refresh client
+-- calendar/          Google Calendar integration and date parsing
+-- internal/          config, output formatting, keychain, token store
+-- tasks/             Google Tasks integration
+-- rules/             repository security rules for agents
+-- specs/             optional task-scoped planning area
+-- main.go            CLI routing and output
+-- README.md
+-- RUNBOOK.md
+-- SNAPSHOT.md
+-- SPEC.md
+-- VERSION
```

## Contributor Commands

```bash
go test ./...
go build -o stan .
```

On machines where the default Go build cache is not writable, use a local or temporary cache:

```bash
GOCACHE=/private/tmp/stan-gocache go test ./...
GOCACHE=/private/tmp/stan-gocache go build -o stan .
```

## Documentation

- `README.md` is the entrypoint and high-level usage guide.
- `RUNBOOK.md` is the operational guide for setup, verification, and troubleshooting.
- `SNAPSHOT.md` summarizes the current repository state.
- `SPEC.md` is the current-state product contract.
- `CHANGELOG.md` records release history.
- `VERSION` contains the current release version.
