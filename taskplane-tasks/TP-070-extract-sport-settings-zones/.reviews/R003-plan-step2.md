# Plan Review — TP-070 Step 2

Verdict: approved with guardrails.

The Step 2 plan matches the task scope: extract only the `update_sport_settings` zones overwrite path into `internal/tools/update_sport_settings_zones.go`, update the handler to delegate to it, and split the focused zone tests into `update_sport_settings_zones_test.go`. I reviewed the current implementation and the Step 1 regression tests; the existing safe/full/omitted-zones cases are a good safety net for this extraction.

No blocking issues found in the plan.

## Guardrails for the extraction

1. Preserve the current gate source and timing. The helper should use the existing `safety.Capability` passed into `newUpdateSportSettingsTool`; do not read `ICUVISOR_DELETE_MODE` or `os.Getenv` from the helper, and do not add any model-controlled `confirm`/override argument.
2. Preserve current behavior ordering:
   - strict decode and zone validation still happen before the gate;
   - safe mode with valid `zones` still returns the public `zoneOverwriteGateMessage` and performs no profile fetch/write;
   - omitted `zones` continues to avoid the delete gate and sends no zone fields upstream.
3. Keep the extraction narrow. Move zone-specific helpers/copying/gate code only; do not extract FTP/HR/pace conversion, result shaping, schemas, or unrelated validation in this task.
4. Keep zones copying defensive. The extracted helper should continue cloning slices (`Boundaries`, `Names`) rather than aliasing request data into `intervals.WriteSportSettingsParams` or response echoes.
5. Preserve public API/schema/wire format. Tool name, input schema, output shape, `_meta.zones_provided`, `zone_definitions_overwritten`, field names, and the gate error wording should not change.
6. Mirror the test split without weakening coverage: move the existing zone-focused tests to the new `_test.go` file and keep threshold/pace/schema tests in the main test file. Add helper-level table tests only if they clarify the no-zones/safe/full cases without duplicating handler coverage excessively.
7. Update `STATUS.md` and `CHANGELOG.md` under `[Unreleased]` as required by the task.

## Suggested verification after extraction

Run at minimum:

```sh
go test ./internal/tools -run 'TestUpdateSportSettings(SafeModeRejectsZonesBeforeWrite|FullModeAppliesZonesAndResponseMeta|OmittedZonesDoesNotWriteZones)$'
go test ./internal/tools
```

Then leave the broader `make build`, `make test`, `make test-race`, `make lint`, adversarial safety tests, and schema snapshot check for Step 3 as planned.
