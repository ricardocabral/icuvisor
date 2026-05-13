# Code Review: TP-020 Step 1 (`add_or_update_event`)

**Verdict: Changes requested**

I reviewed the full diff from `a4ff0a370fa81970ec14bcb8af141bb860aac96c^..HEAD` and read the changed implementation/tests. I also ran:

- `go test ./internal/intervals ./internal/tools` — pass

## Findings

### 1. Planned target fields are written/read as completed metric fields

**Severity: P1**

`target_load` and the planned metric inputs are currently serialized to `icu_training_load`, `distance`, `moving_time`, and `elapsed_time` in the write payload (`internal/intervals/events.go:147-156`). The repo's own upstream evidence for event compliance distinguishes planned targets from completed values: `testdata/extended-metrics/events-compliance.json:6-7` has `load_target` as the planned target and `icu_training_load` as the completed load. `docs/upstream-gaps/periodization-parameters.md` also calls out per-event `load_target`, `time_target`, and `distance_target` fields.

As implemented, a caller using the public `target_load` argument may not set the planned target used by compliance scoring, and the response path (`Event.TrainingLoad` / `getEventsRow.icu_training_load` at `internal/intervals/events.go:53` and `internal/tools/get_events.go:55`) cannot surface `load_target` even if upstream returns it. This breaks the Step 1 acceptance item for `target_load` / planned metrics and will make later compliance/pairing validation misleading.

Please add typed planned target fields (at least `load_target`, and likely `distance_target` / `time_target` where supported), map the tool's planned inputs to those upstream keys, and extend read/write shaping/tests so the confirmation can show the planned target distinctly from completed `icu_training_load`.

### 2. The new tool is absent from schema/catalog snapshot generation

**Severity: P2**

`add_or_update_event` is conditionally registered only when the registry client implements `EventWriterClient` (`internal/tools/registry.go:110-113`), but the catalog/snapshot generator's `schemaCatalogClient` does not implement that interface (`internal/toolchecks/schema_stability.go:275-290`). As a result, `go run ./scripts/snapshot_tool_schemas.go` does not generate an `internal/tools/schema_snapshot/add_or_update_event.json`, and the confusable-name/schema stability checks never see this new public tool.

That leaves the new write API outside the catalog guardrails added for this project, despite the task and prior plan review calling for schema/catalog updates. Please implement `AddOrUpdateEvent` on `schemaCatalogClient`, add an interface assertion for `tools.EventWriterClient`, commit the generated snapshot, and add/update a test that would fail if a newly registered write tool is missing from the generated catalog.

### 3. User-facing catalog/changelog were not updated

**Severity: P2**

The task's documentation section requires updating the README catalog and `CHANGELOG.md` when adding the tool, but the diff leaves the README tool list at only the read tools (`README.md:49-56`) and the `[Unreleased]` changelog has no `add_or_update_event` entry (`CHANGELOG.md:10-22`).

Because this is a newly exposed MCP tool, users and release notes need to know it exists and that it is a non-destructive write gated by write capability. Please add the README catalog entry and an `[Unreleased]` changelog bullet before marking Step 1 complete.
