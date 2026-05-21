# Code Review: Step 2 failing test

Result: approved.

The Step 2 tests now cover the intended regression boundary: `feel` is treated as an upstream-unsupported write field, rejected before any HTTP/write call, and no partial-success result is returned. The accepted seven-field fixture test also documents the probed-good subjective payload separately from the red unsupported-`feel` case.

## Checks run

- `git diff a8e30dd..HEAD --name-only`
- `git diff a8e30dd..HEAD`
- `go test ./internal/intervals ./internal/tools` — fails on the intended red assertions:
  - `TestUpdateWellnessRejectsUnsupportedFeelBeforeRequest`
  - `TestUpdateWellnessSchemaDocumentsRangesUnitsAndReadOnlyFields`
  - `TestUpdateWellnessRejectsUnsupportedFeelBeforeWrite`

## Notes

- The prior R005 concerns are addressed: schema `examples`/`input_examples` are now checked for `feel`, and the unsupported-`feel` tool test asserts an empty `Result`.
- Existing positive tests that still use `feel` will need to be updated during Step 3 along with the implementation fix.
