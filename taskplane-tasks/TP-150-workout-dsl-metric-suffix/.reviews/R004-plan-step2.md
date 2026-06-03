# Plan Review: Step 2 — Implement and test metric suffix behavior

**Verdict:** Approve with required carry-forward notes.

## Why this plan is ready

- It follows the approved Step 1 boundary: keep the default `workoutdoc.Serialize` / `MergeDescription` behavior stable and add an options-aware path for sport-aware writes.
- It explicitly includes the key public-behavior regression: Run + `POWER_HR_PACE` should serialize power zones as `Z2 Power` / `Z2-Z3 Power` when sport context is known.
- It includes coverage for HR/pace orders, `workout_order` decoding/helper behavior, and the previously missed `apply_training_plan` path.

## Carry-forward requirements for implementation

1. Because Step 2 will add `intervals.SportSettings.WorkoutOrder`, run the intervals tests too. The targeted command should include `./internal/intervals` in addition to `./internal/workoutdoc ./internal/tools`.
2. Tool-layer tests should assert the actual write params for the direct paths, not only the serializer/helper: `add_or_update_event` planned WORKOUT, `create_workout`, and `update_workout` when `sport` is supplied. Also keep a fallback test for `update_workout` with `workout_doc` but no `sport` preserving bare output.
3. Keep `internal/workoutdoc` independent of `internal/intervals`; pass primitive/options data in, and avoid post-serialization string rewriting at call sites.
4. For `apply_training_plan`, derive options from the template workout `type` and athlete profile sport settings so dry-run conflict detection and actual writes use the same generated description.

No plan revision is needed before starting Step 2 if these notes are carried into the implementation.
