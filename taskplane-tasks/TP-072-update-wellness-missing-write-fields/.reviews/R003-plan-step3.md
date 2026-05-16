# Plan Review: Step 3

Status: Approved

No blocking issues found with the Step 3 test plan. The planned table-driven coverage plus a combined-fields case maps to the task acceptance criteria for `spO2`, `vo2max`, `abdomen`, `respiration`, and `menstrualPhase`.

Implementation notes:

- Cover both layers explicitly:
  - In `internal/tools/update_wellness_test.go`, assert the handler maps each new argument into `intervals.WriteWellnessParams` and that top-level `_meta.fields_updated` includes the exact public field name.
  - In `internal/intervals/wellness_test.go` (or an equivalent existing test-server harness), assert the outbound PUT JSON body uses the exact upstream keys: `spO2`, `vo2max`, `abdomen`, `respiration`, and `menstrualPhase`. The case-sensitive `spO2` key is the issue-closing path.
- Keep the individual cases table-driven and assert sparse semantics where practical: a single-field update should not accidentally include omitted new fields.
- Validation coverage should include all new validations, not only the examples: `spO2 > 100` (and preferably `< 0`), negative `vo2max`, negative `abdomen`, negative `respiration`, and empty/whitespace-only `menstrualPhase`. Also assert validation failures do not call the writer.
- The combined-fields test should write all five in one request and assert both the request mapping/body and `fields_updated` contain all five exact names. Because `fields_updated` is sorted, compare as a set or to sorted expected values rather than relying on input order.
- Preserve existing strict-decoder/read-only tests; this step should add coverage without loosening unknown-field behavior.
