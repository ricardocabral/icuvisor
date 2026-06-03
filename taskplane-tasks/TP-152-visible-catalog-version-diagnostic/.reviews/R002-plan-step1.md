# Plan Review: Step 1 — Design diagnostic contract

Verdict: **APPROVED**

The revised contract addresses the R001 blockers: it defines a visible catalog fingerprint in the tool description, avoids claiming the server can observe stale client state, and adds same-version catalog/schema drift coverage. The privacy boundary remains appropriately narrow.

## Implementation notes to carry into Step 2

- Ensure the top-level `catalog_hash` returned by the tool comes from the runtime catalog metadata after `NewServer` computes the exposed catalog hash, not from a registration-time placeholder/default.
- Keep the fingerprint algorithm in a package/layer that avoids an `internal/tools` ↔ `internal/mcp` import cycle. If the diagnostic description is generated in `internal/tools`, the comparable fingerprint helper likely cannot live only in `internal/mcp` unless it is injected or moved.
- The description fingerprint should be computed over the same effective catalog semantics the client sees as closely as possible: delete-mode/toolset filtering, plus any registration-time schema/description mutations such as coach-mode athlete routing if applicable. If coach-mode dynamic per-athlete visibility is intentionally excluded, document that limitation in tests or comments.
- Test that visible description fields (`description_server_version`, `description_catalog_fingerprint`, `description_toolset`, `description_delete_mode`) are mirrored or comparable to response fields with unambiguous names.

Proceed to Step 2.
