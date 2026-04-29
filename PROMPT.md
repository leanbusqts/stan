You are implementing the project **Stan** in this repository.

CRITICAL INSTRUCTIONS:

1. Read `SPEC.md` completely before making any changes.
2. Treat `SPEC.md` as the source of truth.
3. Implement ONLY what is explicitly defined in `SPEC.md`.
4. Do NOT invent features.
5. Do NOT add unnecessary dependencies.
6. Keep architecture minimal, clean, idiomatic Go.
7. Favor Go standard library whenever possible.
8. Use small focused packages and files.
9. The project must build and run on macOS/Linux.
10. Output concise progress summaries while working.

--------------------------------------------------
PRIMARY GOAL
--------------------------------------------------

Implement the MVP described in `SPEC.md`.

Main expected capabilities:

- Google OAuth2 auth flow
- Desktop App credentials
- PKCE (S256)
- Local loopback auth server
- Manual fallback auth mode
- Token persistence
- Token auto-refresh
- Token autosave after refresh
- `stan auth login`
- `stan auth status`
- `stan auth logout`

- `stan calendar list`
- `stan calendar add`

- `stan tasks list`
- `stan tasks lists`
- `stan tasks add`

- Flags:
  - --json
  - -q
  - --no-color
  - date filters where required

--------------------------------------------------
MANDATORY WORKFLOW
--------------------------------------------------

Step 1:
Read and summarize `SPEC.md`.

Step 2:
Inspect current repository state.

Step 3:
Create implementation plan aligned to SPEC.

Step 4:
Implement incrementally in safe commits / logical steps.

Step 5:
Run validation after each major step:
- `go fmt ./...`
- `go vet ./...`
- `go test ./...` (if tests exist)
- `go build ./...`

Step 6:
Fix issues until build passes.

--------------------------------------------------
ARCHITECTURE RULES
--------------------------------------------------

Use clean structure similar to:

/main.go
/auth/
/calendar/
/tasks/
/internal/

If repository already has structure:
- adapt carefully
- do not rewrite unnecessarily

Keep responsibilities separated:

- auth package:
  oauth, browser flow, token refresh, logout

- calendar package:
  list/add events

- tasks package:
  lists/list/add tasks

- internal:
  config, token store, formatting helpers

--------------------------------------------------
TOKEN STORAGE RULES
--------------------------------------------------

Implement per SPEC:

- Prefer keychain if SPEC requires it
- fallback file token store
- atomic writes:
  token.json.tmp -> rename

Never print secrets.

--------------------------------------------------
CLI RULES
--------------------------------------------------

No Cobra unless already required by repo.

Prefer:
- os.Args
- flag.FlagSet

Commands must fail with clear human-readable messages.

--------------------------------------------------
OUTPUT RULES
--------------------------------------------------

Default:
human readable

If `--json`:
strict JSON only

If `-q`:
minimal output

--------------------------------------------------
DATE RULES
--------------------------------------------------

Use formats defined in SPEC.md only.

No extra natural language parsing unless SPEC explicitly requires it.

--------------------------------------------------
WHEN BLOCKED
--------------------------------------------------

If something in SPEC is ambiguous:

1. choose simplest implementation
2. stay minimal
3. note assumption in final summary

--------------------------------------------------
DO NOT
--------------------------------------------------

- Do not overengineer
- Do not add web UI
- Do not add TUI
- Do not add telemetry
- Do not add unrelated features
- Do not refactor unrelated code

--------------------------------------------------
FINAL DELIVERABLE
--------------------------------------------------

When finished provide:

1. What was implemented
2. Files created/changed
3. Commands to run
4. Any assumptions
5. Remaining TODOs strictly from SPEC

--------------------------------------------------
START NOW:

1. Read SPEC.md
2. Inspect repo
3. Implement Stan MVP exactly per spec
```
