# RELEASING

`RELEASING.md` is the maintainer checklist for preparing and publishing a Stan release.

When the user asks to generate, publish, or cut a release, follow this checklist end to end: update release docs and version files, run verification, create the release commit, create the annotated tag, and push both the branch and tag unless the user says not to push.

Use:

- `README.md` for user-facing usage
- `RUNBOOK.md` for operational setup and troubleshooting
- `SPEC.md` for the current product contract
- `CHANGELOG.md` for release history
- `VERSION` for the current release number

## Checklist

1. Choose the next semantic version.
2. Update `VERSION`.
3. Move relevant `CHANGELOG.md` entries from `[Unreleased]` into the new version section.
4. Update docs if behavior changed:
   - `README.md`
   - `RUNBOOK.md`
   - `SNAPSHOT.md`
   - `SPEC.md`
   - `RELEASING.md` if the release process changed
5. Run verification:

```bash
GOCACHE=/private/tmp/stan-gocache go test ./...
GOCACHE=/private/tmp/stan-gocache go build -o stan .
git diff --check
```

6. Review the staged diff.
7. Commit the release.
8. Tag the release as `vX.Y.Z`.
9. Push the branch and tag.

## Commands

Example for version `0.1.1`:

```bash
git add README.md RUNBOOK.md SNAPSHOT.md SPEC.md CHANGELOG.md VERSION
git add auth calendar internal tasks main.go
git commit -m "Release stan 0.1.1"
git tag -a v0.1.1 -m "stan 0.1.1"
git push origin main v0.1.1
```

Only stage files that belong to the release. Leave unrelated local work out of the release commit.
