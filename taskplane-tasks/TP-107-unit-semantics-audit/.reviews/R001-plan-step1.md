# Plan Review — Step 1

Verdict: **Changes requested before implementation**

The step direction is right, but the current plan/status checkboxes are too high-level to guarantee the intended unit-semantics coverage. Before coding, make the Step 1 plan explicit about the exact serialization cases that will be locked down.

## Required plan additions

1. Add a direct, table-driven serializer test in `internal/workoutdoc` rather than relying only on existing golden round-trips. Each case should build a small `WorkoutDoc`, call `Serialize`, and assert the exact DSL string.
2. Cover the actual unit families from `workoutTargetUnits`:
   - Power: blank/default percent FTP, `PERCENT_FTP`/`%FTP`, watts aliases, and power zones/ranges.
   - Pace: percent-threshold aliases, pace-zone ranges, numeric `PACE`, and the existing text pace form (e.g. `5:00/km Pace`) if that is the supported way to represent absolute pace today.
   - Heart rate: distinguish `% HR` from `% LTHR` by testing `PERCENT_HR`/`PERCENT_MAX_HR` and `PERCENT_LTHR` aliases; include BPM and HR zone if included in the matrix.
3. Include scalar and range cases where the DSL differs (`75%` vs `88-94%`, `Z2` vs `Z2-Z3`, etc.).
4. Be explicit about how PRD pace-unit language is handled. If `MINS_KM`/`MINS_MILE` are not supported by the workout serializer’s structured target model, do not silently coerce them; either add a documented additive fix with tests or record the unsupported surface as a discovery for follow-up.
5. If any serializer behavior changes, update matching syntax metadata/tests as needed; otherwise keep this step test-only.
6. Finish by running `go test ./internal/workoutdoc` and updating `STATUS.md` discoveries with what was proven.

With those details added, the plan is appropriately scoped for Step 1.
