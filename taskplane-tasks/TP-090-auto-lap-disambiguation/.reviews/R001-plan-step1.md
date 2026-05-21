# Plan Review: Step 1 — Model interval-source heuristics

## Verdict

Changes requested. The current Step 1 plan is still just the task checklist; it does not yet define the heuristic contract tightly enough to make later code/test reviews objective.

## Required plan refinements

1. **Record the exact field inventory before choosing heuristics.**
   - `internal/intervals/activity_details.go` currently types only `id`, `name`, `type`, `unit`, start/end indexes/times/distances, `distance`, `duration`, power/HR/pace, while preserving all upstream keys in `Raw`.
   - The Step 1 plan should say whether source markers will be searched only in typed fields or also in `IntervalsDTO.Raw`, `ActivityInterval.Raw`, and `IntervalGroup.Raw`, and should list any candidate keys found. If no explicit marker exists, record that in `STATUS.md`.

2. **Define a deterministic classifier, not just “near-uniform”.**
   The plan should specify thresholds and precedence, for example:
   - minimum interval count before auto-lap suspicion is allowed;
   - how many rows may be excluded as warmup/cooldown/last partial;
   - accepted lap targets (1 km / 1 mi, and any duration-lap targets if included);
   - absolute/relative tolerances in meters/seconds;
   - handling of missing/zero/negative `distance` or `duration`;
   - whether contiguous `start_distance`/`end_distance` or `start_index`/`end_index` is required.

3. **Avoid false positives for real structured workouts.**
   Repeated structured workouts can also be near-uniform, e.g. 6x1 km reps. The plan needs precedence rules that keep these as `structured_workout` or `unknown` when there are strong structured signals (`icu_groups`, non-generic interval names/types such as warmup/rest/work/recovery, workout-step markers in raw fields, etc.). “Uniform distance” alone is not enough to label `device_laps`.

4. **Clarify unit assumptions.**
   The current shaped response names interval distance as `distance_m`, but the upstream typed DTO field is just `distance`, and `unit` appears to be an intervals.icu unit enum for targets/pace rather than necessarily the distance unit. The plan should state what units the heuristic assumes for `distance`, `start_distance`, and `end_distance`, how 1 mile is represented (`1609.344m` target), and what happens when those assumptions cannot be proven.

5. **Decide helper/API placement now.**
   Since Step 3 needs analyzers to propagate the same signal, the classifier should not be buried in `get_activity_intervals` only. The plan should name the intended shared location and exported/unexported API shape, with typed constants for `structured_workout`, `device_laps`, and `unknown` to avoid string drift.

6. **Specify metadata semantics.**
   The plan should say whether `_meta.auto_lap_suspected` is always emitted as a boolean or only emitted when true, and whether any diagnostic reason/confidence stays internal/test-only. It should also state that unavailable/Strava-blocked responses keep their existing shape and only get the new metadata if the classifier has interval rows to evaluate.

7. **Add Step 1 acceptance examples.**
   Before implementation, capture the intended classifications in `STATUS.md`: structured intervals, 1 km auto-laps, 1 mi auto-laps, unknown/insufficient rows, and a near-uniform structured-repeat negative case. These examples will drive Step 4 fixtures and prevent the analyzer decline behavior from being based on an over-broad heuristic.

## Summary

The overall task direction is sound, but Step 1 should produce a documented, reviewable heuristic contract in `STATUS.md` before code starts. Without the specifics above, later steps risk adding metadata that is additive in schema but unreliable in meaning, especially for structured repeat workouts.
