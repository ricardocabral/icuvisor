# Plan Review — Step 2: Add repeat syntax regressions

Verdict: APPROVE with implementation notes.

The Step 2 plan covers the task requirements: explicit serializer regressions for bare/described repeat headers, malformed repeat-like parse/validation coverage, at least one write-tool path regression, and targeted package tests.

Please tighten these points during implementation:

- For malformed `- 3x`, include a child step in the test input (for example `- 3x\n  - 1m 50%`) so the failure proves the header form is rejected, not merely that an empty repeat block is invalid. Current parsing can otherwise conflate the cases.
- For malformed `-3 x`, be explicit about expected layer behavior. `Parse` should reject it as DSL; `ValidateDescription`/`validate_workout` should surface a diagnostic if it is being treated as an attempted structured repeat, rather than silently passing it as prose.
- Use `internal/tools/add_or_update_event_test.go` for the write-tool regression if possible, because Step 1 found that create/validate paths already exercise repeat fixtures while `add_or_update_event` does not.
- Keep changes focused on regression behavior; avoid broad grammar/documentation changes in Step 2 unless the new tests expose a necessary parser/validator fix. `CHANGELOG.md` belongs in Step 4 per the task plan.

Run and record `go test ./internal/workoutdoc ./internal/tools` before marking Step 2 complete.
