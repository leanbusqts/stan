# Stan

Stan is a local-first Go CLI for Google Calendar and Google Tasks.

## Install

```bash
go build -o stan .
```

## Google Desktop OAuth credentials

1. Open https://console.cloud.google.com/
2. Create or select a project.
3. Enable Google Calendar API and Google Tasks API.
4. Create OAuth client credentials of type `Desktop app`.
5. Download the JSON file and name it `client_secret.json`.

## Credentials location

Stan looks for `client_secret.json` in:

- the current directory
- `~/.config/stan/`

## Token storage

Stan stores OAuth tokens in this priority order:

1. OS keychain
2. `~/.config/stan/token.json`

The file fallback uses atomic writes and `0600` permissions.

## Commands

```bash
stan auth login
stan auth status
stan auth logout

stan calendar list
stan calendar list --days 7
stan calendar list --start 2026-03-20 --end 2026-03-25
stan calendar add "Meeting" --when "10:00"
stan calendar add "Meeting" --when "2026-03-20T10:00" --duration 30m

stan tasks list
stan tasks list --json
stan tasks lists
stan tasks add "Buy milk"
stan tasks add "Buy milk" --list "Personal"
stan tasks add "Buy milk" --due "2026-03-25"
stan tasks add "Buy milk" --notes "semi skimmed"
```

## Global flags

```bash
--json
-q
--no-color
```

## Security notes

- Never commit `client_secret.json`.
- Stan never prints OAuth secrets.
- Refresh tokens are checksummed on save/load.
- If keychain storage is unavailable, Stan falls back to the local token file.
