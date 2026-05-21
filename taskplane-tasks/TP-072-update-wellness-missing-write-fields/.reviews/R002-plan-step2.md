# Plan Review: Step 2

Status: Approved

No blocking issues found with the Step 2 plan.

I verified the current tool path in `internal/tools/update_wellness.go`: strict argument decoding depends on `updateWellnessRequest`, the public MCP schema comes from `updateWellnessInputSchema`, writes are mapped in `wellnessWriteParams`, and the top-level `_meta.fields_updated` echo is built by `updateWellnessFieldsUpdated`. The planned changes cover each required integration point for `spO2`, `vo2max`, `abdomen`, `respiration`, and `menstrualPhase`.

Implementation notes:

- Use Go field names consistent with the client/read structs: `SpO2 *float64`, `VO2Max *float64`, `Abdomen *float64`, `Respiration *float64`, and `MenstrualPhase *string`; keep JSON/schema keys exactly `spO2`, `vo2max`, `abdomen`, `respiration`, and `menstrualPhase`.
- For the schema, include both `minimum` and `maximum` for `spO2` (`0..100`), and `minimum: 0` for the other numeric additions. Keep `menstrualPhase` as a string unless an upstream enum is verified from authoritative docs.
- For `menstrualPhase` validation, reject nil only when absent is acceptable; when present, reject strings that are empty after `strings.TrimSpace`, so whitespace-only values do not count as valid writes.
- Add the new fields to `updateWellnessFieldsUpdated`; also consider updating the category slices (`updateWellnessMeasurementFields` / `updateWellnessFreeTextFields`) so the helper capacity and field grouping remain self-documenting, even though they are currently only used for capacity.
- Preserve the existing strict decoder and read-only field rejection behavior; this task should only expand the writable allow-list, not loosen unknown-field handling.
