# RUNBOOK

`RUNBOOK.md` is the operational guide for using and maintaining `stan`. Use `README.md` for the entrypoint and high-level usage, and `SPEC.md` for the formal current-state product contract.

## Build

Build the local CLI:

```bash
go build -o stan .
```

If the default Go build cache is not writable on macOS:

```bash
GOCACHE=/private/tmp/stan-gocache go build -o stan .
```

## Test

Run the full suite:

```bash
go test ./...
```

Cache workaround:

```bash
GOCACHE=/private/tmp/stan-gocache go test ./...
```

## First-Time OAuth Setup

1. Open Google Cloud Console.
2. Create or select a project.
3. Enable Google Calendar API and Google Tasks API.
4. Create OAuth client credentials of type `Desktop app`.
5. Download the credentials JSON.
6. Save it as either:
   - `./client_secret.json`
   - `~/.config/stan/client_secret.json`

Do not commit `client_secret.json`.

## Login

```bash
./stan auth login
```

Expected behavior:

- Stan prints a Google OAuth URL.
- Stan tries to open the browser.
- Google prompts for consent.
- The local callback server receives the OAuth code.
- Stan exchanges the code and stores the token.

Manual fallback:

- Press Enter at the prompt.
- Paste the full redirect URL from the browser.

Verify:

```bash
./stan auth status
```

`Token expires in: 1.0 hours` refers to the short-lived access token, not necessarily the whole session. Stan refreshes access tokens with the stored refresh token.

## Logout

```bash
./stan auth logout
```

Expected behavior:

- remote token revocation is attempted best-effort
- local keychain and file-backed token state are removed

## Calendar Operations

List events:

```bash
./stan calendar list
./stan calendar list --days 14
./stan calendar list --start 2026-03-20 --end 2026-03-25
```

Create events:

```bash
./stan calendar add "Meeting" --when "10:00"
./stan calendar add "Meeting" --when "2026-03-20T10:00" --duration 30m
./stan calendar add "Meeting" --when "2026-03-20T10:00" --end "2026-03-20T11:00"
```

## Tasks Operations

List task lists:

```bash
./stan tasks lists
```

List the default Stan task list:

```bash
./stan tasks list
```

Default behavior:

- resolves a task list titled `Stan`
- comparison is case-insensitive
- duplicate case-insensitive matches fail as ambiguous
- output follows Google Tasks hierarchy and `position` ordering

Create tasks:

```bash
./stan tasks add "Buy milk"
./stan tasks add "Buy milk" --due "2026-03-25"
./stan tasks add "Buy milk" --notes "semi skimmed"
./stan tasks add "Buy milk" --list "Personal"
```

## Output Modes

Human-readable default:

```bash
./stan tasks list
```

JSON:

```bash
./stan tasks list --json
./stan calendar list --json
```

Quiet:

```bash
./stan tasks list -q
```

Disable color:

```bash
./stan tasks list --no-color
```

## Troubleshooting

### Browser says authenticated but CLI says session expired

Rebuild from the current checkout and retry:

```bash
GOCACHE=/private/tmp/stan-gocache go build -o stan .
./stan auth login
```

This release fixed the PKCE verifier/challenge exchange. An older binary can still show the old behavior.

### `task list "Stan" not found`

Create a Google Tasks list named `Stan` in Google Tasks. Case does not matter.

### Multiple Stan task lists

If Google Tasks has both `Stan` and `stan`, Stan refuses to choose one. Rename or remove duplicates.

### Go test fails with build-cache permission errors

Use a writable cache:

```bash
GOCACHE=/private/tmp/stan-gocache go test ./...
```

### Auth state is stale

Try:

```bash
./stan auth logout
./stan auth login
```

## Release Checklist

1. Update `README.md`, `RUNBOOK.md`, `SNAPSHOT.md`, `SPEC.md`, `CHANGELOG.md`, and `VERSION`.
2. Run:

```bash
GOCACHE=/private/tmp/stan-gocache go test ./...
GOCACHE=/private/tmp/stan-gocache go build -o stan .
```

3. Review the staged diff.
4. Commit.
5. Tag the release as `vX.Y.Z`.
6. Push the branch and tag.
