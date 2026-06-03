# Review R002 — Plan Step 2

Verdict: approved with required guardrails.

The Step 2 direction is acceptable: replace the curated whitelist with live full-registry coverage, add a missing-coverage guard, and keep generation deterministic/no-network. That matches the Step 1 policy and TP-153 mission.

Required implementation points:

- Generate/compare against the actual exposed MCP catalog, not only `tools.NewRegistryWithOptions` plus a raw registrar. Prefer `mcp.CollectToolCatalog` (or equivalent shared path) so coach-mode `athlete_id` injection and runtime filtering semantics are included without duplicating schema logic.
- Make canonical snapshot options explicit in one helper: delete mode `full`, toolset `full`, coach mode enabled, and a permissive test coach roster/ACL that exposes every public tool. Both the tool registry options and MCP collection config need the same coach settings.
- Remove the static `schemaCatalogToolNames` whitelist as the source of truth. If an exclusion mechanism is introduced, keep it as an explicit `map[name]reason` that is empty for this task and assert unknown/excluded names fail loudly.
- Add a unit test that the generated snapshot name set equals the live full coach-enabled exposed catalog minus explicit exclusions. This is the key anti-regression guard against future tools silently missing snapshots.
- Be careful with Step 2/Step 3 sequencing: do not add a committed-snapshot freshness unit test that must pass before the new JSON files are generated, unless Step 2 also regenerates snapshots. The existing freshness check can fail in CI after Step 2 and be satisfied by Step 3.

Targeted tests are the right validation, provided they include `internal/mcp` because the schema collection path should exercise MCP registration behavior.
