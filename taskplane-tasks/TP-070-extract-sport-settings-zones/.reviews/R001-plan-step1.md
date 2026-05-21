# Plan Review — TP-070 Step 1

Verdict: approved with minor guardrails.

The Step 1 plan is appropriately scoped: verify/lock the safety-gate regression coverage before extracting the zones merge path. Current coverage already appears to satisfy the step:

- `TestUpdateSportSettingsSafeModeRejectsZonesBeforeWrite` covers zones supplied under `safety.ModeSafe`, checks the public gate message, and asserts no write call is made.
- `TestUpdateSportSettingsFullModeAppliesZonesAndResponseMeta` covers zones supplied under `safety.ModeFull`, asserts the zones are written, and checks response/meta fields.
- `TestUpdateSportSettingsOmittedZonesDoesNotWriteZones` also covers the important “no zones change/no gate needed” path mentioned in the task file scope.

I also ran the targeted tests:

```sh
go test ./internal/tools -run 'TestUpdateSportSettings(SafeModeRejectsZonesBeforeWrite|FullModeAppliesZonesAndResponseMeta|OmittedZonesDoesNotWriteZones)'
```

Result: pass.

Recommendations for executing Step 1:

1. Do not start the extraction in this step; keep Step 1 limited to confirming these regressions or adding only genuinely missing cases.
2. If you touch tests, keep them capability-based (`safety.NewCapability(...)`) rather than mutating `ICUVISOR_DELETE_MODE`; the production env parsing/gating is already covered in `internal/safety`.
3. Preserve the no-model-controlled-confirm invariant: no test or code should introduce a `confirm` argument or equivalent override.
4. When marking Step 1 done, note in `STATUS.md` that existing tests already lock safe-mode rejection, full-mode allowance, and omitted-zones behavior.

No blocking issues found in the plan.
