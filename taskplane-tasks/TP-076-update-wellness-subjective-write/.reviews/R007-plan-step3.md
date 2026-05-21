# Plan Review: Step 3 fix the client / tool

Result: approved.

The Step 3 plan targets the right boundaries for the live-probe finding: `feel` is not an accepted upstream write field, so the client should reject it before HTTP, the MCP tool should reject it before invoking the writer, and the public schema/examples/field metadata should stop advertising it while read-side `feel` remains intact.

## Checks run

- Read `PROMPT.md` and `STATUS.md`.
- Inspected `internal/intervals/wellness.go`, `internal/intervals/wellness_test.go`, `internal/tools/update_wellness.go`, and `internal/tools/update_wellness_test.go`.
- Ran `go test ./internal/intervals ./internal/tools`; failures are the expected red tests from Step 2:
  - `TestUpdateWellnessRejectsUnsupportedFeelBeforeRequest`
  - `TestUpdateWellnessSchemaDocumentsRangesUnitsAndReadOnlyFields`
  - `TestUpdateWellnessRejectsUnsupportedFeelBeforeWrite`

## Notes for implementation

- Keep the `feel` rejection ahead of any network I/O. In the tool path, that means validation should happen before profile lookup/writer invocation; in the intervals client, `WriteWellnessParams{Feel: ...}` should fail before `doJSONBody`.
- If `Feel` is removed from the strict-decoded request struct, add a pre-decode raw-field check similar to the read-only field guard; otherwise `DecodeStrict` will produce a generic unknown-field error instead of the required public message.
- Reconcile the existing positive tests that still send `feel` by changing only the write input. It is still useful for response fixtures to contain `feel` to prove read-side shaping and scale metadata remain supported.
- Prefer a single constant/helper for the exact public error string (`field_not_writable: feel (not accepted by intervals.icu wellness write)`) to avoid drift between the intervals boundary and the tool validation path.
