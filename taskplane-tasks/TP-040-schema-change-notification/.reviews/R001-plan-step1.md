# Review R001 — Plan review for Step 1: Catalog hash

Verdict: **Request changes before implementation.**

The step goal is right, but the current plan in `STATUS.md` is too thin for the tricky parts of this repository. Step 1 should be tightened before coding so the hash matches the actual MCP catalog seen by the client and stays deterministic.

## Required plan fixes

1. **Hash the exposed registered catalog, not the potential tool list.**
   - `internal/tools/defaultRegistry.Register` currently wraps the downstream registrar with `catalogCollectingRegistrar`, but that collector records every attempted tool before `internal/mcp.safeRegistrar` applies `Toolset` and delete/write capability filtering.
   - The catalog hash must cover only tools actually registered with the SDK and visible to the MCP client. For safe/core mode this excludes skipped full/delete/write tools. Computing from `collector.tools`, schema snapshots, or a second synthetic registry pass would produce the wrong runtime hash.
   - Plan should state that hash input is captured after `safeRegistrar.toolsetAllows` and `safeRegistrar.capabilityAllows` pass, using the same `tools.Tool` values sent to `sdkmcp.Server.AddTool`.

2. **Define exactly what fields are hashed.**
   - Step text says `(tool_name, marshalled JSON schema)` but also requires that a description-only change changes the hash. Clarify whether this means:
     - input-schema property descriptions only, or
     - tool-level `Description` as well.
   - Since the mission says “tool catalog or argument schemas differ” and MCP clients cache the catalog, the safer plan is to hash at least `name`, tool-level `description`, and `input_schema`. If `output_schema` is advertised in the catalog by the SDK, either include it too or explicitly document why Step 1 excludes it.

3. **Use unambiguous canonical framing, not raw concatenation.**
   - The prompt says SHA-256 of sorted concatenation, but a robust plan should specify length-prefixes or a JSON array of per-tool records before hashing. Plain concatenation of `name` + schema bytes can theoretically collide across boundaries.
   - Suggested format: build `[]catalogHashTool{{Name, Description, InputSchema, OutputSchema?}}`, sort by `Name`, canonical JSON marshal it, then SHA-256 and lowercase hex encode. Alternatively, length-prefix every field.

4. **Re-use or mirror canonical JSON carefully.**
   - `internal/toolchecks.CanonicalJSON` already produces deterministic JSON for schema snapshots. Step 1 should either reuse a shared canonicalization helper or implement equivalent deterministic JSON locally in `internal/mcp` without creating an awkward runtime dependency on the CI-oriented `toolchecks` package.
   - Tests should fail on non-marshalable schema values rather than silently skipping them.

5. **Plan the `Server.CatalogHash()` storage path now.**
   - The task explicitly expects `internal/mcp/catalog_hash.go` and `Server.CatalogHash() string` computed once at server start. The Step 1 plan should include adding a `catalogHash string` field to `mcp.Server` and exposing the method, even if `_meta` wiring is Step 2.

## Test plan additions for Step 1

Add tests that cover these repository-specific cases:

- Registration order does not change the hash.
- Map key order in nested schema maps does not change the hash.
- Active capability/toolset filtering changes the hash when the exposed client catalog changes, and skipped tools are not included.
- Adding/removing a tool changes the hash.
- Renaming an argument changes the hash.
- Editing an argument description changes the hash.
- If included by design: editing tool-level description and/or output schema changes the hash.
- A golden fixture pins the exact hash for a small fake catalog, using deterministic fake tools instead of the full production registry.

## Notes

- Avoid computing the hash from `toolchecks.GenerateSchemaSnapshots()` as-is: it uses a synthetic all-capabilities client and would not necessarily match the runtime catalog for the current `Toolset`/capability mode.
- Logging the hash once at startup is acceptable later, but Step 1 should not add per-response logging.
