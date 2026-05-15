# R011 code review — Step 3: Tool registry plumbing

Verdict: **REVISE**

I ran:

- `git diff a527c36..HEAD --name-only`
- `git diff a527c36..HEAD`
- `go test ./...`

The test suite passes, but I found blocking issues in the coach-mode plumbing.

## Findings

### 1. Activity-ID tools get `athlete_id` ACL checks but are not routed or roster-bound

Several athlete-scoped tools still call upstream object endpoints that do not include the resolved athlete in the path, while `resolvePathAthleteID` only rewrites paths shaped as `.../athlete/{configuredID}/...`.

Relevant code:

- `internal/intervals/client.go:150-162` only replaces a path segment after literal `"athlete"` when it equals `c.athleteID`.
- `internal/intervals/activity_details.go:20` calls `"activity", activityID`.
- `internal/intervals/activity_streams.go:62` calls `"activity", activityID, "streams"`.
- `internal/intervals/delete.go:58` calls `"activity", activityID`.

These tools are listed as athlete-scoped in `internal/toolcatalog`, so the MCP wrapper now accepts/strips `athlete_id` and applies ACLs, but the selected target athlete is not used by the intervals request. If a coach key can access activities outside the configured roster by direct activity ID, the LLM can operate on them by supplying an allowed roster `athlete_id` plus an out-of-roster `activity_id`. Even within the roster, the `athlete_id` argument does not actually constrain which athlete's activity is read/modified/deleted.

This violates the Step 3 routing requirement and the threat-model invariant that `athlete_id` cannot escape the configured roster. Please either route these endpoints through athlete-scoped upstream paths if intervals.icu supports them, or add an ownership check against the resolved target athlete before returning data or performing writes/deletes. Add regression tests for representative direct-object tools such as `get_activity_details`, `get_activity_streams`, `add_activity_message`, `link_activity_to_event`, and `delete_activity`.

### 2. `icuvisor_list_advanced_capabilities` coach filtering is not tied to the MCP coach gate

The actual registration gate lives in `internal/mcp.safeRegistrar`, but the advanced-capabilities payload is filtered earlier in `internal/tools` only if the caller remembered to pass `RegistryOptions.CatalogFilter`:

- `internal/tools/registry.go:221` builds `newListAdvancedCapabilitiesTool(filteredCatalog(collector.tools, r.catalogFilter), ...)`.
- `internal/app/app.go:298-306` wires that filter for the production app.
- `internal/mcp/protocol_test.go:829-839` also wires a test-only filter manually.

This means `mcp.NewServer` with a coach-enabled `Options.Config` and a default registry that lacks `CatalogFilter` will correctly hide coach-denied tools from `tools/list` via `safeRegistrar`, but `icuvisor_list_advanced_capabilities` can still describe denied tools because its payload was already constructed from the unfiltered collector catalog. The Step 3 requirement was that denied tools are absent from the actual exposed catalog and not leaked through capability discovery; this currently depends on duplicate app-side wiring rather than the authoritative MCP gate.

Please centralize the coach-filtered catalog source so the advanced tool cannot diverge from `safeRegistrar` decisions, or make `tools.NewRegistryWithOptions` receive enough coach config/evaluator state to apply the same filter by construction. Add a regression test that uses the real registry with coach-enabled `mcp.Options.Config` and no manually supplied `CatalogFilter`, then asserts `icuvisor_list_advanced_capabilities` does not mention coach-denied tools.

### 3. Coach-mode-off catalog/behavior changes need an explicit decision or gating

`safeRegistrar.prepareTool` unconditionally adds the public `athlete_id` schema property to every athlete-scoped tool, even when `ICUVISOR_COACH_MODE` is off:

- `internal/mcp/server.go:369-375` applies `schemaWithAthleteID` based only on `toolcatalog.IsAthleteScopedTool`.
- `internal/mcp/server.go:391-409` also accepts a matching `athlete_id` in non-coach mode and strips it before the strict tool decoder.

The task acceptance criteria say coach mode off/default should leave the catalog and behavior unchanged versus today. This implementation changes both the advertised input schemas and call behavior for the default mode. If the intended Step 3 product decision is now “all single-athlete tools also advertise `athlete_id`,” please record that as a deliberate acceptance-criteria amendment in `STATUS.md`/docs. Otherwise, gate the schema injection and wrapper acceptance to effective coach mode so the default catalog remains byte-for-byte compatible.

## Notes

- `go test ./...` passes.
- The resource bypass decision is documented and production `defaultStartServer` disables `icuvisor://athlete-profile` in coach mode, which addresses the app path for this step.
