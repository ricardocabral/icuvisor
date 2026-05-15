# Review R007 — Plan review for Step 2: `_meta` injector

Verdict: **APPROVE**

The current Step 2 plan in `STATUS.md` addresses the remaining gaps from R005/R006 and is specific enough to implement.

## What looks good

- The runtime metadata path is now explicit: `internal/mcp.NewServer` will publish the computed `Server.CatalogHash()` into `internal/response` without adding the hash to tool descriptions or schemas.
- The plan calls for concurrency-safe current/first-seen snapshot handling and a documented per-process fallback caveat, which matches the SDK/session-handle constraints described in the prompt.
- `_meta.catalog_hash` is planned as response-owned metadata on every shaped response, with caller-provided schema-change keys overwritten to prevent accidental/spoofed metadata.
- The direct-response audit is broad enough now: it covers `StructuredContent: payload`, `StructuredContent: response`, and direct `json.Marshal(...)` paths, and explicitly includes the known bypasses `list_advanced_capabilities`, `update_sport_settings`, and `update_wellness`.
- The no-server/default catalog-hash behavior for direct tool tests is now called out, which should keep `_meta.catalog_hash` deterministic and present even when tests invoke tool handlers without constructing an MCP server.
- The `schemaChangeMessage(previousVersion, currentVersion)` seam is explicit and testable.

## Implementation notes

- Keep the runtime metadata setter/reset helpers small and package-scoped/test-oriented where possible so the process-global state does not become a broad API surface.
- When converting bypassing tools, ensure the JSON text content and `StructuredContent` are produced from the same shaped value so they cannot diverge.
- Use a single locked/atomic snapshot read for version/hash pairs; avoid independently loading version and hash if that could produce mixed pairs during tests that simulate changes.

Proceed with Step 2 implementation.
