# Plan Review R002 — Step 2

Verdict: Approved.

The Step 2 plan is consistent with the approved Step 1 contract and covers the required implementation surfaces: response `_meta` additions, both affected write tools, warning-present/absent tests, and targeted `go test ./internal/tools` verification.

Implementation guardrails to preserve the contract:

- Add the field as optional JSON metadata on both response meta structs, e.g. `_meta.description_only_workout_warning`, and keep it omitted when empty.
- For `add_or_update_event`, trigger only when `event_id` is present, `category` is `WORKOUT` case-insensitively, `description` is actually written (`Description != nil`), and no `workout_doc` is supplied.
- For `update_workout`, trigger on field presence: `descriptionProvided && !workoutDocProvided`, not just `Description != nil`, so explicit null/clear semantics remain covered if accepted by decoding.
- Do not interpret “description-only” as “no other fields changed”; the warning should still appear if description is supplied without `workout_doc` alongside name/tags/folder/sport updates.
- Include negative tests for create-shaped WORKOUT events, non-WORKOUT event updates, and writes that include `workout_doc` for both tool families where applicable.
- Update output schema descriptions to mention the new `_meta` field. Input schema snapshots only need regeneration if input schema text/examples changed; current snapshots are input-schema-only.

No blockers for implementation.
