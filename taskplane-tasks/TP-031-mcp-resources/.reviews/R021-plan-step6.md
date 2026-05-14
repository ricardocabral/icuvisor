# Plan Review R021 — Step 6: Trim inline tool descriptions

Verdict: REVISE

I read `PROMPT.md` and `STATUS.md`. The Step 6 section in `STATUS.md` only says "In Progress" and does not contain an implementation plan to review. Given this step is specifically about shrinking the model-facing catalog while keeping TP-015 guards green, the plan needs to be explicit before implementation.

Blocking items the Step 6 plan should add:

1. **Inventory the exact description surfaces to change.** At minimum, call out the current long custom-item inline prose in `internal/tools/get_custom_item_by_id.go` and replace it with a terse first sentence plus `see icuvisor://custom-item-schemas`. Also account for stale response/test wording such as `_meta.schema_documentation = "inline_v0.2_tool_description; moves_to_resource_v0.4"`. If the plan changes JSON Schema property descriptions in `create_custom_item`, `update_custom_item`, `add_or_update_event`, `create_workout`, or `update_workout`, it must say so explicitly because those descriptions are snapshot-guarded.

2. **Preserve tool-selection and safety wording.** Trimming should not remove distinguishing first sentences, non-destructive/delete-mode language, `include_full` guidance, free-text-vs-`workout_doc` mutual exclusion, or live custom-item validation behavior. Resource pointers should supplement terse operational guidance, not replace it.

3. **Define how TP-015 guards stay green.** The plan should distinguish top-level `Tool.Description` edits from input-schema description edits. Top-level edits need `go run ./scripts/check_confusable_names.go`; input-schema edits also require regenerating/checking `internal/tools/schema_snapshot/*.json` and running `go run ./scripts/check_schema_stability.go` in the same way CI does. Avoid broad schema-description edits unless they are required for the resource move.

4. **README and changelog updates.** Step 6's checklist includes README documentation for all four resource URIs; the task-level documentation requirements also require `CHANGELOG.md`. The plan should add a concise README MCP Resources section covering `icuvisor://workout-syntax`, `icuvisor://event-categories`, `icuvisor://custom-item-schemas`, and `icuvisor://athlete-profile`, and an `[Unreleased]` changelog bullet.

5. **Tests to update.** Name the tests/snapshots that will change, especially `internal/tools/get_custom_items_test.go` for the custom-item resource note/meta expectations, any workout/event schema snapshot tests if schema descriptions change, and the confusability/schema-stability checks.

Once `STATUS.md` has a concrete Step 6 plan covering the above scope and verification, this should be straightforward to approve.
