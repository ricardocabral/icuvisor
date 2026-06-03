# Plan Review: Step 1 — Design the sport-aware suffix boundary

**Verdict:** Approved for implementation.

## Why this plan is ready

- The revised plan resolves the prior blockers: it identifies `workout_order` as the source field, covers `update_workout` when no sport is supplied, and brings `apply_training_plan` into scope.
- Keeping `workoutdoc.Serialize` / `MergeDescription` unchanged while adding an options-aware path preserves existing no-context behavior and avoids breaking current WorkoutDoc golden coverage.
- Centralizing suffix rendering in `internal/workoutdoc` is the right boundary; call sites should pass options rather than rewrite serialized DSL strings.
- The chosen explicit behavior for known orders (`POWER_HR_PACE`, `HR_POWER_PACE`, `PACE_HR_POWER`) is clear: emit family-specific zone suffixes for power, HR, and pace zones whenever supported sport context is known.

## Implementation notes to carry into Step 2

- Add the `intervals.SportSettings.WorkoutOrder` decode field and tests even though `internal/intervals/*` was not in the original file-scope list; this is necessary for the selected boundary.
- Keep `workoutdoc` independent of `intervals` types. Pass a small option/primitive into the serializer, with unknown or missing `workout_order` returning zero/default options.
- For `update_workout`, make the fallback explicit in tests or docs: no supplied `sport` means existing bare serialization remains.
- Include `apply_training_plan` tests or assertions that the plan template sport/type is used to derive the same serialization options as direct planned-workout writes.

No additional plan changes are required before starting Step 2.
