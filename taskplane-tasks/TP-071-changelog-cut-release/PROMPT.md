# TP-071 — Cut a tagged release before v0.5 internal-beta (CHANGELOG hygiene)

## Mission

The current `CHANGELOG.md` `[Unreleased]` section accumulates Wave 1–4 features (coach mode, setup subcommand, signed DMG, KR5 benchmark harness, full toolset, write tools, deletes). Meanwhile:

- `ROADMAP.md` marks those items `[x]` complete.
- `CLAUDE.md` "Build, test, release" describes a `tag vX.Y.Z` workflow that has not been exercised since the initial commit.
- `docs/internal-beta/` is about to ship users a binary with no released version to install or refer to.

Goal: cut a real release (most likely `v0.4.0`) before TP-041's v0.5 dogfood begins.

Concretely:

1. Decide the version (likely `v0.4.0` — the last fully-shipped wave per ROADMAP; **confirm with `git log` and ROADMAP**).
2. Move all currently-unreleased entries under a new `## [v0.4.0] - YYYY-MM-DD` section in `CHANGELOG.md`.
3. Run `make snapshot` (GoReleaser dry-run) and resolve any issues before tagging.
4. Tag `v0.4.0` on `main`.
5. Confirm the release workflow ran green.
6. Update `docs/internal-beta/install.md` (if it exists) and any other doc that references a version, to point at the tag.

Audit ref: 2026-05-16 documentation audit, CHANGELOG vs docs phasing mismatch.

PRD anchors: §7.4 release process.
CLAUDE.md hard rules: "Tag `vX.Y.Z` on `main` to trigger the release workflow. Tags are immutable."

Complexity: Blast radius 3 (release artefact + docs), Pattern novelty 2 (first real tag), Security 2 (signing must work), Reversibility 1 (re-tag is forbidden; bad release → patch v0.4.1) = 8 → Review Level 2. Size: S.

## Dependencies

- **TP-055** (reconcile doc conflicts) — soft. If TP-055 lands first, CHANGELOG and ROADMAP are already in sync. If this lands first, leave TP-055 a clean post-release baseline to work from. Coordinate via STATUS.md.
- **TP-037** (signed macOS installer) — must be working. Cutting a release that fails signing is worse than not cutting one.

## Context to Read First

- `CLAUDE.md` "Build, test, release" section.
- `CHANGELOG.md` — current `[Unreleased]` block.
- `ROADMAP.md` — checkbox state per wave.
- `.github/workflows/release.yml` (or equivalent) — what runs on tag push.
- `.goreleaser.yaml` (or `.goreleaser.yml`) — release artefact config.
- `docs/internal-beta/install.md` if present.

## File Scope

- `CHANGELOG.md` — main edits.
- `docs/internal-beta/install.md` and any other doc referencing a version.
- README.md — if it mentions "unreleased" or a placeholder version, fix it.
- `STATUS.md`.

Out of scope:
- Code changes — this is a release-hygiene task.
- Changing release tooling beyond what `make snapshot` surfaces.
- Bumping to v1.0 (the PRD has more to land before v1.0).
- Pushing to any non-main branch.

## Steps

### Step 1: Confirm the version
- [ ] Run `git tag --list` — should be empty or only prerelease tags. Record state.
- [ ] Read ROADMAP.md and decide: is v0.4.0 the right tag, or should it be v0.4.0-beta.1? Record the decision and rationale in `STATUS.md`.
- [ ] Get explicit human sign-off on the version number before tagging. (Tags are immutable.)

### Step 2: CHANGELOG move
- [ ] Move every entry currently under `[Unreleased]` into a new `## [v0.4.0] - 2026-05-16` (use the actual cut date) section.
- [ ] Leave `[Unreleased]` empty (or seeded with TP-042..TP-070 entries already in flight).
- [ ] Add a comparison link at the bottom (`[v0.4.0]: https://github.com/ricardocabral/icuvisor/releases/tag/v0.4.0`).

### Step 3: Snapshot dry-run
- [ ] `make snapshot`. Resolve any GoReleaser errors. Inspect the produced artefacts in `dist/`.
- [ ] If signing locally is possible, verify the signed DMG opens. (TP-037 owns the signing infra; loop in if blocked.)

### Step 4: Tag
- [ ] Confirm `main` is clean and up-to-date.
- [ ] `git tag -a v0.4.0 -m "Release v0.4.0"` (or signed tag if convention requires).
- [ ] `git push origin v0.4.0`.
- [ ] Watch the release workflow. If it fails, **do not retag** — open a fix PR and tag `v0.4.1` per CLAUDE.md.

### Step 5: Update install/docs references
- [ ] `docs/internal-beta/install.md` (or equivalent) points to `v0.4.0` artefacts.
- [ ] Any "latest" version in README is consistent.

### Step 6: Verify
- [ ] GitHub Releases page shows `v0.4.0` with signed assets.
- [ ] `gh release view v0.4.0` succeeds.
- [ ] `bin/icuvisor --version` (after rebuilding) reports `v0.4.0`.
- [ ] Commit: `TP-071 cut v0.4.0 release`.

## Acceptance Criteria

- `v0.4.0` (or agreed alternative) is tagged on `main`.
- Release workflow succeeded; signed assets published.
- `CHANGELOG.md` has a real `[v0.4.0]` section.
- Install docs reference the tag.
- No retag. Any fix is a `v0.4.1`.

## Do NOT

- Do not retag if the release workflow fails. Ship a patch.
- Do not push the tag without human sign-off on the version number.
- Do not tag from a branch other than `main`.
- Do not include unreleased TP-042..TP-070 work in `v0.4.0` unless it has actually merged.
- Do not bypass signing (`--no-gpg-sign`) or skip pre-flight checks.

## Documentation

- `STATUS.md`
- `CHANGELOG.md` (main artefact)
- Any install doc referencing a version.

## Git Commit Convention

Conventional Commits, prefixed `TP-071`. Tag message itself is `Release v0.4.0`.

---

## Amendments

_Add amendments below this line only._
