# Code Review: Step 2 — Implement and test metric suffix behavior

**Verdict:** Request changes.

## Findings

1. **Missing required direct write-path regression tests.** Step 2 changes the actual `add_or_update_event` and `update_workout` serialization paths (`internal/tools/add_or_update_event.go:94`, `internal/tools/update_workout.go:79`), but the new tool-layer coverage only exercises `create_workout` and `apply_training_plan`. The approved Step 2 plan explicitly required actual write-param assertions for `add_or_update_event`, `update_workout` with supplied `sport`, and the `update_workout` no-sport fallback. Please add those tests so these changed handlers cannot regress to zero/default serialization options unnoticed.

## Notes

- The core serializer boundary looks sound: default `Serialize` / `MergeDescription` behavior remains unchanged, while option-aware paths emit `Z* Power` when sport order context is known.
- I ran: `go test ./internal/workoutdoc ./internal/tools ./internal/intervals` — all passed.
