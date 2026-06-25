# RUNBOOK

`RUNBOOK.md` is the operational guide for using and maintaining `stan`. Use `README.md` for the entrypoint and high-level usage, and `SPEC.md` for the formal current-state product contract.

## Build

Install the global CLI:

```bash
./install.sh
```

By default this writes `stan` to `~/bin/stan`. Override the target with:

```bash
./install.sh --bin-dir /path/to/bin
./install.sh --prefix /path/to/prefix
```

Build the local CLI:

```bash
go build -o stan .
```

If the default Go build cache is not writable on macOS:

```bash
GOCACHE=/private/tmp/stan-gocache go build -o stan .
```

Check the built CLI version:

```bash
./stan version
```

Run local diagnostics:

```bash
./stan doctor
./stan doctor --json
```

`doctor` checks the embedded version, credentials location, config directory, and local token state. It does not call Google APIs.

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
6. Save it as Stan's local credentials:

```bash
mkdir -p ~/.config/stan
mv ~/Downloads/client_secret*.json ~/.config/stan/client_secret.json
```

Stan also supports `./client_secret.json` in the repository working directory for local experiments.

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
./stan calendar add --when "10:00" "Meeting"
./stan calendar add --when "2026-03-20T10:00" --duration 30m "Meeting"
./stan calendar add --when "2026-03-20T10:00" --end "2026-03-20T11:00" "Meeting"
```

## Tasks Operations

List task lists:

```bash
./stan tasks lists
```

List the default Stan task list:

```bash
./stan tasks list
./stan tasks list --verbose
```

Show one task with its subtasks and metadata:

```bash
./stan tasks show "Agent47"
./stan tasks show --json "Agent47"
./stan tasks show --list "Stan" "Agent47"
```

Default behavior:

- resolves a task list titled `Stan`
- comparison is case-insensitive
- duplicate case-insensitive matches fail as ambiguous
- `tasks list` shows only top-level tasks by default
- `tasks list --verbose` includes subtasks and follows Google Tasks hierarchy and `position` ordering
- `tasks show` accepts either an exact task ID or an exact case-insensitive task title
- if multiple tasks share the same title, use the task ID

Create tasks:

```bash
./stan tasks add "Buy milk"
./stan tasks add --due "2026-03-25" "Buy milk"
./stan tasks add --notes "semi skimmed" "Buy milk"
./stan tasks add --list "Personal" "Buy milk"
```

## Output Modes

Human-readable default:

```bash
./stan tasks list
```

JSON:

```bash
./stan tasks list --json
./stan tasks list --verbose --json
./stan tasks show --json "Agent47"
./stan calendar list --json
```

Tasks JSON output exposes only `title` and `notes` for each task. `tasks show --json` also includes nested `subtasks`.

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
