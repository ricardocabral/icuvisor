# TP-073 — Move `get_workouts_in_folder` description behind `include_full`

## Mission

Resolves [GitHub issue #12](https://github.com/ricardocabral/icuvisor/issues/12) — *TP-016: review `get_workouts_in_folder` default verbosity*.

PRD §7.2.D + CLAUDE.md hard rule 5 require tools to be terse by default; heavy payloads sit behind `include_full: true`. The v0.2 dogfood (TP-016) found that `get_workouts_in_folder` currently includes the full `description` field (which can be multi-paragraph coaching prose) in its terse default response. The response size stayed below the 30k-token soft ceiling, but coaching-prose descriptions consume conversation budget unnecessarily when the LLM is iterating over a folder listing.

Fix: drop `description` from the terse default; keep it behind `include_full: true` alongside the existing `workout_doc` opt-in.

PRD anchors: §7.2.D terse-by-default contract; line 332 (~30k-token soft ceiling).
CLAUDE.md: hard rule 5.

Complexity: Blast radius 1, Pattern novelty 1 (matches existing `workout_doc` gating in the same file), Security 1, Reversibility 1 = 4 → Review Level 1. Size: XS.

## Dependencies

- None. Independent of the other v0.2/v0.3 dogfood follow-ups (TP-072, TP-074, TP-075, TP-076, TP-077).
- Aligns conceptually with the `get_workout_library` terse pattern at the folder level. Do NOT change `get_workout_library` (folder-level descriptions are short metadata; keeping them terse is fine).

## Context to Read First

- [`docs/prd/PRD-icuvisor.md`](../../docs/prd/PRD-icuvisor.md) §7.2.D — terse-by-default contract.
- [`docs/dogfood/v0.2-findings.md`](../../docs/dogfood/v0.2-findings.md) lines around `get_workouts_in_folder` (search for the tool name) — the dogfood verdict.
- [`taskplane-tasks/TP-016-v02-dogfood-validation/STATUS.md`](../TP-016-v02-dogfood-validation/STATUS.md) — original follow-up tracking.
- [`internal/tools/get_workouts_in_folder.go`](../../internal/tools/get_workouts_in_folder.go):
  - Response struct `workoutInFolderRow`: lines 30–45.
  - Shaping `workoutInFolderToRow()`: lines 119–128.
  - Input schema: lines 130–135 (confirm `include_full` is already declared — it is, since `workout_doc` already uses it).
- [`internal/tools/get_workout_library_test.go`](../../internal/tools/get_workout_library_test.go):
  - `TestGetWorkoutsInFolderFiltersAndPreservesWorkoutDocWithIncludeFull`: lines 118–155.
  - `TestGetWorkoutsInFolderHidesWorkoutDocByDefault`: lines 157–177.

## File Scope

- `internal/tools/get_workouts_in_folder.go` — gate `description` behind `includeFull`.
- `internal/tools/get_workout_library_test.go` — extend the two existing tests; add an explicit "description hidden by default" assertion.
- `CHANGELOG.md` — `[Unreleased]` under "Changed" (this is a default-response shape change; document it clearly).
- `STATUS.md` (this dir).

Out of scope:
- Changing `get_workout_library` (folder listings stay as-is).
- Changing any other field on `workoutInFolderRow`.
- Restructuring the tool's input schema.

## Steps

### Step 1: Reproduce the verbose default in a test

- [ ] In `TestGetWorkoutsInFolderHidesWorkoutDocByDefault` (line 157), add an assertion that `description` is also absent from the terse default row.
- [ ] Confirm the test fails on `main`.

### Step 2: Fix the shaping function

- [ ] In `workoutInFolderToRow()`, only populate the `Description` field when `includeFull` is true. (Either pass `includeFull` into the helper, or null the field after-the-fact in the caller — match the existing pattern used for `workout_doc`.)
- [ ] Keep `workout_doc_summary` in terse mode if it's already present (the dogfood note didn't flag it).

### Step 3: Tests

- [ ] Extend `TestGetWorkoutsInFolderFiltersAndPreservesWorkoutDocWithIncludeFull` (line 118) to assert `description` IS present when `include_full: true`.
- [ ] Confirm the modified terse test from Step 1 now passes.
- [ ] Run `make test` and `make test-race`.

### Step 4: Build + lint

- [ ] `make build`, `make lint`.

### Step 5: Close the GitHub issue

- [ ] Update `CHANGELOG.md` under `[Unreleased] → Changed`: `get_workouts_in_folder no longer includes workout descriptions in the terse default response; set include_full: true to access them.`
- [ ] Update `STATUS.md`.
- [ ] Commit: `fix(workouts): gate description behind include_full in get_workouts_in_folder (TP-073, closes #12)`.
- [ ] Reference `Closes #12` in the PR body. After merge, verify auto-close; otherwise `gh issue close 12 --comment "Fixed in <commit-sha> / <PR>"`.

## Acceptance Criteria

- `description` is absent from `get_workouts_in_folder` rows when `include_full` is false (the default).
- `description` is present when `include_full: true`.
- Both tests pass; both terse-and-full coverage paths are explicit.
- `make build`, `make test`, `make test-race`, `make lint` clean.
- CHANGELOG entry under "Changed" makes the default-shape change explicit (LLM operators may have prompts that assume `description` is present — they need a heads-up).
- GitHub issue #12 closed.

## Do NOT

- Do not drop `description` from `workoutInFolderRow` entirely — `include_full: true` callers still need it.
- Do not change `get_workout_library`'s folder-level `description` field; it's short metadata and the dogfood didn't flag it.
- Do not introduce a new opt-in argument; reuse the existing `include_full` flag.
- Do not silently rename the field. Schemas are additive-only per the v0.2 stability rules.

## Documentation

- `CHANGELOG.md` `[Unreleased]` under "Changed".
- `STATUS.md` in this dir.

## Git Commit Convention

Conventional Commits, prefixed with TP-073. Example:

```
fix(workouts): gate description behind include_full in get_workouts_in_folder

TP-073. Closes #12.
```

---

## Amendments

_Add amendments below this line only._
