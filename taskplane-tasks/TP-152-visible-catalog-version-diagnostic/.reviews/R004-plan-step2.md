# Plan Review: Step 2 — Implement tool and tests

Verdict: **REVISE**

The Step 1 contract is sound, but STATUS does not yet contain a concrete Step 2 implementation plan. The current Step 2 section is only the original checklist, which leaves a few implementation-critical seams unresolved.

## Blocking issues

1. **Live `catalog_hash` source is not planned.**  
   The handler will be built before `mcp.NewServer` computes the exposed catalog hash. The Step 2 plan must say how `icuvisor_check_server_version` will read the live hash at call time, e.g. by adding a small exported, locked response metadata getter or equivalent injection. Do not compute a registration-time hash placeholder for the top-level `catalog_hash`.

2. **Fingerprint helper/package boundary is unspecified.**  
   The plan needs to state where the description fingerprint algorithm will live so `internal/tools` and `internal/mcp` do not form an import cycle. It should also state how the diagnostic tool normalizes its own injected fingerprint token to a sentinel before hashing.

3. **Effective-catalog semantics are underspecified.**  
   The visible `description_catalog_fingerprint` must be based on the effective catalog the client sees as closely as possible: toolset and delete-mode/capability filtering, plus the diagnostic and advanced-capabilities tools. If coach-mode per-athlete visibility or schema injection is intentionally excluded, Step 2 should explicitly test or document that limitation.

4. **Registration/catalog updates need explicit coverage.**  
   The plan should call out the exact shared catalog updates: add the new constant to `internal/toolcatalog/catalog.go`, include it in `allToolNames`, keep it **out** of `athleteScopedToolNames`, add it to the `meta` group in `internal/tools/catalog.go`, and update tier membership expectations.

## Expected revised Step 2 plan

Please update STATUS with a short implementation plan that covers:

- new tool constructor/handler shape and no-argument validation;
- how the handler obtains current runtime version/hash and returns `description_*`, `toolset`, `delete_mode`, `status`, `action`, and `_meta`;
- where the reusable fingerprint function lives and how it avoids self-reference;
- registration order for base tools, advanced-capabilities, and the diagnostic tool;
- tests for output/no-leak, no network dependency, catalog hash sensitivity, same-version fingerprint drift, known-tool/catalog descriptor membership, and targeted `go test` command.
