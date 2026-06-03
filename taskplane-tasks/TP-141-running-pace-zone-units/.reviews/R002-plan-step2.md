# Plan Review — Step 2

Result: Approved with required clarifications.

## Required clarifications

- The new threshold-pace regression should exercise the missing `seconds_per_mile` input as a real cross-unit conversion, not only a same-unit echo. For example, use a Run sport setting whose upstream `PaceUnits` is `MINS_KM` and assert `seconds_per_mile` is converted to the expected seconds-per-km value in `WriteSportSettingsParams`, plus `_meta.pace_input_unit` / `_meta.pace_upstream_unit`.
- The pace-zone round-trip test should cover `kind: "pace"` under `ICUVISOR_DELETE_MODE=full`, asserting both the writer params and response echo preserve boundary values and zone names. The existing power-zone full-mode test is not sufficient evidence for Run pace-zone labels/boundaries.
- If schema wording is tightened, lock the wording in `TestUpdateSportSettingsSchemaDocumentsInputsAndZoneGate` (or equivalent), especially that pace boundaries are durations in seconds per the sport pace distance unit and not speed. Include km/mile examples if that is the wording change.

## Notes

- The planned targeted command `go test ./internal/tools ./internal/units` is appropriate if edits stay within the listed files. If any read-path wording moves into `internal/athleteprofile`, include that package in the targeted test command as well.
- CHANGELOG updates can remain for Step 4, unless Step 2 is committed independently with user-visible wording changes.
