# Plan Review: Step 1 — Define the typed structs

**Decision: needs refinement before implementation.**

The step direction is correct, but the current plan is too underspecified for the three audited sites. Please tighten the struct design before swapping bodies so the implementation does not either preserve `map[string]any` under a new name or accidentally change wire semantics.

## Required plan adjustments

1. **Separate upstream/write types from tool response types.**
   - Put the workout-library write body in `internal/intervals` (unexported is fine, e.g. `workoutLibraryWriteRequest`) because it is used by the HTTP client.
   - Keep shaped tool response summaries in `internal/tools` (unexported unless genuinely shared outside the package). `WorkoutDocSummary` / training-plan summary are presentation shapes, not intervals HTTP client contracts.

2. **Do not use blanket `omitempty` where sparse updates need explicit empty values.**
   The write request must preserve current update behavior:
   - `FolderIDSet: true` with an empty folder ID must still send `"folder_id":""` to move a workout to the top level.
   - `TagsSet: true` with an empty slice must still send `"tags":[]` to clear tags.
   - `DescriptionSet: true` must still reject nil but allow an empty string if that is current behavior.

   A typed sparse body should therefore use pointer fields (for example `FolderID *string 'json:"folder_id,omitempty"'`, `Tags *[]string 'json:"tags,omitempty"'`) rather than value fields with `omitempty` that would drop explicit zero values.

3. **Define the exact typed summary shapes.**
   Please spell out the structs and fields before coding:
   - Workout doc summary should cover the existing output keys: `present`, `step_count`, `name`, and `top_level_keys`.
   - Training plan summary should cover the existing output keys: `id`, `name`, `description`, `folder_id`, `type`, `category`, `child_count`, `workout_count`, and `top_level_keys`.

   Use pointers where `omitempty` must distinguish absent from zero, and keep JSON key names byte/shape-compatible with current fixtures.

4. **Decide how to handle `full` without reintroducing `map[string]any`.**
   The cited `Full map[string]any` fields are response bodies and are part of this task. If `full` remains an opaque upstream passthrough, plan to represent it as `json.RawMessage` (or another typed/opaque JSON value) and add a helper that marshals the preserved raw payload while keeping the visible JSON shape unchanged. Leaving `Full map[string]any` would miss the acceptance criteria.

5. **Be careful with `workoutdoc` reuse.**
   Reuse `internal/workoutdoc.WorkoutDoc` only where the data is actually in that canonical structured form. The current summary helper accepts upstream `map[string]any` and `[]any`; forcing all summaries through `workoutdoc.WorkoutDoc` could silently drop unknown keys and change `top_level_keys` / `step_count` behavior.

## Suggested Step 1 outcome

Before moving to Step 2, document or create these types:

- `internal/intervals`: unexported typed write request with pointer fields for sparse update semantics.
- `internal/tools`: unexported typed workout doc summary and typed training plan summary structs.
- `internal/tools`: an explicit opaque `full` representation strategy, preferably `json.RawMessage`, with a helper for converting preserved raw maps.

With those details added, the plan will align with the prompt and should keep the wire shape unchanged.
