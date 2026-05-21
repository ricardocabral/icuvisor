# TP-023 — Workout-library CRUD (`create_workout`, `update_workout`, `delete_workout`)

## Mission

Land the workout-library write path. Creates and updates are ungated; `delete_workout` is gated by `ICUVISOR_DELETE_MODE`. Structured steps round-trip via the TP-019 DSL serializer.

Roadmap items (ROADMAP.md v0.3):

- Workout-library CRUD: `create_workout`, `update_workout`, `delete_workout` (delete gated).

PRD anchors: §7.2.C workout-library catalog, §7.4 workout-doc upstream asymmetry.

Complexity: Blast radius 2, Pattern novelty 2, Security 2, Reversibility 2 (delete is destructive) = 8 → Review Level 2. Size: M.

## Dependencies

- **TP-018** — safety gate
- **TP-019** — `workout_doc` serializer
- **TP-013** — workout-library reads (response parity)

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` §7.2.C workout-library, §7.4
- `ROADMAP.md` v0.3
- `internal/tools/get_workout_library*.go`, `internal/tools/get_workouts_in_folder*.go`
- `internal/workoutdoc/`

## File Scope

Expected files:

- `internal/tools/create_workout.go` + `_test.go`
- `internal/tools/update_workout.go` + `_test.go`
- `internal/tools/delete_workout.go` + `_test.go`
- `internal/intervals/` — typed CRUD methods if not present
- `CHANGELOG.md`
- `README.md` catalog
- `taskplane-tasks/TP-023-workout-library-crud/STATUS.md`

## Steps

### Step 1: `create_workout`

- [ ] Inputs: `name`, `folder_id` (optional), `description` (free-text), `workout_doc` (structured; serialized via TP-019), `tags[]`, `sport`
- [ ] Emit DSL string for `workout_doc` on upload
- [ ] Response is the read shape for the new workout (round-trip parity with TP-013)
- [ ] Tests: create with structured steps, create with free-text only, golden-fixture round-trip from TP-019

### Step 2: `update_workout`

- [ ] Inputs: `workout_id` + sparse fields (same set as create)
- [ ] Partial-update semantics: omitted fields untouched
- [ ] Tests: rename, swap `workout_doc`, append tag

### Step 3: `delete_workout` (gated)

- [ ] Registered only in `full` mode (TP-018 `CanDelete`)
- [ ] Inputs: `workout_id`
- [ ] No `confirm` argument; the gate is the registration
- [ ] Tests: success in `full`; absent from catalog in `safe` and `none`

### Step 4: Verify

- [ ] `make test`, `make build`, `make lint`, `go test -race ./...`
- [ ] Manual smoke against the test athlete: create → update → re-read fidelity; delete in `full` mode

## Reference Implementation Policy

- `hhopke/intervals-icu-mcp` (MIT) may be consulted for endpoint shapes. Do not depend on it.
- GPL/copyleft implementation code is off limits — do not read, copy, paraphrase, or transliterate.

## Acceptance Criteria

- `create_workout` and `update_workout` registered in `safe` and `full`; absent in `none`.
- `delete_workout` registered only in `full`.
- Round-trip parity with TP-013 reads.
- No `confirm` argument anywhere.

## Do NOT

- Do not implement bulk delete here.
- Do not send structured `workout_doc` to intervals.icu; serialize via TP-019.
- Do not allow `update_workout` to clear `workout_doc` unless an explicit empty step list is supplied.

## Documentation

Must update:

- `STATUS.md`
- `README.md` catalog
- `CHANGELOG.md`

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-023`, for example: `TP-023 add create_workout tool`.

---

## Amendments

_Add amendments below this line only._
