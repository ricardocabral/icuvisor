# TP-066 — Extract Strava heuristic + cursor codec from `internal/tools/get_activities.go` (audit God-module)

## Mission

`internal/tools/get_activities.go` is 738 LOC and (even after TP-044 lands the pagination driver refactor) still folds together:

- Strava-blocked detection: `isStravaBlocked` (~lines 650-687), a 37-LOC decision tree branching on upstream markers and `N/A` field heuristics.
- Cursor codec: opaque `next_page_token` encode/decode.
- Unit / pace shaping helpers (`round`, `intValue`, `firstFloat`).
- Schema literals.
- Handler glue.

Goal:

1. Extract Strava detection to `internal/tools/get_activities_strava.go` with a table-driven test covering the audit's listed cases (Strava-imported marker present; `N/A` HR/cadence on power-only file; manual entries).
2. Extract cursor codec to `internal/tools/get_activities_cursor.go` (or `internal/tools/cursor.go` if generic enough across tools — but resist generalizing unless 2+ tools share the codec today).
3. Move row-shaping helpers to `internal/tools/get_activities_row.go`.

This is a focused, behaviour-preserving split. No schema or wire change.

Audit ref: 2026-05-16 Go audit, God-module section.

PRD anchors: §7.2.C `get_activities`; §7.2.E pagination invariants; PRD note on Strava-imported labelling.
CLAUDE.md hard rules: small files; one obvious concern per file; "small interface".

Complexity: Blast radius 2 (tools package internal), Pattern novelty 1, Security 1, Reversibility 1 = 5 → Review Level 1. Size: S.

## Dependencies

- **TP-044** — must land first. TP-044 refactors the pagination driver; doing this split on top of TP-044's cleaner shape is much safer.
- **TP-048** — soft. If `tools.DecodeStrict` / `tools.TextResult` have landed, use them.

## Context to Read First

- `internal/tools/get_activities.go` (post-TP-044) — full file.
- `internal/tools/get_activities_test.go` — locked-in cases.
- The PRD section on Strava-imported activities labelling — the contract `isStravaBlocked` defends.

## File Scope

- New: `internal/tools/get_activities_strava.go` — pure-function Strava detection + its table test (`get_activities_strava_test.go`).
- New: `internal/tools/get_activities_cursor.go` — encode/decode + invariants (opacity, byte-identical round-trip) + test.
- New (optional): `internal/tools/get_activities_row.go` — row-shaping helpers if ≥ 30 LOC.
- Trim: `internal/tools/get_activities.go` to handler + schema + glue.
- `CHANGELOG.md`, `STATUS.md`.

Out of scope:
- Changing the Strava detection rules.
- Changing the `next_page_token` format (opaque contract; must round-trip unchanged).
- Generalizing the cursor codec across tools.
- Changing tool schema or output shape.

## Steps

### Step 1: Capture golden tests
- [ ] Before any move, ensure there is a test that asserts the `next_page_token` format is byte-identical for at least one full page sweep. If not, add one (using a fixed seed / fixed page).
- [ ] Same for `isStravaBlocked` — table test covering each branch.

### Step 2: Extract Strava heuristic
- [ ] Move `isStravaBlocked` + helpers to `get_activities_strava.go`. Keep package-private.
- [ ] Move tests to `get_activities_strava_test.go`.
- [ ] Run all checks.

### Step 3: Extract cursor codec
- [ ] Move encode/decode + token format constants to `get_activities_cursor.go`.
- [ ] Run all checks; token must round-trip byte-identical against pre-refactor snapshot.

### Step 4: Extract row helpers (if ≥ 30 LOC)
- [ ] Move `round`, `intValue`, `firstFloat` etc. to `get_activities_row.go`.
- [ ] If under 30 LOC, leave them in `get_activities.go` — premature abstraction is worse than the smell.

### Step 5: Verify
- [ ] `make build` / `test` / `test-race` / `lint`.
- [ ] `scripts/snapshot_tool_schemas.go` diff empty.
- [ ] `wc -l internal/tools/get_activities*.go` — main file ≤ ~350 LOC.

## Acceptance Criteria

- `isStravaBlocked` lives in its own file with a table-driven test.
- Cursor codec lives in its own file with a round-trip test.
- Main `get_activities.go` shrinks to handler + schema + glue (≤ ~350 LOC).
- Wire shape and tool schema are byte-identical.
- All `make` checks pass.

## Do NOT

- Do not change Strava-detection rules without an explicit `STATUS.md` note and PRD pointer.
- Do not change `next_page_token` format. Opacity is a contract.
- Do not generalize the cursor codec across tools in this task.
- Do not touch unit/pace canonicalization rules.

## Documentation

- `STATUS.md`
- `CHANGELOG.md` `[Unreleased]` under "Changed" (internal refactor).

## Git Commit Convention

Conventional Commits, prefixed `TP-066`. One commit per extraction.

---

## Amendments

_Add amendments below this line only._
