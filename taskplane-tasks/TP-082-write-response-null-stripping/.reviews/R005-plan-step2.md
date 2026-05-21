# Plan Review: TP-082 Step 2 — Add failing golden tests

## Verdict: REVISE

The updated Step 2 plan is much more concrete and covers the full audited write-tool set. It also correctly calls out key-presence assertions and the likely custom-item red tests. However, a couple of planned assertions still rely on fields that are currently omitted before the shared response shaper ever sees them, so the expected pass/fail matrix is not reliable yet.

## Required changes

1. **Fix the `omitempty` false/empty-string cases in the matrix.**
   The plan says `update_sport_settings` should preserve `zone_definitions_overwritten: false` and expects that test to pass today, but `updateSportSettingsEcho.ZoneDefinitionsOverwritten` is tagged `json:"zone_definitions_overwritten,omitempty"`. `jsonenc` drops false values for `omitempty` fields before null stripping. Update the plan to either:
   - assert a false value that is actually emitted today, such as `_meta.zones_provided: false`; or
   - intentionally make `zone_definitions_overwritten:false` a red test and list it in the expected failing tests for Step 3.

2. **Do not assert `update_workout` empty-string preservation on `description` unless it is intentionally red.**
   `workoutTemplateRow.Description` is a plain string with `omitempty`, so `description:""` will be omitted before the shaper. The Step 2 matrix currently lists `description:"" when intentionally supplied` while expecting `update_workout` to pass today. Either remove that assertion and use pointer/map-backed fields (`indoor:false`, `distance_meters:0`) for the pass case, or mark it as an intentional failure and include the corresponding Step 3 change in scope.

3. **Make the “expected red tests before Step 3” list match the assertions.**
   The current plan names only `create_custom_item` and `update_custom_item` as expected failures. If the tests include either of the assertions above as written, `update_sport_settings` and/or `update_workout` may also fail. The matrix should explicitly say which failures are intended and why; unexpected failures can still be logged as discoveries after running the targeted command.

## Non-blocking guidance

- Keep the useful key-presence helper approach: terse assertions should check absence with `_, ok := row[key]`, and full/raw assertions should check present-with-nil separately.
- For event/workout zero values, prefer pointer-backed row fields (`load_target:0`, `distance_meters:0`, `indoor:false`) rather than non-pointer `omitempty` fields.
- The targeted `go test ./internal/tools -run ...` command is appropriate once the matrix is corrected.

After those matrix fixes, this Step 2 plan should be ready to execute.
