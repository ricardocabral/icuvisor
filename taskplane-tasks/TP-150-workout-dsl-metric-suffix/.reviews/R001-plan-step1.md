# Plan Review: Step 1 — Design the sport-aware suffix boundary

**Verdict:** Needs changes before implementation.

## What looks good

- Keeping `workoutdoc.Serialize` / `MergeDescription` unchanged for no-context callers is the right compatibility boundary.
- Centralizing suffix rendering in `internal/workoutdoc` via an options-aware API is preferable to string-rewriting in each write tool.
- The documented behavior of emitting explicit zone-family suffixes (`Power`, `HR`, `Pace`) whenever sport priority context is known is a reasonable clean-room response to the upstream ambiguity.

## Blocking issues to resolve

1. **`workout_order` is not currently decoded anywhere.**
   The plan says write call sites will pass the sport's `workout_order` from the athlete profile, but `intervals.SportSettings` has no such field today, and the current task file scope/status does not mention adding it. Implementation should explicitly add/decode the upstream field (likely in `internal/intervals/types.go`) and test it, or choose a different source. Without this, the proposed boundary cannot be implemented.

2. **`update_workout` does not always know the target sport.**
   `create_workout` has required `sport`, and `add_or_update_event` has `type` for WORKOUT events, but `update_workout` allows `workout_doc` updates without `sport`. The plan needs a concrete rule for that case: infer existing sport somehow, require/pass sport for sport-aware serialization, or intentionally fall back to bare serialization and document the remaining limitation. Otherwise a common update path may keep the bug.

3. **Calendar writes through `apply_training_plan` are not addressed.**
   `apply_training_plan` calls `eventWriteParams` via `eventParamsFromPlanWorkout`. If `eventWriteParams` changes to accept serialization options, this path must be updated too; if it is intentionally out of scope, document why. It is also a planned-workout calendar write path and can carry `workout_doc`.

## Recommended plan adjustments

- Add a small tool-layer helper that selects options from `(profile, sport)`, e.g. normalize/allow only `POWER_HR_PACE`, `HR_POWER_PACE`, `PACE_HR_POWER`; unknown/missing order returns zero options.
- Keep `workoutdoc` independent of `intervals` types; pass primitive options such as `SerializeOptions{ExplicitZoneMetricSuffixes: true}` or a normalized workout order string.
- Update STATUS.md with the decisions for missing `workout_order`, `update_workout` without sport, and `apply_training_plan` coverage before starting Step 2.
