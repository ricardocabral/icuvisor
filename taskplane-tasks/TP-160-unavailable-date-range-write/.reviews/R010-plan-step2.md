# R010 Plan Review — Step 2

Verdict: APPROVE

The Step 2 plan is sufficient to proceed. The Step 1 contract is detailed enough to drive the implementation, and the hydrated Step 2 checklist covers the key integration surfaces: range validation/write behavior, schema/catalog registration, safety gates, snapshots, toolcatalog ACLs, and targeted test runs.

Non-blocking implementation notes:

- Registration in this repo is centered in `internal/tools/catalog.go` (`registryBaseTools` and `toolCatalogGroup`), not only `internal/tools/registry.go`; make sure the new tool is wired there.
- Add the tool to `internal/toolcatalog/catalog.go` as an athlete-scoped known tool before regenerating schema snapshots, otherwise coach-mode `athlete_id` injection and ACL validation will be wrong.
- Do not reuse `eventCreatePreflightFromEvents` blindly: it treats a matching `external_id` as a duplicate before comparing writable fields, while the accepted contract requires drifted same-external-id rows to be reported as conflicts and still create the requested marker.
- Expect to update schema snapshot count/fixtures and catalog-tier expectations, then run `go run ./scripts/snapshot_tool_schemas.go` plus `go test ./internal/tools ./internal/mcp ./internal/toolchecks`.

I also ran `go test ./internal/tools -run AddUnavailable`; it currently fails at compile time because the Step 1 tests reference the not-yet-implemented `newAddUnavailableDateRangeTool`, `addUnavailableDateRangeName`, and external-ID helper, which is expected for the start of Step 2.
