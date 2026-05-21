# Code Review — Step 1: Map the current assertion chain

**Verdict: request changes before moving to Step 2.**

The Step 1 notes are substantially more useful than the original plan, but I found a few factual gaps that will directly affect the refactor if they are used as the implementation checklist.

## Blocking findings

1. **`schemaCatalogClient` coverage is wrong for `ApplyTrainingPlanClient`.**
   - `STATUS.md:78` marks `ApplyTrainingPlanClient` as not covered by `schemaCatalogClient`.
   - That interface embeds `WorkoutLibraryClient`, `EventsClient`, and `EventWriterClient` (`internal/tools/apply_training_plan.go:27-32`). `schemaCatalogClient` implements all of those methods, so the current registry assertion succeeds and `apply_training_plan` is in the generated snapshots (`STATUS.md:114`).
   - This matters for Step 3: any schema fake/filter that follows the table literally could accidentally drop `apply_training_plan`, causing snapshot/catalog drift. Update the inventory to mark it as covered and explicitly call out that the coverage is structural via embedded interfaces rather than an explicit compile-time assertion.

2. **Toolchecks inventory omits `GenerateToolCatalog` / confusable-name checks.**
   - The notes discuss `GenerateSchemaSnapshots()` and `schema_stability.go`, but `internal/toolchecks/confusable_names.go:35-38` also constructs `tools.NewRegistryWithOptions(schemaCatalogClient{}, ...)`.
   - When `NewRegistryWithOptions` stops accepting interface-shaped fakes and/or `schemaCatalogClient` is collapsed, this caller must be migrated too. Otherwise Step 3 can leave `scripts/check_confusable_names.go` broken even if schema stability passes.
   - Add `GenerateToolCatalog` to the toolchecks migration plan, ideally sharing the same filtered catalog fixture used for schema snapshots so both checks keep identical membership.

3. **Registry fake call-site inventory misses the protocol dispatch test that needs real handler behavior.**
   - `STATUS.md:110` says MCP protocol tests have two real-registry fake call sites, but `internal/mcp/protocol_test.go` has the two `advancedProtocolClient` registrations at `:625` and `:643` plus `TestProtocolGetAthleteProfileDispatch` using `tools.NewRegistry(testProfileClient{...})` at `:751-760`.
   - The latter calls the actual `get_athlete_profile` handler and expects fake profile data. A no-network dummy `*intervals.Client` is not a drop-in migration unless the test is also changed to serve an HTTP fixture or avoid the real registry. Because `newGetAthleteProfileTool` is unexported to package `mcp`, this needs an explicit migration plan.

## Non-blocking notes

- The Step 1 status still says “In Progress” even though all Step 1 boxes are checked. Once the above corrections are made, mark Step 1 complete or move the current step to Step 2 so reviewers can tell whether implementation is intended to start.
- Consider making the `*intervals.Client` interface-satisfaction evidence reproducible in Step 4 (for example, a focused registry-level test or compile-time assertions in a test file). The current evidence was temporary, which is fine for mapping, but it is not a lasting regression guard.
