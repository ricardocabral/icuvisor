# Plan Review: Step 2 — Swap request/response bodies

**Decision: approved with implementation notes.**

The Step 2 checklist is aligned with the prompt and the completed Step 1 direction: swap the actual request/response body fields to the typed structs / `json.RawMessage`, update round-trip coverage, and run the normal build/test/lint checks.

## Notes to carry into implementation

- Keep the scope narrow. The remaining `map[string]any` occurrences in schema literals and internal raw-preservation helpers are exempt for this task. Do not expand this into refactoring `Raw map[string]any`, `TrainingPlan.TrainingPlan any`, or schema generation.
- For `internal/intervals/workout_library.go`, preserve sparse update semantics exactly:
  - create requires non-empty trimmed `name` and `type`;
  - update sends `folder_id:""` when `FolderIDSet` is true and the trimmed value is empty;
  - update sends `tags:[]` when `TagsSet` is true with an empty slice;
  - `DescriptionSet` with nil should still be rejected, but an explicit empty string should be allowed.
- For tool responses, keep `full` as explicit opaque JSON passthrough (`json.RawMessage` or equivalent), not as a `map[string]any` field. Default responses should remain terse and omit `full`; summaries should stay typed with the existing JSON keys.
- Add or update tests for both affected tools. There does not appear to be existing dedicated `get_training_plan` tool coverage, so Step 2 should add it rather than only relying on interval-client tests.
- The round-trip tests should prove the wire JSON is unchanged. If comparing bytes, compare compact/canonical JSON so map key ordering/whitespace does not make the test brittle; otherwise use semantic JSON equality plus targeted assertions for fields that must be omitted or preserved.
- Update `CHANGELOG.md` and `STATUS.md` before finishing the step.
- Run the requested checks (`make build`, `make test`, `make test-race`, `make lint`) or record any checks not run with the reason.

With those constraints, proceed with Step 2.
