# TP-077 — `create_workout` synthetic write refused

## Mission

Resolves [GitHub issue #9](https://github.com/ricardocabral/icuvisor/issues/9) — *TP-029: `create_workout` refused during dogfood*.

During the v0.3 dogfood (TP-029, W-10 in [`docs/dogfood/v0.3-findings.md:28`](../../docs/dogfood/v0.3-findings.md)), an attempt to create a synthetic workout-library item against the dedicated test athlete failed: upstream refused the POST and `get_workout_library` re-read showed no top-level synthetic workout. Without a created workout, W-11 (`update_workout`) and the D-04 destructive case were both blocked.

Suspected defects (from triage, ordered by likelihood — confirm via live probe):

1. **Field name `type` vs `sport`** — [`internal/intervals/workout_library.go:208-230`](../../internal/intervals/workout_library.go) maps the tool's `sport` parameter to JSON key `"type"`. Upstream may expect `"sport"`, or it may expect `"type"` with values from a different enum than activity-type values.
2. **`folder_id` requirement / format** — The test passed `folder_id: "f-20"` (synthetic) and the tool may require a real folder owned by the athlete, or it may require `null` for top-level workouts. Upstream may reject creates that reference a non-existent folder.
3. **Hierarchy requirement** — Workout library writes may require the workout to be nested under an existing folder structure that the tool doesn't enforce.
4. **Less likely:** `workout_doc` structured payload validation (if the request included structured steps, upstream may reject them on workout-library writes per the known mvilanova #56 — but the existing serializer should already handle this by reducing to the DSL description string).

This task requires **live API probing** against the `.env-dev` test athlete to isolate the upstream contract.

PRD anchors: §7.2.C workout-library CRUD; `workout_doc` description-string DSL serializer (v0.3, TP-019).
CLAUDE.md: hard rule 5; MCP-server conventions ("Schemas matter").

Complexity: Blast radius 1, Pattern novelty 2 (live probe + potentially a folder-seeding step), Security 1, Reversibility 2 (writes against dedicated test athlete) = 6 → Review Level 2. Size: M.

## Dependencies

- None hard. Independent of TP-072, TP-073, TP-074, TP-075, TP-076.
- Useful prior art: [TP-023](../TP-023-workout-library-crud/) shipped the original CRUD path; [TP-019](../TP-019-workout-doc-serializer/) shipped the description-string DSL serializer that constrains the write-side `workout_doc` shape.

## Context to Read First

- [`docs/prd/PRD-icuvisor.md`](../../docs/prd/PRD-icuvisor.md) — workout-library CRUD section; `workout_doc` write-path serializer rule.
- [`docs/dogfood/v0.3-findings.md`](../../docs/dogfood/v0.3-findings.md) line 28 (W-10) and per-tool triage row for `create_workout` / `update_workout`.
- [`internal/tools/create_workout.go`](../../internal/tools/create_workout.go):
  - Request struct: lines 26–33.
  - Handler: lines 55–81.
- [`internal/intervals/workout_library.go`](../../internal/intervals/workout_library.go):
  - `CreateLibraryWorkout`: lines 126–137.
  - `writeWorkoutBody`: lines 208–230. **This is where the suspected `type` vs `sport` mismatch lives.**
- [`internal/tools/create_workout_test.go`](../../internal/tools/create_workout_test.go) — existing pattern.
- Existing TP work: [`taskplane-tasks/TP-019-workout-doc-serializer/`](../TP-019-workout-doc-serializer/), [`taskplane-tasks/TP-023-workout-library-crud/`](../TP-023-workout-library-crud/).

## File Scope

- `internal/intervals/workout_library.go` — fix the field-name mapping in `writeWorkoutBody`.
- `internal/tools/create_workout.go` — adjust validation if folder constraints surface (e.g. require non-empty `folder_id`, or accept `null` for top-level).
- `internal/tools/create_workout_test.go` — update the existing test to match the corrected payload; add coverage for the folder-id permutation.
- `internal/intervals/testdata/workout_library/` — fixture from the live probe.
- `CHANGELOG.md` — `[Unreleased]` under "Fixed".
- `STATUS.md` (this dir).

Out of scope:
- Refactoring `update_workout` / `delete_workout` — their tests are blocked until create works, but their implementations are separate.
- Changing the `workout_doc` DSL serializer (TP-019 territory).
- Adding new workout-library tool capabilities.

## Steps

### Step 1: Live probe to isolate the contract

- [ ] Source `.env-dev`.
- [ ] First, use `get_workout_library` against the test athlete to inventory existing folders. Capture a real folder ID you can write into (or confirm whether top-level writes work with no folder).
- [ ] **Out-of-process probe** (curl / scratch Go — do NOT commit): POST a minimal workout create directly to the intervals.icu workout-library endpoint. Vary fields:
  - `{ "name": "tp-077-probe", "type": "Ride", "folder_id": "<real-folder-id>" }`.
  - Swap `"type"` for `"sport"`.
  - Try `folder_id: null` for top-level.
  - Try with `description` (the DSL string) included.
- [ ] Note which permutation is accepted. Capture sanitized request/response under `internal/intervals/testdata/workout_library/create_request.json` + `create_response.json`.
- [ ] **Clean up:** delete every probe-created workout (use the existing `delete_workout` tool in `full` delete mode, or the intervals.icu UI). Verify via `get_workout_library` that nothing extra remains.

### Step 2: Add a failing test

- [ ] Update `TestCreateWorkoutWithStructuredStepsSerializesDSLAndReturnsReadShape` (or add a new sibling test) so the expected outbound body matches the working payload from Step 1.
- [ ] Confirm the test fails on `main`.

### Step 3: Fix the client + tool

- [ ] Apply the minimum diff in `writeWorkoutBody` (likely a JSON key rename).
- [ ] If folder semantics changed: update the tool's input schema description for `folder_id` to make the contract explicit (`"folder_id: ID of an existing folder owned by the athlete; omit for top-level workouts"`).
- [ ] If folder_id is required, add a clear public validation error when it's missing rather than letting the request fail upstream.

### Step 4: Build + lint + race + live re-validation

- [ ] `make build`, `make test`, `make test-race`, `make lint`.
- [ ] Live re-validation via stdio MCP: `create_workout` → confirm via `get_workout_library` and `get_workouts_in_folder` that the new workout appears at the expected location; then `delete_workout` cleanup; confirm gone.

### Step 5: Document amendment

- [ ] If the upstream contract surfaced a non-obvious shape (e.g. `sport` vs `type` discrepancy with other endpoints), capture it in `docs/upstream-gaps/workout-library-create-payload.md`.

### Step 6: Close the GitHub issue

- [ ] Update `CHANGELOG.md` under `[Unreleased] → Fixed`: `create_workout now succeeds against intervals.icu (was refused upstream due to <root cause>).`
- [ ] Update `STATUS.md`.
- [ ] Commit: `fix(workouts): repair create_workout payload (TP-077, closes #9)`.
- [ ] Reference `Closes #9` in the PR body. After merge, verify auto-close; otherwise `gh issue close 9 --comment "Fixed in <commit-sha> / <PR>"`.

## Acceptance Criteria

- A `create_workout` call against the test athlete is accepted upstream, and the new workout appears in `get_workout_library` / `get_workouts_in_folder` at the expected location.
- The tool's input schema clearly documents the `folder_id` contract.
- A new or updated unit test exercises the corrected payload using a captured fixture.
- `make build`, `make test`, `make test-race`, `make lint` pass.
- Probe-created workouts have been deleted from the test athlete.
- GitHub issue #9 closed.

## Do NOT

- Do not commit probe scratch files.
- Do not leave probe-created workouts on the test athlete.
- Do not paste raw workout IDs / folder IDs / dates that could identify the test account without sanitization.
- Do not introduce a `confirm: true` LLM-controlled override (CLAUDE.md hard rule 5).
- Do not change the `workout_doc` description-DSL serializer — its read→modify→write round-trip is locked by golden-file tests (TP-019).

## Documentation

- `CHANGELOG.md` `[Unreleased]` under "Fixed".
- `STATUS.md` in this dir.
- *Optional:* `docs/upstream-gaps/workout-library-create-payload.md`.

## Git Commit Convention

Conventional Commits, prefixed with TP-077. Example:

```
fix(workouts): repair create_workout payload

TP-077. Closes #9.
```

---

## Amendments

_Add amendments below this line only._
