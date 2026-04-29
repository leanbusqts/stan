# Stan CLI — Implementation Spec

## Review Workflow (IMPORTANT)

This specification has been:
- Reviewed by Grok and Gemini
- Iterated across multiple versions
- Refined for robustness, UX, and implementation clarity

This version includes:
- Core MVP features
- Robust OAuth handling
- Advanced security (keychain + checksum)
- CLI UX improvements (readable + scriptable output)

It is considered **final and ready for implementation**.

---

# 1. Project Overview

## Name
Stan

## Description
Stan is a local-first CLI tool that allows users to interact with Google Calendar and Google Tasks directly from the terminal.

---

# 2. Technical Stack

## Language
Go (>= 1.21)

## Constraints
- No CLI frameworks
- Standard library preferred
- Minimal dependencies

## Allowed Libraries
- golang.org/x/oauth2
- golang.org/x/oauth2/google
- google.golang.org/api/calendar/v3
- google.golang.org/api/tasks/v1

## Optional (Security)
- github.com/zalando/go-keyring

---

# 3. Features (MVP)

## Authentication
- OAuth2 Authorization Code Flow
- Desktop App credentials
- PKCE (S256)
- Local server + manual fallback
- Token persistence
- Automatic refresh
- Auto-save on refresh
- Revoked token detection (`invalid_grant`)
- `stan auth status`
- `stan auth logout`

---

## Calendar
- List events (default: next 7 days)
- Create event
- Filtering: --days, --start, --end
- Duration: --duration

---

## Tasks
- List tasks
- List task lists
- Create task
- Multi-list support (--list)
- Due date (--due)
- Notes (--notes)

---

# 4. CLI Interface

```

stan auth login
stan auth status
stan auth logout

stan calendar list --days 7
stan calendar list --start 2026-03-20 --end 2026-03-25
stan calendar add "Meeting" --when "10:00" --duration 30m

stan tasks list
stan tasks lists
stan tasks add "Buy milk" --list "Personal" --due "2026-03-25"

# Global flags

--json
-q
--no-color

```

---

# 5. CLI Parsing Strategy

- os.Args → command routing
- flag package → per-command FlagSet

---

# 6. OAuth2 Authentication

- Desktop App credentials (mandatory)
- Scopes:
  - calendar.events
  - tasks
- Redirect:
  - http://127.0.0.1:{port}
- Dynamic port (port 0)
- PKCE enabled
- Fallback manual auth (paste URL)

---

# 7. Token Management

## Storage priority
1. Keychain (if available)
2. File fallback

## File path
~/.config/stan/token.json

## Features
- Atomic write (tmp + rename)
- Schema versioning
- Email stored
- SHA256 checksum (refresh_token)
- Auto-refresh + auto-save

---

# 8. Calendar Integration

## Default behavior
- Range: now → +7 days
- Calendar: primary

## Add event
- title
- when
- duration OR end
- Validation: end > start

---

# 9. Tasks Integration

## Features
- list tasks
- list lists
- add task
- multi-list support

---

# 10. Date Handling

Supported:
- RFC3339
- date-only
- time-only

Implementation:
- time.ParseInLocation

---

# 11. Output System (UI / Visualization)

## Philosophy

Stan output must be:

- Human-readable by default
- Machine-readable when needed
- Predictable and consistent
- Minimal but expressive

---

## Output Modes

### 1. Default (Human-readable)

- Grouped logically (e.g. by day)
- Aligned columns
- Icons for context
- Optional color

---

### 2. JSON Mode

```

--json

```

- Raw structured output
- No formatting
- Designed for piping (`jq`, scripts)

---

### 3. Quiet Mode

```

-q

```

- Minimal output
- Only essential values (titles or IDs)

---

## Global Flags

- --json
- -q
- --no-color

---

## Calendar Output Format

Example:

```

📅 Today
10:00  Meeting
14:30  Call

📅 Tomorrow
09:00  Gym

📅 Next Days
Fri 11:00 Dentist

```

---

## Tasks Output Format

Example:

```

Tasks (@default)

[ ] Buy milk
[ ] Call mom
[x] Pay bills

```

---

## Visual Enhancements

### Icons

- 📅 → calendar
- [ ] → pending task
- [x] → completed task
- ⚠️ → near due
- 🔥 → urgent
- ⏳ → upcoming

---

### ANSI Colors

(No external libraries)

Use:

- Red → overdue
- Green → today / completed
- Yellow → upcoming

Disable via:
```

--no-color

```

---

## Formatting Rules

- Fixed spacing for columns
- Consistent indentation
- No random line breaks
- Deterministic output order

---

# 12. Error Handling

- Clear, human-readable messages
- No stack traces

Special case:

```

Session expired or revoked.
Run: stan auth login

```

---

# 13. UX Enhancements

- Headless detection → skip browser
- Manual auth fallback
- Onboarding instructions
- Direct Google Cloud Console link

---

# 14. Robustness

- Silent retry (1–2 attempts)
- Simple exponential backoff
- Auth timeout (~30s)

---

# 15. Definition of Done

- Auth works (both modes)
- Token stored securely
- Auto-refresh works
- Calendar + Tasks fully usable
- Output readable + JSON compatible
- No crashes in normal usage

---

# FINAL STATUS

- Fully reviewed
- Hardened
- UX-aware
- Security-aware
- Ready for implementation
