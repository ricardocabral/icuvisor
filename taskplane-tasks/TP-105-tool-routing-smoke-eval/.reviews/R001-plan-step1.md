# Plan Review: Step 1 — Design eval fixture and expected-result format

**Verdict: Needs revision**

I could not find a concrete Step 1 plan artifact beyond `STATUS.md` marking the step in progress. As-is, the plan is not specific enough to approve because the key design choices for fixture location, schema, mode-specific expectations, and validation source are unspecified.

## Required clarifications before implementation

1. **Reuse/fit with existing eval harness**
   - The repo already has `scripts/eval/` with scenario validation and `scripts/benchmark/` patterns. The plan should explicitly state whether this new smoke eval extends `scripts/eval/` or creates a separate `scripts/eval/tool_routing` area, and why.

2. **Catalog validation source**
   - The mission requires loading icuvisor's registered tool definitions. Do not validate only against generated `web/data/tools.json`; that can be stale. The plan should define how Step 1 tests will validate expected tool names against registered core/full + safe/full catalogs without executing handlers.

3. **Expected-result schema**
   - This eval is about the **first selected tool**, not a list of expected tools. The fixture format should include something like: `id`, `prompt`, `catalog_mode`/`toolset`, `expected_first_tool` or explicit `expected_no_tool`, optional `allowed_first_tools`, and `notes`.
   - Destructive-tool cases need mode-specific expectations, because safe catalogs may not expose delete tools while full catalogs do.

4. **Comparison semantics**
   - Define how result comparison handles exact match, allowed alternatives, no-tool, unavailable/skipped cases, and unknown tools. Add table-driven unit tests for these paths.

5. **Initial coverage**
   - The Step 1 plan should name representative cases for: confusable activity reads, event write/delete routing under safe/full modes, workout-library operations, and analyzer/helper routing.
   - Fixtures must avoid real athlete IDs, API keys, exact private dates, and any dependency on network/provider access.

Once those details are in the plan, the step should be straightforward and low-risk.
