# Plan Review: TP-082 Step 1 — Audit write response shaping

## Verdict: needs changes before implementation

The task prompt is clear, but the current Step 1 plan in `STATUS.md` is only the generic checklist. For this task, the audit result is the control point for the rest of the work: it must define exactly which write tools are in scope, which are intentional exceptions, and which tests/code paths Step 2/3 will target. Please make the Step 1 plan concrete and record the audit matrix in `STATUS.md` before changing code.

## Required plan changes

1. **Audit from the registered catalog, not only the prompt file list.**
   `internal/tools/catalog.go` currently registers these `RequirementWrite` tools:
   - `add_or_update_event`
   - `link_activity_to_event`
   - `add_activity_message`
   - `update_wellness`
   - `update_sport_settings`
   - `apply_training_plan`
   - `create_workout`
   - `update_workout`
   - `create_custom_item`
   - `update_custom_item`

   The prompt's file scope omits `link_activity_to_event.go` and `add_activity_message.go`, but both can include raw upstream echo data when `include_full` is true. Step 1 should explicitly include them in the audit or record why they are out of scope.

2. **Record a per-tool response-shaping matrix in `STATUS.md`.**
   For each write tool, capture at least:
   - upstream echo type returned by the client, if any;
   - response builder used (`eventRow`, `wellnessRow`, `workoutToRow`, `shapeGetCustomItemByIDResponse`, bespoke confirmation, etc.);
   - `encodeShaped` include-full value (`args.IncludeFull`, `false`, or hard-coded `true`);
   - row collections passed to the shaper;
   - terse default behavior for null object keys;
   - full/debug behavior for preserving raw nulls;
   - planned Step 2 test location.

3. **Decide and document intentional exceptions before code.**
   Some current paths appear to be deliberately terse and may not support `include_full` at all (`create_workout`, `update_workout`, `apply_training_plan`, possibly `update_sport_settings`). If that remains the desired behavior, Step 1 must record them as intentional exceptions so Step 2 does not invent unsupported `include_full` behavior.

4. **Call out likely high-risk/currently divergent paths.**
   The audit should explicitly inspect these before test writing:
   - `create_custom_item` and `update_custom_item` call `encodeShaped(..., true, ...)` unconditionally and use the full custom-item read shape, so nulls are likely preserved by default. This looks directly relevant to the mission and should not be missed.
   - `update_wellness` shapes the inner wellness row first, then shapes the wrapper again; confirm this double-shaping is intentional and does not create misleading nested `_meta` behavior while testing null stripping.
   - `add_or_update_event` and `apply_training_plan` use `eventRow`; verify default raw `event.full` is absent and `include_full`/created rows behave as intended.
   - `link_activity_to_event` and `add_activity_message` are terse confirmations by default but preserve upstream `Raw` only under `include_full`; classify them explicitly.

5. **Define the evidence-gathering commands/files.**
   The plan should say the worker will inspect `RequirementWrite` registrations plus all `encodeShaped` calls in the write tools, not rely on memory. A good minimum is:
   - `grep -R "Requirement: RequirementWrite" internal/tools`
   - `grep -R "encodeShaped" internal/tools/{add_activity_message,add_or_update_event,link_activity_to_event,update_wellness,update_sport_settings,apply_training_plan,create_workout,update_workout,create_custom_item,update_custom_item}.go`
   - targeted reads of the shared row builders in `get_events.go`, `get_wellness_data.go`, `get_workout_library.go`, and `get_custom_item_by_id.go`.

## Non-blocking notes

- Step 1 can remain read-only, but it should update `STATUS.md` with the audit table and any exceptions before Step 2 starts.
- The later tests should preserve meaningful zero/false/empty-string values, so Step 1 should identify which tools can realistically emit those values in fixtures.
- If the worker chooses not to include delete tools, record that they are excluded because this task is about write echo payloads and the delete tools return confirmations rather than sparse upstream objects.
