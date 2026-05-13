# TP-019 — `workout_doc` write-path serializer (description-string DSL round-trip)

## Mission

intervals.icu rejects structured `workout_doc` payloads on writes (mvilanova #56); the upstream write path accepts only the free-text description-string DSL. Build a serializer that takes the structured `workout_doc` we expose on reads (TP-013) and emits the description-string DSL on writes, then lock the read → modify → write → read fidelity with golden-file tests.

Roadmap items (ROADMAP.md v0.3):

- `workout_doc` write-path serializer: structured steps round-trip back to the description-string DSL on upload.
- Read → modify → write → read fidelity locked by golden-file tests.

PRD anchors: §7.2.C `create_workout` / `update_workout` / event `workout_doc`, §7.4 upstream-asymmetry note.

Complexity: Blast radius 2 (workout-library + event writes share it), Pattern novelty 3 (DSL grammar), Security 1, Reversibility 2 = 8 → Review Level 2. Size: M.

## Dependencies

- **TP-013** — workout-library + custom-items reads (defines the structured `workout_doc` shape we round-trip)
- **TP-007** — response shaping (consumed indirectly via reads)

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` §7.2.C, §7.4
- `ROADMAP.md` v0.3
- `internal/tools/` workout-library reads
- intervals.icu public docs on workout-description DSL syntax (forum + API reference); record the canonical link in `STATUS.md`

## File Scope

Expected files:

- `internal/workoutdoc/` — new package: types for the structured form, a `Serialize(WorkoutDoc) (string, error)` emitter, and a `Parse(string) (WorkoutDoc, error)` consumer (parser may already exist on the read side; deduplicate if so)
- `internal/workoutdoc/testdata/` — golden files: pairs of structured JSON + DSL strings covering each step type
- `internal/workoutdoc/*_test.go`
- `CHANGELOG.md`
- `taskplane-tasks/TP-019-workout-doc-serializer/STATUS.md`

Do **not** add write tools that consume the serializer here — that is TP-020 / TP-023.

## Steps

### Step 1: Enumerate DSL coverage

- [ ] List every step / segment type the read side currently surfaces (intervals, ramps, reps, recoveries, freeride, target ranges, cadence/HR/power targets, %FTP / %LTHR / absolute)
- [ ] Identify which of those have a stable DSL form vs which are read-only conveniences; document gaps in `STATUS.md`
- [ ] Record the upstream DSL grammar references in `STATUS.md`

### Step 2: Serializer

- [ ] `Serialize(WorkoutDoc) (string, error)` emits a deterministic DSL string (stable ordering of attributes per step)
- [ ] Reject unsupported step types with a typed error containing the offending step (no silent drops)
- [ ] Preserve free-text `description` on steps verbatim

### Step 3: Parser parity (read-side compat)

- [ ] Confirm the read-side parser produces the same structured shape the serializer round-trips through; if it does not, write a thin adapter
- [ ] Do not change the public read shape (TP-013 owns that)

### Step 4: Golden-file round-trip tests

- [ ] `testdata/` contains pairs: `XX-structured.json` + `XX-dsl.txt`
- [ ] Test 1: parse DSL → struct → re-serialize → byte-equal to original DSL (or documented canonicalization)
- [ ] Test 2: load structured JSON → serialize → parse → deep-equal to original struct
- [ ] Cover every step type from Step 1; one golden per type at minimum

### Step 5: Hook points (no consumers yet)

- [ ] Public API: `workoutdoc.Serialize`, `workoutdoc.Parse`, the `WorkoutDoc` type
- [ ] Wire nothing into MCP tools here; downstream tasks consume

### Step 6: Verify

- [ ] `make test`, `make build`, `make lint`, `go test -race ./...`

## Reference Implementation Policy

- `hhopke/intervals-icu-mcp` (MIT) may be consulted for DSL form, **only** as black-box behavioural reference. Do not copy.
- `mvilanova/intervals-mcp-server` is GPLv3 — do not read, copy, paraphrase, or transliterate. This is the most acute risk area in v0.3; if you find yourself wanting to peek, stop and post in `STATUS.md` instead.

## Acceptance Criteria

- `internal/workoutdoc` exposes `Serialize`, `Parse`, and a typed `WorkoutDoc`.
- Golden-file round-trip tests pass for every supported step type.
- Unsupported step types fail with a typed, actionable error.
- No write tool yet consumes the serializer (consumers land in TP-020 / TP-023).

## Do NOT

- Do not paste DSL text from any GPL project into testdata; build fixtures from intervals.icu's own documentation or against the maintainer's own account.
- Do not silently drop step types the serializer cannot emit.
- Do not change the read-side `workout_doc` shape — that is a TP-013 contract.

## Documentation

Must update:

- `STATUS.md`
- `CHANGELOG.md`

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-019`, for example: `TP-019 add workoutdoc serializer scaffold`.

---

## Amendments

_Add amendments below this line only._
