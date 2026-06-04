# Review R001 — Plan Step 1

Verdict: approved with implementation caveats.

The chosen policy is sound: one canonical full-mode, full-toolset, coach-enabled snapshot set is the right maintenance tradeoff for guarding every public MCP input schema, including delete/write and coach-only tools. Having no intentional exclusions for TP-153 also matches the mission and avoids future silent gaps.

Required caveats for Step 2:

- The current generator path in `internal/toolchecks/schema_stability.go` registers via `tools.NewRegistryWithOptions` and a custom `schemaRegistrar`; it does **not** pass through `internal/mcp`'s `safeRegistrar.prepareTool`, where coach-mode `athlete_id` is injected. To satisfy the stated coach policy, Step 2 must either generate from the actual MCP registration/list-tools path or move/reuse the athlete-id schema injection helper in a shared location. Do not duplicate a second hard-coded `athlete_id` schema that can drift from runtime MCP behavior.
- Make the broad mode explicit in generation/tests: full delete capability, full toolset, coach mode enabled with a permissive test roster/ACL. Relying on defaults would under-document the policy and may accidentally snapshot safe/core semantics.
- Add a test that the generated snapshot name set equals the live full coach-enabled public registry, with an explicit exclusion map that is currently empty. This is what prevents returning to a curated whitelist.

With those constraints, the Step 1 policy is acceptable to implement.
