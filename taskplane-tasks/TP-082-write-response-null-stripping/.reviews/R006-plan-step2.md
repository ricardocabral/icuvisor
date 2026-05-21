# Plan Review: TP-082 Step 2 — Add failing golden tests

## Verdict: APPROVE

The Step 2 plan is now concrete enough to execute. It covers all audited write tools, identifies the expected red cases, uses key-presence assertions to distinguish absent keys from present JSON nulls, and avoids adding unsupported `include_full` API surface.

## What I checked

- Read `PROMPT.md` and the current `STATUS.md` Step 2 plan.
- Re-checked the prior R004/R005 objections against the updated matrix.
- Spot-checked the relevant response structs/builders:
  - `updateSportSettingsEcho.ZoneDefinitionsOverwritten` is still `omitempty`, and the plan now uses `_meta.zones_provided:false` instead.
  - `workoutTemplateRow.Description` is still a plain `omitempty` string, and the plan no longer requires empty-description preservation for the passing `update_workout` case.
  - `create_custom_item` and `update_custom_item` still hard-code `encodeShaped(..., true, ...)`, so those default null-stripping tests are correctly listed as the intentional red tests before Step 3.
  - `update_wellness` still double-shapes the inner row, so preserving the existing nested `_meta` behavior in tests is the right constraint.

## Notes for implementation

- Keep the planned presence checks strict:
  - terse/default: fail if the nullable key exists at all;
  - full/raw: require the key to exist and its value to be `nil`.
- For zero/false/empty preservation, stick to the matrix’s pointer-backed or map-backed fields. Do not reintroduce assertions on non-pointer `omitempty` fields unless intentionally making them red and documenting a Step 3 behavior change.
- The expected failing set before Step 3 should remain limited to the custom-item write default null-stripping tests. If additional tests fail, record them as discoveries before broadening the fix.
- After adding tests, run the targeted `go test ./internal/tools -run ...` command from the plan and record the red test names in `STATUS.md` before moving to Step 3.
