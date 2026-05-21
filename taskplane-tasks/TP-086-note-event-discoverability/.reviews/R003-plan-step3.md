# Review R003 — Step 3 plan

Decision: **Approved with required clarifications.**

The Step 3 checklist is the right validation gate after Step 2 changed both the assistant-facing `add_or_update_event` schema examples and public Hugo content. However, the plan should be executed with explicit regeneration/verification commands so the generated schema snapshots do not remain stale.

## Required clarifications for execution

1. **Regenerate schema snapshots, not just website tool data.**
   Step 2 changed `addOrUpdateEventInputExamples()`, and the committed snapshot is currently stale. A temp generation shows drift in `internal/tools/schema_snapshot/add_or_update_event.json` for the new NOTE examples. Run:

   ```bash
   go run ./scripts/snapshot_tool_schemas.go
   ```

   The expected committed generated diff should be limited to `internal/tools/schema_snapshot/add_or_update_event.json` plus task status/review bookkeeping. If other schema snapshots change, stop and explain why before committing.

2. **Treat `make docs-tools` as optional/no-op for the current edits.**
   Step 2 did not change top-level tool names/descriptions/catalog metadata, and `cmd/gendocs` emits `web/data/tools.json` without `input_examples`. A temp run of `go run ./cmd/gendocs --out <tmp>` produced no diff against `web/data/tools.json`. It is fine to run `make docs-tools` as a sanity check, but do not commit `web/data/tools.json` unless it actually changes for a justified reason.

3. **Run the Hugo build because `web/content/reference/tools.md` changed.**
   Run:

   ```bash
   make web-build
   ```

   If Hugo is unavailable in the worker environment, record that explicitly in `STATUS.md` with the command attempted and the failure. Do not silently mark the docs build complete.

4. **Run targeted tool metadata/schema checks.**
   At minimum, run the example/schema tests that cover the edited surface:

   ```bash
   go test ./internal/tools -run 'Test(RegisteredV03WriteToolsExposeInputExamples|ComplexWriteToolInputExamplesValidateAgainstSchema|AddOrUpdateEventStandardCategoryExamplesUseDocumentedValues|AddOrUpdateEventRegistrationMetadata)'
   go run ./scripts/check_schema_stability.go
   ```

   The first command verifies the new examples are valid and use documented event categories; the second verifies checked-in snapshots match the live registry after regeneration. Full `make test`/`make build`/`make lint` can remain in later delivery/verification steps unless the worker chooses to run them early.

5. **Keep the Step 3 commit focused.**
   This step should not add a separate `add_note` tool, broaden event-write behavior, or move the changelog work forward unless the task runner intentionally combines steps. Update `STATUS.md` with the exact commands/results and any environment limitations.

With those clarifications, the Step 3 plan should satisfy the task requirement that generated docs/schema artifacts remain in sync after expanding NOTE-event discoverability.
