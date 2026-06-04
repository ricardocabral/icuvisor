# Plan Review: Step 2 — Implement tool and tests

Verdict: **APPROVE**

The revised Step 2 plan addresses the R004 blockers. It now specifies a live runtime metadata getter for `catalog_hash`, avoids an `internal/tools` ↔ `internal/mcp` import cycle by keeping the visible fingerprint helper in `internal/tools`, defines self-reference normalization, clarifies registration order, and calls out shared catalog updates plus targeted test coverage.

## Notes to carry into implementation

- When computing `description_catalog_fingerprint`, mirror the actual safe-registration filters as closely as possible: capability/delete-mode, toolset, and the tools added after base registration (`icuvisor_list_advanced_capabilities` and the diagnostic placeholder). If coach-mode `athlete_id` schema injection is not represented in this fingerprint, make that limitation explicit in a test/comment alongside the already-planned dynamic ACL limitation.
- Ensure the diagnostic response compares like with like: visible `description_*` fields in the tool description should match top-level response fields with the same names/semantics. Keep `catalog_hash` as the live MCP catalog hash from response runtime metadata, not as a substitute for the description fingerprint.
- The no-leak test should assert absence of configured athlete IDs, API-key-like values, local paths/usernames, and raw env/config values; returning normalized `toolset` and `delete_mode` is acceptable.

Proceed to implementation.
