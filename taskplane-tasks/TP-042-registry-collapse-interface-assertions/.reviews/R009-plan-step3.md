# Plan Review — Step 3: Collapse `schemaCatalogClient`

**Verdict: approve.**

The Step 3 direction is sound as long as it follows the concrete migration already recorded in `STATUS.md`: use the no-network real `*intervals.Client` path plus the schema catalog allow-list/filter, and delete the now-dead `schemaCatalogClient` method fan-out. That preserves the existing 30-tool snapshot surface while removing the hand-maintained interface stub that this task targets.

## Implementation guardrails

- Do not introduce a new minimal interface fake for registry construction. After Step 2, `tools.NewRegistryWithOptions` requires `*intervals.Client`; using a fake would mean adding a new abstraction/adapter and would undercut the refactor. Use the existing dummy real client with the failing `schemaCatalogRoundTripper`.
- Keep the schema catalog allow-list (or an equivalent filtering registrar) in the toolchecks path. Registering an unrestricted real client would expand schema generation from the committed 30 snapshots to the full 38 production tools and break the byte-identical snapshot requirement.
- Delete the entire `schemaCatalogClient` block: type declaration, compile-time assertions, and all stub methods. It should not remain as unused compatibility baggage after the real-client helper is in place.
- Keep `GenerateSchemaSnapshots()` and `GenerateToolCatalog()` sharing the same catalog-generation helper so schema stability and confusable-name checks inspect the exact same tool set.
- Preserve the no-network guarantee: schema/catalog generation should still fail fast if a tool registration path attempts HTTP.
- Clean up imports after deleting the stub, then run gofmt/goimports.

## Verification expected for Step 3

- `grep -R "schemaCatalogClient" internal/toolchecks` returns no matches.
- `go test ./internal/toolchecks` passes.
- `go run ./scripts/check_schema_stability.go` passes with no snapshot rewrites.
- `go run ./scripts/check_confusable_names.go` passes.
- If any snapshot file changes, treat that as a regression unless there is a separate compatibility plan; this step should be wiring-only and byte-identical.
