# Plan Review: TP-082 Step 2 — Add failing golden tests

## Verdict: needs changes before implementation

The Step 1 audit is useful, but the current Step 2 plan in `STATUS.md` is still only the generic checklist plus planned file locations. For this step, the test plan needs to be more concrete so the worker adds coverage that actually proves null-key stripping, preserves meaningful zero/false/empty values, and produces an expected failing suite before Step 3.

## Required plan changes

1. **Define the exact per-tool test cases before editing tests.**
   Add a Step 2 table to `STATUS.md` with, for each in-scope write tool:
   - test file/name to add or update;
   - sparse upstream fixture fields containing nulls;
   - fields that prove zero/false/empty-string preservation;
   - whether the tool has an `include_full` path and what raw-null assertion will be made;
   - whether the test is expected to pass today or fail today.

   The table should cover all audited write tools: `add_or_update_event`, `link_activity_to_event`, `add_activity_message`, `update_wellness`, `update_sport_settings`, `apply_training_plan`, `create_workout`, `update_workout`, `create_custom_item`, and `update_custom_item`.

2. **Make key-presence assertions explicit.**
   Several existing tests use patterns like `row["weight"] == nil` or `full["extra"] != nil`, which do not distinguish “key absent” from “key present with JSON null”. Step 2 must plan helper assertions, for example:
   - default/terse: `if _, ok := row["nullable_field"]; ok { ... }`
   - full/raw: `value, ok := full["nullable_field"]; if !ok || value != nil { ... }`
   - optional recursive helper for “no nulls anywhere in this terse object” if useful.

   Without this, the tests can pass while failing to prove the task’s core behavior.

3. **Identify the intentional failing cases.**
   The audit found the likely current bug in `create_custom_item` and `update_custom_item`: both hard-code full shaping and should preserve nulls by default today. Step 2 should explicitly expect those default custom-item null-stripping tests to fail before Step 3. If other new tests fail, record them as discoveries; if no tests fail, record why the Step 3 code change is still needed or revise the audit.

4. **Use fixtures that can actually prove zero/false/empty preservation.**
   Some response structs use `omitempty`, so a raw zero or empty string may be omitted before the shared shaper sees it. The plan should choose fields that survive the existing response shape, such as:
   - pointer-backed numeric/bool rows (`load_target: 0`, `distance_meters: 0`, `indoor: false`);
   - custom-item map fields (`index: 0`, `hide_script: false`, `description: ""`);
   - sport-settings echo fields that intentionally include false values (for example `zone_definitions_overwritten: false`).

   Avoid asserting preservation on non-pointer struct fields with `omitempty` unless the behavior is intentionally being changed in Step 3.

5. **Do not invent unsupported `include_full` arguments.**
   For tools that currently have no user-facing `include_full` (`update_sport_settings`, `apply_training_plan`, `create_workout`, `update_workout`, and custom-item writes), Step 2 should assert terse/default behavior only unless Step 3 intentionally adds new API surface. Full/raw-null preservation should be tested only where already supported (`add_or_update_event`, `link_activity_to_event`, `add_activity_message`, `update_wellness`) or explicitly documented as unavailable.

6. **Plan the verification command and status update.**
   The Step 2 plan should say which targeted command will be run after adding tests, e.g. `go test ./internal/tools -run 'Test(AddOrUpdateEvent|LinkActivityToEvent|AddActivityMessage|UpdateWellness|UpdateSportSettings|ApplyTrainingPlan|CreateWorkout|UpdateWorkout|CreateCustomItem|UpdateCustomItem)'`, and that `STATUS.md` will record the expected failing tests before moving to Step 3.

## Guidance for likely test placement

- Strengthen existing include-full tests for `link_activity_to_event` and `add_activity_message` to assert raw null key presence, not just nil lookup.
- Add or strengthen `update_wellness` tests so terse mode asserts null-key absence and `include_full` asserts raw null-key presence under `wellness.full` while preserving the existing nested `_meta` behavior.
- Add default sparse-null tests for both custom-item write tools; these should be the primary red tests today.
- For `apply_training_plan`, assert created event rows do not include `full` or null keys and still preserve meaningful pointer-backed zeros if present.

Once this concrete Step 2 test matrix is recorded, the plan should be ready to execute.
