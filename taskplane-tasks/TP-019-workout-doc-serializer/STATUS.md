# TP-019-workout-doc-serializer: TP-019-workout-doc-serializer — Status

**Current Step:** Step 4: Golden-file round-trip tests
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-13
**Review Level:** 0
**Review Counter:** 0
**Iteration:** 1
**Size:** M

---

### Step 1: Enumerate DSL coverage
**Status:** ✅ Complete

- [x] List every step / segment type the read side currently surfaces (intervals, ramps, reps, recoveries, freeride, target ranges, cadence/HR/power targets, %FTP / %LTHR / absolute)
- [x] Identify which of those have a stable DSL form vs which are read-only conveniences; document gaps in `STATUS.md`
- [x] Record the upstream DSL grammar references in `STATUS.md`

---

### Step 2: Serializer
**Status:** ✅ Complete

- [x] `Serialize(WorkoutDoc) (string, error)` emits a deterministic DSL string (stable ordering of attributes per step)
- [x] Reject unsupported step types with a typed error containing the offending step (no silent drops)
- [x] Preserve free-text `description` on steps verbatim

---

### Step 3: Parser parity (read-side compat)
**Status:** ✅ Complete

- [x] Confirm the read-side parser produces the same structured shape the serializer round-trips through; if it does not, write a thin adapter
- [x] Do not change the public read shape (TP-013 owns that)

---

### Step 4: Golden-file round-trip tests
**Status:** 🟨 In Progress

- [ ] `testdata/` contains pairs: `XX-structured.json` + `XX-dsl.txt`
- [ ] Test 1: parse DSL → struct → re-serialize → byte-equal to original DSL (or documented canonicalization)
- [ ] Test 2: load structured JSON → serialize → parse → deep-equal to original struct
- [ ] Cover every step type from Step 1; one golden per type at minimum

---

### Step 5: Hook points (no consumers yet)
**Status:** ⬜ Not Started

- [ ] Public API: `workoutdoc.Serialize`, `workoutdoc.Parse`, the `WorkoutDoc` type
- [ ] Wire nothing into MCP tools here; downstream tasks consume

---

### Step 6: Verify
**Status:** ⬜ Not Started

- [ ] `make test`, `make build`, `make lint`, `go test -race ./...`

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Read-side workout_doc coverage is passthrough, not a typed parser: `intervals.Workout.WorkoutDoc` and `intervals.Event.WorkoutDoc` are `any`; default tools expose only `workout_doc_summary` (`present`, `top_level_keys`, `step_count`, `name`), while `get_workouts_in_folder include_full:true` and event `full` payloads preserve raw upstream JSON. Repo fixtures currently surface `workout_doc.steps`, duration-only steps, duration+power targets, duration+pace targets, and repeat blocks (`reps` + nested `steps`). No repo fixture currently surfaces ramp/freeride/recovery labels, cadence targets, HR targets, absolute watts/bpm/pace, target ranges, or step descriptions. | Use `internal/workoutdoc` typed shape to cover the stable DSL subset and reject unsupported raw shapes rather than silently dropping them. | `internal/intervals/workout_library.go`; `internal/intervals/events.go`; `internal/tools/get_events.go`; `internal/tools/get_workouts_in_folder.go`; `internal/intervals/testdata/workout_library/*.json` |
| Stable DSL forms identified from upstream docs: line steps with duration or distance plus target, target ranges, cadence suffixes, ramps, `freeride`, repeat blocks (`Nx` header/line plus child step lines), and cue text before the first duration. Stable target families include power (`%FTP` implicit as `75%`, absolute watts, zones), heart rate (`% HR`, `% LTHR`, bpm/ranges), pace (`% Pace`, zone pace, absolute `mm:ss/unit` pace), and cadence `rpm`. Current read-only conveniences/gaps: `workout_doc_summary`, top-level `name`, `target(s)`, `moving_time`, `distance`, and load fields are summaries/metadata, not DSL steps; repo fixtures do not prove separate JSON encodings for ramps, freeride, recoveries, cadence, HR, absolute targets, RPE, or step cue descriptions, so the serializer should support explicit typed forms for them and fail on unknown raw fields. | Canonical serializer subset should document unsupported/lossy fields through typed errors now; downstream write tools can surface warnings later. | Upstream forum topic 123701; PRD §7.2.C; PRD §7.4 #19 |
| There is no existing read-side workout DSL parser in this worktree; TP-013 currently preserves raw `workout_doc` values. `internal/workoutdoc.WorkoutDoc` uses JSON tags matching the known raw read shape (`steps`, `duration`, `reps`, `power`, `pace`, etc.), and `Parse` now targets the same typed struct that `Serialize` emits. | No adapter into MCP read tools is required for TP-019; consumers can convert preserved raw JSON into `workoutdoc.WorkoutDoc` without changing tool response shape. | `internal/workoutdoc/parse.go`; `internal/workoutdoc/types.go`; `internal/intervals/testdata/workout_library/*.json` |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-13 | Task staged | STATUS.md auto-generated by task-runner |
| 2026-05-13 12:44 | Task started | Runtime V2 lane-runner execution |
| 2026-05-13 12:44 | Step 1 started | Enumerate DSL coverage |

---

## Blockers

*None*

---

## Notes

- Upstream DSL/API references recorded for TP-019:
  - Intervals.icu API docs shell: https://intervals.icu/api-docs.html (loads OpenAPI from https://intervals.icu/api/v1/docs).
  - OpenAPI write endpoints state event/workout creation accepts native Intervals.icu format in the `description` field: `/api/v1/athlete/{id}/events` `createEvent`, `/api/v1/athlete/{id}/workouts` `createWorkout`; read schemas expose `workout_doc` only as an untyped object.
  - Workout Builder Syntax Quick Guide: https://forum.intervals.icu/t/workout-builder-syntax-quick-guide/123701 (duration/distance, target families/ranges, cadence, ramps, freeride, repeats, cue text, timed prompts).
  - Original Workout builder announcement: https://forum.intervals.icu/t/workout-builder/1163 (plain-text workout steps, percentages, ranges, ramps, repeat examples).
  - API asymmetry discussion: https://forum.intervals.icu/t/api-create-or-update-workout-without-using-description-text-syntax/124215 (public report that JSON `workout_doc.steps` is not accepted as the write representation).