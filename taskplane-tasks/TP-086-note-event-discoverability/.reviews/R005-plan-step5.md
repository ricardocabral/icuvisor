# Review R005 — Step 5 plan

Decision: **Approved with required clarifications.**

The Step 5 checklist is the right final verification gate for this task: it covers targeted tests, the full unit suite, build, lint, and failure documentation. Because earlier steps changed `internal/tools/add_or_update_event.go`, the generated schema snapshot, public Hugo content, and `CHANGELOG.md`, execute the gate with explicit commands and record exact results rather than marking the broad checkboxes generically.

## Required clarifications for execution

1. **Run the targeted checks for the edited assistant-facing surface.**
   At minimum, rerun the focused checks that prove the NOTE examples remain valid and the committed schema snapshot matches the live registry:

   ```bash
   go test ./internal/tools -run 'Test(RegisteredV03WriteToolsExposeInputExamples|ComplexWriteToolInputExamplesValidateAgainstSchema|AddOrUpdateEventStandardCategoryExamplesUseDocumentedValues|AddOrUpdateEventRegistrationMetadata)'
   go run ./scripts/check_schema_stability.go
   ```

   These checks are still useful even if Step 3 ran them, because Step 5 is the final verification gate after Step 4 delivery edits.

2. **Run the full required project gates and capture command outcomes.**
   Execute the Step 5 commands exactly unless an environment dependency is unavailable:

   ```bash
   make test
   make build
   make lint
   ```

   If any command fails, fix task-related failures before proceeding. If a failure is clearly pre-existing or environmental (for example, `golangci-lint` is not installed), document the command, failing package/tool, and concise error summary in `STATUS.md` under Step 5 rather than leaving the checkbox ambiguous.

3. **Preserve docs-build evidence for the final gate.**
   The Step 5 checklist does not name the site build, but the task completion criteria require docs/site verification and this task changed `web/content/reference/tools.md`. Either rerun:

   ```bash
   make web-build
   ```

   or explicitly reference the successful Step 3 `make web-build` result in `STATUS.md` and confirm no later `web/` changes occurred. If Hugo is unavailable, record the attempted command and environment limitation.

4. **Check for generated-artifact drift before marking complete.**
   After the commands, run a clean-tree sanity check such as:

   ```bash
   git status --short
   git diff --check
   ```

   Do not leave uncommitted drift in `internal/tools/schema_snapshot/add_or_update_event.json`, `web/data/tools.json`, or build output. If verification commands modify generated files, explain the drift and commit the intended artifacts at the next step boundary.

5. **Update `STATUS.md` with evidence, not just checkmarks.**
   Add the exact commands/results to the execution log or Step 5 notes, including any pre-existing unrelated failures. Keep the scope narrow: no new `add_note` tool, no event-write behavior changes, and no unrelated refactors during the verification step.

With these clarifications, the Step 5 plan is sufficient to validate the NOTE-event discoverability changes before documentation/delivery finalization.
