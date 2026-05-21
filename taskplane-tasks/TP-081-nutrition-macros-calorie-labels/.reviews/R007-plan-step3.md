# R007 Plan Review — Step 3: Regression tests

**Verdict:** REVISE.

The current Step 3 plan in `STATUS.md` only repeats the high-level prompt checklist. Before coding, please expand it into a concrete regression matrix that locks down the public response contract introduced in Steps 1–2. This task is specifically about avoiding ambiguous calorie/macros keys, so the tests need to name the tools, keys, metadata, and absence cases they will assert.

## Required plan changes

1. **Cover both activity tools explicitly.**
   Add planned tests for `get_activities` and `get_activity_details` that assert:
   - upstream activity `calories` is emitted as top-level `calories_burned`;
   - no ambiguous top-level `calories`, `calories_intake`, `calories_total`, or macro keys are emitted for activities;
   - `_meta.field_semantics.calories_burned` is present when at least one returned row/detail has `calories_burned`;
   - absent activity calories are stripped and do not create misleading field semantics;
   - present zero calories are preserved as `calories_burned: 0`.

2. **Cover wellness nutrition translation explicitly.**
   Add planned tests for `get_wellness_data` using the nutrition-bearing wellness fixture (`manual_only.json`) and an absence fixture (`custom_fields.json` or equivalent) that assert:
   - `kcalConsumed`/`carbohydrates`/`protein`/`fatTotal` become `calories_intake`/`carbs_g`/`protein_g`/`fat_g` with exact fixture values;
   - legacy upstream nutrition keys are not emitted as top-level terse row fields;
   - row `_meta.field_semantics` contains only the nutrition fields that are actually present;
   - absent nutrition fields are stripped and do not create a `field_semantics` entry.

3. **Include `include_full` behavior in the test matrix.**
   The plan should state that raw upstream nutrition names remain available under `full` only when `include_full: true`, while normalized top-level keys remain disambiguated. This is a key part of the approved Step 2 behavior and should be regression-tested.

4. **Specify the targeted test command.**
   The plan should name the targeted command(s), e.g. `go test ./internal/tools` (and any narrower package command if used). Step 4/5 can own the full quality gate, but Step 3 should still record the targeted read-tool verification.

Once these points are added to `STATUS.md`, the Step 3 plan should be ready to implement.
