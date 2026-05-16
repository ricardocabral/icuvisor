# TP-059 — Replace `map[string]any` with typed structs in workout/training bodies (audit High)

## Mission

CLAUDE.md mandates typed structs over `map[string]any`. The 2026-05-16 audit identified three sites still using untyped maps for request/response bodies and schema literals:

- `internal/tools/get_workout_library.go:62-63` — `Full` / `WorkoutDocSummary` field expressed as `map[string]any`.
- `internal/tools/get_training_plan.go:45-46` — similar.
- `internal/intervals/workout_library.go:208` — write body assembled as `map[string]any` instead of a typed request struct.

Goal: introduce typed request/response structs (e.g., `WorkoutDocSummary`, `WorkoutLibraryWriteRequest`) and switch all three sites. Use `json.RawMessage` only for genuinely opaque passthrough (e.g., when the client must forward an upstream blob unchanged via `include_full`).

Schema literals (e.g., `getWorkoutLibraryInputSchema`) can remain as Go-side schema definitions for now, but the **request/response bodies** must be typed.

Audit ref: 2026-05-16 Go audit, "High" severity.

PRD anchors: §7.2.C tool contracts, §7.4 reliability.
CLAUDE.md hard rules: "prefer typed structs over `map[string]any`".

Complexity: Blast radius 2 (three files + their tests), Pattern novelty 1, Security 1, Reversibility 2 (struct field set is now part of the contract) = 6 → Review Level 1. Size: S.

## Dependencies

- **TP-048** — coordinate. TP-048 introduces `tools.DecodeStrict[T]`; once it lands, the new typed structs in this task plug into `DecodeStrict[T]` naturally. If TP-048 lands first, use it. If this lands first, leave a TODO referencing TP-048 at each decode site.
- **TP-019** (workout_doc serializer) — context only; the typed `WorkoutDocSummary` here must round-trip with the serializer's types.

## Context to Read First

- `CLAUDE.md` — Go conventions on JSON / typed structs.
- `internal/tools/get_workout_library.go` — full file.
- `internal/tools/get_training_plan.go` — full file.
- `internal/intervals/workout_library.go` — focus on the write helpers around line 208.
- `internal/workoutdoc/parse.go`, `internal/workoutdoc/serialize.go` — existing typed model for workout DSL.
- Their `_test.go` files — locked-in fixtures.

## File Scope

- `internal/tools/get_workout_library.go` — replace the `map[string]any` field with a typed struct; update tests.
- `internal/tools/get_training_plan.go` — same.
- `internal/intervals/workout_library.go` — typed request body for the write path.
- New file `internal/intervals/workout_library_types.go` (or inline in `workout_library.go`) — the new request/response structs.
- Their test files.
- `CHANGELOG.md`, `STATUS.md`.

Out of scope:
- Tool schemas (the JSON Schema literals can stay as-is for this task).
- Other tools using `map[string]any` legitimately for opaque passthrough.
- The schema literal for `getWorkoutLibraryInputSchema` itself (separate concern; flag in `CONTEXT.md` Technical Debt if it bothers you).

## Steps

### Step 1: Define the typed structs
- [ ] Decide where each struct lives (tools package vs intervals package). Prefer intervals package for request/response shared with the HTTP layer.
- [ ] Use `json:"…,omitempty"` consistently; reuse types from `internal/workoutdoc` where applicable.

### Step 2: Swap request/response bodies
- [ ] Replace each `map[string]any` body with the typed struct.
- [ ] Update tests; verify golden fixtures decode/encode byte-identical (use `cmp.Diff` if needed).
- [ ] `make build` / `test` / `test-race` / `lint`.

### Step 3: Verify
- [ ] `grep -n "map\[string\]any" internal/tools/get_workout_library.go internal/tools/get_training_plan.go internal/intervals/workout_library.go` — zero hits for request/response bodies. (Schema literals exempt.)
- [ ] All `make` checks pass.
- [ ] Commit: `TP-059 use typed structs for workout & training bodies`.

## Acceptance Criteria

- No `map[string]any` in request/response bodies in the three cited files.
- New types are unexported unless they need to be shared across packages.
- Round-trip tests prove the wire shape is unchanged.
- `make` checks pass.

## Do NOT

- Do not rename existing fields or change wire JSON keys.
- Do not absorb the schema literals into typed structs in this task — schema generation is a separate concern.
- Do not introduce a third-party schema/generator library.

## Documentation

- `STATUS.md`
- `CHANGELOG.md` `[Unreleased]` under "Changed".

## Git Commit Convention

Conventional Commits, prefixed `TP-059`.

---

## Amendments

_Add amendments below this line only._
