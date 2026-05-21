# Review R005 — Plan review for Step 2: `_meta` injector

Verdict: **Changes requested before implementation.**

The Step 2 checklist in `STATUS.md` has the right high-level goals, but it is not yet specific enough to avoid a partial implementation. The main risk is that simply adding fields inside `response.Shape` will not, by itself, get the runtime catalog hash into every actual MCP tool response.

## Required plan revisions

1. **Specify the runtime metadata wiring path.**
   Step 1 computes `Server.CatalogHash()` only after tool registration. The Step 2 plan needs to say how that value reaches the response shaper without making the catalog hash part of the tool catalog itself.
   - Acceptable approach: add concurrency-safe runtime metadata in `internal/response` and have `internal/mcp.NewServer` (or `defaultStartServer` immediately after `NewServer`) set the current `{version, catalog_hash}` once the hash is known.
   - Include a test/reset hook so package tests do not leak first-seen state into each other.
   - Keep the hash out of tool descriptions and schemas; otherwise the hash would perturb the catalog it is supposed to describe.

2. **Audit response paths that bypass `response.Shape`.**
   Most tools eventually call `response.Shape` via `encodeShaped`, but at least these current handlers marshal their payloads directly:
   - `internal/tools/list_advanced_capabilities.go`
   - `internal/tools/update_sport_settings.go`

   If Step 2 only changes `addCommonMeta`/`response.Shape`, those tool responses will not carry `_meta.catalog_hash` or schema-change fields, violating the acceptance criterion that every tool response carries the new metadata. The plan should explicitly convert these paths to the shared shaper (or choose a central MCP-boundary injection design that updates both `StructuredContent` and JSON text content consistently).

3. **Define the first-seen state model, not just the fields.**
   The plan should name the stored values and synchronization strategy:
   - current snapshot: normalized `current_version` and `catalog_hash`
   - first-seen snapshot: normalized `previous_version` and `previous_catalog_hash`, keyed by session when available, otherwise by a documented process-level fallback key
   - one atomic/mutex-protected read-modify-write so concurrent tool calls cannot race or produce mismatched version/hash pairs

   The per-process fallback caveat is acceptable because the SDK session handle is not available at the response shaper boundary, but it needs a code note/docstring explaining that this only supports simulated in-process divergence and future session plumbing, not real cross-process restart detection.

4. **Make common metadata authoritative.**
   `addCommonMeta` currently preserves caller-provided `_meta` except for `units`. The Step 2 plan should state that response-owned keys are overwritten/removed from caller metadata before insertion: `server_version`, `catalog_hash`, `schema_changed`, `schema_change_message`, `previous_version`, `current_version`, and `previous_catalog_hash`. Otherwise a tool-specific DTO could accidentally spoof stale schema metadata.

5. **Pin the message/template seam.**
   The plan mentions a templated message, but should include the intended function/seam, e.g. an unexported `schemaChangeMessage(previousVersion, currentVersion string) string`. Tests can then assert the exact user-facing sentence required by the prompt.

6. **Clarify default/fixed catalog hash behavior in tests.**
   Since adding `_meta.catalog_hash` will touch many existing JSON expectations, decide in the plan whether the default test hash is a fixed process value or whether individual tests call a helper to set one. The field should be present in shaped responses even when a test did not construct an `mcp.Server`; otherwise direct tool tests will not exercise the same response shape as runtime MCP calls.

## Minor status issue

`STATUS.md` still says `Current Step: Step 1: Catalog hash` even though Step 1 is complete and Step 2 is being planned. Update that header before starting implementation.

Once the plan includes these details, the Step 2 implementation should be straightforward and low-risk.
