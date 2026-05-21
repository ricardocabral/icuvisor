# Review R006 — Plan review for Step 2: `_meta` injector

Verdict: **REVISE**

The revised Step 2 checklist in `STATUS.md` addresses most of R005: it now calls out runtime metadata wiring, authoritative response-owned keys, synchronization for first-seen/current snapshots, the per-process fallback caveat, and a testable message seam. That is a good direction.

Two plan gaps remain before implementation.

## Required revisions

1. **Broaden the direct-response audit beyond the two named tools.**

   The plan currently says to convert direct JSON response paths for:
   - `internal/tools/list_advanced_capabilities.go`
   - `internal/tools/update_sport_settings.go`

   But `internal/tools/update_wellness.go` also returns a root wrapper directly after shaping only the nested `wellness` row (`shapeUpdateWellnessResponse` then `json.Marshal(payload)` / `StructuredContent: payload` around lines 125-133). Its root `_meta` is not passed through `response.Shape`, so Step 2 would not add top-level `_meta.catalog_hash` or schema-change fields there.

   Revise the plan to audit all `StructuredContent: payload` / `StructuredContent: response` / direct `json.Marshal(...)` tool paths, and explicitly include `update_wellness` unless you choose a central MCP-boundary injection design that covers all tool results consistently.

2. **Clarify no-server/default catalog-hash behavior for direct tool tests and helpers.**

   The plan says `internal/mcp.NewServer` will set runtime catalog metadata after computing `Server.CatalogHash()`, with test reset/set hooks. It should also state what `response.Shape` emits when no `mcp.Server` has been constructed in that process, which is how many `internal/tools` unit tests call handlers.

   The field should still be present and deterministic in those tests (for example via a package default fixed value or a helper used by tool-test setup), rather than omitted or left to an empty value accidentally. This is needed to satisfy “every tool response carries `_meta.catalog_hash`” and to avoid destabilizing existing JSON assertions.

Once those two details are added to `STATUS.md`, the Step 2 implementation plan is specific enough to proceed.
