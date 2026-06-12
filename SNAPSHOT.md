# SNAPSHOT

## 1. Project Overview

- **Name:** `stan`
- **Purpose:** local-first Go CLI for Google Calendar and Google Tasks
- **Command:** `stan`

## 2. Current Status

- **Runtime:** single Go CLI entrypoint in `main.go`
- **Parsing:** standard library `flag` with command-specific `FlagSet` routing
- **Authentication:** OAuth Authorization Code Flow with local callback, PKCE S256, browser launch, and manual redirect fallback
- **Token storage:** keychain-first storage with `~/.config/stan/token.json` fallback
- **Token refresh:** OAuth clients auto-refresh and save updated tokens
- **Calendar:** supports event listing from primary calendar and event creation
- **Tasks:** supports task-list discovery, task listing, task detail lookup, and task creation
- **Default task list:** resolves a Google Tasks list titled `Stan`, case-insensitively
- **Task listing:** default list output shows top-level tasks only; `--verbose` includes subtasks in Google Tasks hierarchy and sibling order
- **Output:** human-readable default plus JSON, quiet, and no-color modes
- **Version:** `0.1.0`

## 3. Current Commands

- `stan auth login`
- `stan auth status`
- `stan auth logout`
- `stan calendar list [--days N] [--start YYYY-MM-DD] [--end YYYY-MM-DD]`
- `stan calendar add --when "10:00" [--duration 30m] [--end 2026-03-20T10:30] "Meeting"`
- `stan tasks list [--verbose]`
- `stan tasks lists`
- `stan tasks show [--list Stan] "Agent47"`
- `stan tasks add [--list Personal] [--due 2026-03-25] [--notes "..."] "Buy milk"`

Global flags:

- `--json`
- `-q`
- `--no-color`

## 4. Key Repository Structure

- `auth/` - OAuth login, status, logout, and authorized HTTP client behavior
- `calendar/` - Google Calendar integration
- `internal/` - local config, output helpers, keychain integration, token persistence
- `tasks/` - Google Tasks integration, default list resolution, task ordering
- `rules/` - security rules for agents
- `specs/` - optional task-scoped planning artifact
- `main.go` - command routing and human output rendering
- `README.md` - entrypoint and high-level guide
- `RUNBOOK.md` - operational guide
- `SPEC.md` - current-state product contract
- `CHANGELOG.md` - release history
- `VERSION` - current release version

## 5. Constraints And Risks

- Google Desktop OAuth credentials must be provided locally.
- The default Tasks workflow requires a Google Tasks list named `Stan`.
- Duplicate case-insensitive task-list names such as `Stan` and `stan` are ambiguous.
- Keychain behavior is platform-dependent; file fallback is available.
- Network calls depend on Google API availability and granted scopes.
- The repo currently has no Makefile; contributor checks are direct Go commands.

## 6. Last Updated

- June 12, 2026

## 7. Verification Notes

- `GOCACHE=/private/tmp/stan-gocache go test ./...` passed on June 12, 2026.
- `GOCACHE=/private/tmp/stan-gocache go build -o stan .` passed on June 12, 2026.
- Manual auth regression target: `./stan auth login` should complete after browser consent and `./stan auth status` should show a logged-in user.
- Manual tasks regression target: `./stan tasks list` should print only top-level `Tasks (Stan)`, while `./stan tasks list --verbose` should preserve the visual order and indentation shown in Google Tasks.
