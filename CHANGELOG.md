# CHANGELOG

## [Unreleased]

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
