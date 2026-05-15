# Review R002 — Plan review for Step 1: Catalog hash

Verdict: **Approved for implementation**, with a few implementation guardrails to preserve determinism.

The updated Step 1 plan in `STATUS.md` addresses the blocking concerns from R001:

- It explicitly targets the **exposed registered catalog after toolset/capability filtering**, rather than the pre-filter `catalogCollectingRegistrar` list.
- It includes the LLM-facing contract fields that should affect the hash: tool name, tool-level description, input schema, and advertised output schema.
- It plans to store the computed value on `internal/mcp.Server` and expose `Server.CatalogHash() string`.
- The listed tests now cover registration order, nested map order, capability/toolset filtering, add/remove/rename, description edits, tool-description edits, and output-schema edits.

## Guardrails for implementation

These do not require another plan review, but should be followed while coding:

1. **Use an explicitly framed canonical payload.**
   `"canonical sorted records"` should become a concrete format in code, e.g. a JSON array of records sorted by `name`, or length-prefixed fields. Avoid raw byte concatenation of `name + schema` boundaries. Pin the expected hash for a small fixture so accidental format drift is visible.

2. **Capture only successfully registered tools.**
   Append a tool to the hash input only after `toolsetAllows` and `capabilityAllows` pass and the same `tools.Tool` has been handed to `sdkmcp.Server.AddTool` without panic. `list_advanced_capabilities` should be included if it is actually exposed, but the full pre-filter collector contents should not be hashed directly.

3. **Fail closed on canonicalization errors.**
   If a schema cannot be marshaled/canonicalized, `NewServer` should fail with a wrapped registration/hash error rather than silently omitting that tool or hashing an empty schema.

4. **Keep the hash independent of version/build metadata.**
   The catalog hash should represent the exposed catalog and schema contract only. `server_version` is tracked separately in Step 2 and should not perturb the hash for an otherwise identical catalog.

5. **Normalize absent output schemas consistently.**
   Decide once whether a missing output schema is encoded as `null` or omitted from the record, and cover that in the golden fixture. Either is fine as long as it is deterministic and documented by the test.

6. **Avoid an awkward runtime dependency on CI-only helpers.**
   Reusing `internal/toolchecks.CanonicalJSON` is acceptable only if it does not make `internal/mcp` depend on test/CI-oriented behavior in a confusing way. Otherwise, add a small local canonicalization helper in `internal/mcp` or move a shared helper to a neutral internal package.
