# R007 plan review — Step 3: Registry filtering composition

Verdict: **changes requested before implementation**

I reviewed `PROMPT.md`, the current `STATUS.md`, R006's Step 2 notes, and the existing `internal/mcp`/`internal/tools`/`internal/safety` plumbing. The Step 3 checklist captures the right outcome, but the plan is not yet specific enough to implement safely. The critical missing pieces are where the resolved toolset is applied, how skip-count semantics compose with delete-mode, and which tests will prove that hidden tools are absent from `tools/list` rather than registered with runtime errors.

## What looks good

- Step 3 is correctly focused on registration-time filtering, not response `_meta` surfacing or the advanced-capabilities tool.
- The existing `safeRegistrar` is the right chokepoint to extend: it already validates tool definitions, applies the delete-mode registration gate, and is what feeds the SDK's `tools/list` surface.
- Step 2 already provided the required metadata (`Tool.Toolset` plus `EffectiveToolset`) and catalog membership table, so Step 3 should not need a production name-to-tier map.

## Required plan adjustments

1. **Pin the active-toolset propagation path.**
   - The plan must say how the resolved `Config.Toolset`/`ServerInfo.Toolset` reaches `internal/mcp.safeRegistrar` without re-reading the environment and without adding any model-controlled override.
   - A good shape is either a dedicated `Toolset safety.Toolset` on `mcp.Options` or a deliberate use of `opts.Config.Toolset`; in either case `app.defaultStartServer` must pass the already-resolved value down to MCP.
   - Default/empty active toolset must resolve to `core`, matching Step 1 and the task acceptance criteria.

2. **Keep filtering in the registration chokepoint and validate before skipping.**
   - Extend `safeRegistrar` (or a small helper it calls) so a tool registers only when both gates allow it: active toolset allows `tool.EffectiveToolset()` and capability allows the tool's write/delete requirement.
   - Validation must still run before applying `EffectiveToolset`. This preserves the R006 requirement that unknown non-empty in-code toolsets fail closed with an error instead of being normalized to `full` and skipped/registered silently.
   - Do not implement a production name-to-tier map in MCP or safety; the self-declared `Tool` metadata remains the source of truth.

3. **Define tier semantics explicitly.**
   - `core` should register only tools declared `core`.
   - `full` should register both `core` and `full` tools (subject to delete-mode), because it represents the full surface, not only the full-only subset.
   - Empty tool declarations still behave as `full` after validation, so unmarked future tools do not expand the default core catalog.

4. **Specify skip-count semantics and log keys.**
   - Replace the single `skipped_count` plan with clear per-gate counters, e.g. `registered_count`, `skipped_toolset_count`, and `skipped_capability_count` (optionally `evaluated_count`).
   - State whether a tool disallowed by both gates increments both skip counters or only the first gate checked. I recommend independent gate evaluation so the log reports how many tools each gate would suppress; if you choose ordered counts, document the gate order because the counts will otherwise be ambiguous.
   - Keep the startup log count-only: no tool names and no descriptions.

5. **Add a composition test matrix before coding.**
   - Use a synthetic registry with at least: core read, core write, full read, full write/delete. Then table-test active toolset (`core`/`full`) crossed with delete mode (`none`/`safe`/`full`).
   - Assert the exact names returned by `tools/list` for each combination, proving full-only tools are absent in `core` and delete/write tools are absent when capability disallows them.
   - Include the important cases: `core + full delete mode` still hides full-only tools; `full + safe` still hides delete tools; `full` includes core tools.

6. **Cover protocol absence and logging.**
   - Add/extend a protocol-level test that a full-only tool hidden by `core` is not in `ListTools`; ideally also assert that calling it produces the SDK's unknown-tool protocol error rather than a registered tool returning a custom error.
   - Update the registration log test to assert the new per-gate count fields and continue asserting that tool names are not leaked.

7. **Account for existing tests whose fixtures are unmarked.**
   - Once default active toolset is `core`, current test-only tools with empty `Toolset` will be treated as `full` and skipped unless the test sets the active toolset to `full` or marks the fixture tool as `core`.
   - The plan should call out updates to `testEchoRegistry`, `capabilityRegistry`, and related protocol tests so Step 3 does not accidentally weaken the default-core behavior just to preserve old tests.

Once `STATUS.md` records these implementation choices and the test matrix, Step 3 should be ready to implement. Keep Step 4 (`icuvisor_list_advanced_capabilities`) and Step 5 (`_meta.toolset`, docs, changelog) out of this step unless the task is intentionally re-scoped.
