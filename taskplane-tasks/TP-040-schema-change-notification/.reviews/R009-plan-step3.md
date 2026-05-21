# Review R009 — Plan review for Step 3: Tests

Verdict: **REVISE**

The Step 3 checklist in `STATUS.md` is directionally correct, but it is too high-level to approve as a test plan for this task. It misses one explicit acceptance requirement and should fold in the unresolved regression from the previous code review.

## Required plan changes

1. **Add an actual tool-result JSON assertion, not only `response.Shape` struct/map assertions.**
   The prompt explicitly requires confirming `_meta.schema_changed` appears in the actual response JSON. The plan should name a handler/encode-path test that parses the JSON text returned in `tools.Result.Content` and asserts the same metadata is present in `StructuredContent`. This also verifies the common `encodeShaped` boundary rather than only the internal Go map.

2. **Add a regression test for `list_advanced_capabilities` server version metadata.**
   R008 found that `list_advanced_capabilities` still calls `encodeShaped(..., "", ...)`, which makes its response JSON emit `_meta.server_version: "dev"` even for a released server version. Step 3 should explicitly add a test that exercises this tool through the real registry/server-version path and asserts the response JSON has the configured version plus the runtime `catalog_hash`. This test should fail until the Step 2 bug is fixed.

3. **Make global runtime metadata isolation part of the plan.**
   The catalog first-seen/current metadata is process-global. Any tests that call `setRuntimeCatalogMetadataForTest` must reset it with `t.Cleanup`, and schema-change simulations should not run in parallel with other response-shaper tests. Otherwise tests may become order-dependent or spuriously emit `schema_changed`.

4. **Spell out the golden-file stabilization mechanism.**
   “Tool golden files unaffected” should state the concrete approach: either keep the deterministic default catalog hash for direct tool tests or inject a fixed test catalog hash before golden assertions, then verify `go test ./...` does not require build-specific golden changes. This is part of the acceptance criteria, not just a nice-to-have.

## What can stay

- The existing catalog-hash determinism/sensitivity coverage listed from Step 1 is appropriate for Step 3 verification; keep it as a regression suite.
- The injector tests should cover session-start, steady-state, simulated hash change, spoofed caller `_meta` keys being overwritten, and the exact `schemaChangeMessage` template.

Please update the Step 3 plan in `STATUS.md` with these concrete test targets before implementation.
