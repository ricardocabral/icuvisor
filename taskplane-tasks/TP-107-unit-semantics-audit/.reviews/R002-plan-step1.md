# Plan Review — Step 1 (R002)

Verdict: **REVISE**

The updated status adds the right high-level intent, but it still does not contain the concrete serializer matrix requested in R001. Before implementation, make the plan explicit enough that a worker cannot accidentally miss aliases or scalar/range distinctions.

Required additions:

1. Name the direct table-driven test to add in `internal/workoutdoc` and state that each case builds a `WorkoutDoc`, calls `Serialize`, and asserts the exact DSL string.
2. List the exact target-unit cases to cover, including aliases:
   - power: blank/default percent FTP, `PERCENT_FTP`, `%FTP`, watts aliases, and power zone scalar/range.
   - pace: percent-threshold aliases, `PACE` numeric scalar/range, pace zone scalar/range, and current text pace form such as `5:00/km Pace`.
   - heart rate: `% HR` vs `% LTHR` via `PERCENT_HR`/`PERCENT_MAX_HR` and `PERCENT_LTHR` aliases, plus BPM and HR zone scalar/range.
3. Explicitly decide how `MINS_KM`/`MINS_MILE` are handled in Step 1: either add an additive serializer/syntax fix with tests, or add an unsupported-unit regression/discovery. Do not silently coerce these into `PACE`.
4. Keep serializer changes test-driven and limited to mismatches with the documented contract; otherwise this step should be test-only.
5. Include the targeted verification command: `go test ./internal/workoutdoc`, and plan to update `STATUS.md` discoveries with what was proven.

Once those concrete cases are written into the plan/status, Step 1 will be scoped appropriately.
