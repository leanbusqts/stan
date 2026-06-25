# CHANGELOG

## [Unreleased]

## [0.1.3] - 2026-06-25
### Added
- Added `install.sh` to build and install `stan` as a global command.

### Changed
- Updated source onboarding to use `./install.sh` as the primary path.

## [0.1.2] - 2026-06-12
### Added
- Added `stan version`, `stan doctor`, and explicit root `stan help` usage.

### Changed
- Expanded source checkout onboarding with concrete build, credential placement, and login commands.

## [0.1.1] - 2026-06-12
### Added
- Added `stan tasks show <id-or-title>` to inspect a task with its descendant subtasks and metadata.
- Added `stan tasks list --verbose` to include subtasks in list output.
- Added `RELEASING.md` as the maintainer release checklist.

### Changed
- Changed `stan tasks list` to show only top-level tasks by default.
- Simplified human-readable `stan tasks show` output to only print task notes below each title.
- Simplified Tasks JSON output to expose only `title`, `notes`, and nested `subtasks` for task detail output.
- Updated `AGENTS.md` so release tasks load `RELEASING.md`.

## [0.1.0] - 2026-06-12
### Added
- Added release metadata through `VERSION`.
- Added operational documentation in `RUNBOOK.md`.
- Added current-state repository documentation in `SNAPSHOT.md`.
- Added release history in `CHANGELOG.md`.
- Documented the current product contract in root `SPEC.md`.

### Changed
- Updated `README.md` with current OAuth setup, token storage, command usage, Tasks behavior, repository structure, and contributor checks.
- Changed the default Google Tasks list from `@default` to a case-insensitive task list title match for `Stan`.
- Made `stan tasks add` without `--list` create tasks in the default `Stan` list.
- Preserved Google Tasks hierarchy and sibling ordering in `stan tasks list` by using task `parent` and `position`.
- Indented subtasks in human-readable task output.
- Loaded all task result pages before ordering task output.

### Fixed
- Fixed OAuth PKCE login by passing the PKCE verifier directly to `oauth2.S256ChallengeOption`, preventing Google from rejecting the token exchange after browser authorization.
