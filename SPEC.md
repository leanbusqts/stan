# SPEC

This root `SPEC.md` describes the current product contract and technical behavior of `stan` itself. It is not the default place to draft a new feature spec or implementation plan; use `specs/spec.yml` for task-scoped planning work.

## 1. Project Identity

- **Name:** `stan`
- **Type:** local-first Go CLI
- **Public command:** `stan`
- **Primary purpose:** interact with Google Calendar and Google Tasks from a terminal with secure local OAuth storage and scriptable output

## 2. Product Summary

Stan provides:

- embedded version reporting
- local diagnostic checks
- Google OAuth login for Desktop App credentials
- secure token persistence through keychain-first storage with file fallback
- Google Calendar event listing and event creation
- Google Tasks list discovery, task listing, and task creation
- human-readable output by default
- JSON and quiet output modes for scripts

The product is intentionally small:

- no CLI framework
- standard library routing and `flag` parsing
- minimal Google API dependencies
- local-only credential and token state

## 3. Goals

- Keep Calendar and Tasks workflows usable from the terminal.
- Keep OAuth tokens out of logs and source files.
- Prefer OS keychain storage and fall back to a permission-restricted token file.
- Keep default output readable while preserving script-friendly modes.
- Preserve Google Tasks ordering and hierarchy in task output.
- Keep the command surface explicit and easy to test.

## 4. Non-Goals

- Full Google Workspace replacement.
- Background sync or daemon behavior.
- Cloud-hosted relay services.
- Multi-account profile management.
- Vendor-specific agent configuration.
- Project generation or task planning beyond repository documentation.

## 5. Core Workflows

### 5.1 Build

Contributor command:

```bash
go build -o stan .
```

Expected outcome:

- produce a local `stan` executable from the current checkout

### 5.2 Utility commands

Public commands:

```bash
stan version
stan doctor
stan help
```

Expected behavior:

- `version` prints the embedded release version from `VERSION`
- `doctor` reports version, credentials state, config directory state, and local token state without calling Google APIs
- `help` prints the root command surface

### 5.3 Authenticate

Public commands:

```bash
stan auth login
stan auth status
stan auth logout
```

Expected behavior:

- load Google Desktop OAuth credentials from `./client_secret.json` or `~/.config/stan/client_secret.json`
- start a local callback listener on `127.0.0.1` with a dynamic port
- use OAuth Authorization Code Flow with PKCE S256
- open the browser when possible
- support manual redirect URL paste fallback
- store tokens after a successful code exchange
- report login state and access-token expiry through `auth status`
- revoke the current token best-effort and remove local token state on logout

OAuth scopes:

- `https://www.googleapis.com/auth/calendar.events`
- `https://www.googleapis.com/auth/tasks`
- `openid`
- `email`

### 5.4 Token Storage

Storage priority:

1. OS keychain
2. `~/.config/stan/token.json`

File fallback contract:

- create `~/.config/stan` with `0700`
- write token files with `0600`
- write atomically through temporary file plus rename
- persist schema version, email, token, and refresh-token checksum

Refresh behavior:

- API clients use an OAuth token source
- refreshed access tokens are saved automatically
- missing refresh tokens retain the previously stored refresh token when possible
- `invalid_grant`, `400`, and `401` auth failures map to the user-facing expired-session message

### 5.5 Calendar

Public commands:

```bash
stan calendar list [--days N] [--start YYYY-MM-DD] [--end YYYY-MM-DD]
stan calendar add --when "10:00" [--duration 30m] [--end 2026-03-20T10:30] "Meeting"
```

Expected behavior:

- list events from the primary calendar
- default list range is now through seven days ahead
- support explicit `--start` and `--end`
- create events with a title, start, and either duration or explicit end
- reject invalid ranges where end is not after start

Supported date/time input:

- RFC3339
- date-only
- time-only interpreted in the local timezone

### 5.6 Tasks

Public commands:

```bash
stan tasks list [--verbose]
stan tasks lists
stan tasks show [--list Stan] "Agent47"
stan tasks add [--list Personal] [--due 2026-03-25] [--notes "..."] "Buy milk"
```

Default list behavior:

- `tasks list` uses a Google Tasks list titled `Stan`
- `tasks add` without `--list` also uses `Stan`
- matching is case-insensitive
- duplicate case-insensitive matches are rejected as ambiguous

Task listing behavior:

- load all available result pages
- preserve Google Tasks sibling order through the `position` field
- preserve hierarchy through the `parent` field
- `tasks list` returns only top-level tasks by default
- `tasks list --verbose` returns subtasks under their parent with indentation in human-readable output
- treat tasks with missing parents as top-level tasks

Task detail behavior:

- `tasks show <id-or-title>` returns one task plus all descendant subtasks
- lookup prefers exact task ID matches before title matches
- title matching is exact and case-insensitive
- duplicate title matches are rejected as ambiguous
- human-readable output includes notes when present
- JSON output exposes only `title` and `notes` for each task and uses nested `subtasks` for descendants

Task creation behavior:

- create in the default `Stan` list unless `--list` is supplied
- support optional due date
- support optional notes

## 6. Public Command Surface

Supported user-facing commands:

- `stan version`
- `stan doctor`
- `stan help`
- `stan auth login`
- `stan auth status`
- `stan auth logout`
- `stan calendar list`
- `stan calendar add`
- `stan tasks list`
- `stan tasks lists`
- `stan tasks show`
- `stan tasks add`

Supported global flags:

- `--json`
- `-q`
- `--no-color`

Command constraints:

- `--json` and `-q` are mutually exclusive
- unknown commands fail with a clear error
- command-specific positional arguments are validated explicitly

## 7. Output Contract

Default output:

- concise human-readable text
- Calendar events grouped by day label
- Tasks headed by `Tasks (Stan)`
- task completion state shown as `[ ]` or `[x]`
- subtasks indented by hierarchy depth
- optional ANSI colors for calendar/task urgency and status

JSON output:

- structured command results
- no decorative formatting
- suitable for piping to tools such as `jq`

Quiet output:

- compact essential values
- one item per line where applicable

## 8. Error Contract

Auth-expired message:

```text
Session expired or revoked.
Run: stan auth login
```

Other errors should be:

- direct
- user-readable
- free of stack traces
- free of OAuth secrets, access tokens, refresh tokens, and credential payloads

## 9. Repository Structure Contract

- `auth/` - OAuth login, status, logout, token refresh client
- `calendar/` - Google Calendar integration and calendar option parsing
- `internal/` - config, output formatting, keychain integration, token store
- `tasks/` - Google Tasks integration and task ordering
- `rules/` - repository security rules for agents
- `specs/` - optional task-scoped planning area
- `main.go` - CLI routing and human output
- `README.md` - entrypoint and usage guide
- `RUNBOOK.md` - operational guide
- `SNAPSHOT.md` - current repository snapshot
- `CHANGELOG.md` - release history
- `VERSION` - current release version

## 10. Security Contract

- Never commit `client_secret.json`.
- Never print OAuth access tokens or refresh tokens.
- Never hardcode credentials.
- Keep refresh tokens in keychain when available.
- Use strict local file permissions for token fallback.
- Validate untrusted command input before use.
- Avoid unsafe shell command construction.

## 11. Verification Contract

Primary checks:

```bash
go test ./...
go build -o stan .
```

Known local workaround:

```bash
GOCACHE=/private/tmp/stan-gocache go test ./...
GOCACHE=/private/tmp/stan-gocache go build -o stan .
```

## 12. Current Release

- Version: `0.1.1`
- Release date: 2026-06-12
