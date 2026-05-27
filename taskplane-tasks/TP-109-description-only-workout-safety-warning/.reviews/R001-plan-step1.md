# Plan Review R001 — Step 1

Verdict: Approved.

The proposed contract is appropriately additive: a new optional `_meta.description_only_workout_warning` avoids overloading the existing `_meta.workout_doc_warning`, which has different semantics (uploaded DSL did not render/parse upstream). The warning copy is terse, actionable, and non-blocking, so it should not prevent legitimate prose-only strength/session updates.

Implementation notes to preserve the approved contract:

- For `add_or_update_event`, the planned trigger is acceptable: update-shaped write (`event_id` present), `category` equal to `WORKOUT` case-insensitively, `description` actually written, and no `workout_doc`.
- For `update_workout`, use field presence for the description trigger (`descriptionProvided`) rather than only `Description != nil`, because sparse update semantics distinguish omitted fields from supplied fields. This also keeps any explicit null/clear behavior covered by the warning if accepted by decoding.
- Treat "description-only" as "description supplied without `workout_doc`", even if other sparse metadata fields (name/tags/folder/etc.) are also updated.
- Add negative tests for create-shaped WORKOUT events, non-WORKOUT event updates, and writes that include `workout_doc`, so the warning remains targeted.
- Update output-schema descriptions if those are the public place where `_meta` response fields are documented; input schema snapshots likely only need changes if schema text/examples are changed.

No blockers for Step 2.
