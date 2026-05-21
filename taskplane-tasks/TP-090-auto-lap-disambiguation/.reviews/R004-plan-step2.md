# Plan Review: Step 2 — Add additive meta to interval reads

## Verdict

Approved. The Step 2 checklist is brief, but it is backed by the detailed Step 1 heuristic contract now recorded in `STATUS.md`, which is specific enough to guide implementation.

## Review scope

- Read `PROMPT.md` and `STATUS.md` for TP-090.
- Checked the current `get_activity_intervals` shaping path in `internal/tools/get_activity_details.go`.
- Checked the current interval DTOs in `internal/intervals/activity_details.go` and analyzer package placement in `internal/analysis`.

## Why the plan is acceptable

- The planned shared classifier with typed source constants matches the Step 1 decision to avoid burying inference inside `get_activity_intervals` only.
- The metadata fields are additive and preserve the stable interval row fields, which satisfies the task's compatibility requirement.
- The Step 1 notes already define the missing details Step 2 needs: source precedence, explicit marker handling, structured-signal precedence, tolerances, unit assumptions, insufficient-data behavior, and unavailable/Strava response semantics.

## Implementation cautions for Step 2

1. **Avoid leaking interval-only metadata to other activity tools.**
   `activityReadMeta` is currently shared by `get_activity_details`, successful interval responses, and interval unavailable responses. If Step 2 simply adds `IntervalSource string` and `AutoLapSuspected bool` to that struct without care, zero values can appear on non-interval responses or unavailable responses. Prefer an interval-specific meta struct/embedding, or use pointer/omitempty fields plus explicit interval-success population so that:
   - successful interval responses always include `_meta.interval_source`;
   - successful interval responses always include boolean `_meta.auto_lap_suspected`, including `false`;
   - `get_activity_details` and existing unavailable/Strava-blocked interval shapes do not gain meaningless zero-value classifier metadata.

2. **Keep the helper reusable and pure.**
   Place the classifier in `internal/analysis` as planned, with constants for `structured_workout`, `device_laps`, and `unknown`. It should accept small interval/group samples or a DTO adapter rather than depending on tool response structs, so Step 3 analyzers can use the same source of truth.

3. **Preserve the conservative precedence from Step 1.**
   Structured evidence must beat uniformity. Near-uniform rows with groups, workout-step raw markers, or non-generic work/rest/warmup/cooldown labels should not become `device_laps`.

4. **Do not expand public metadata beyond the task fields.**
   Keep diagnostic reasons/confidence internal or test-only unless a later step intentionally updates schemas/docs for additional public fields.

5. **Lock down at least minimal coverage with the implementation if practical.**
   Step 4 owns the full fixture matrix, but adding focused classifier/meta tests in Step 2 would reduce risk around the shared `activityReadMeta` behavior and the always-present false boolean case.

## Summary

Proceed with Step 2. The main risk is implementation shape rather than plan direction: ensure interval-source metadata is emitted only where the task says it should be, while preserving existing activity detail and unavailable response contracts.
