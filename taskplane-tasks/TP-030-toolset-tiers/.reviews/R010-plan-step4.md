# R010 plan review — Step 4: `icuvisor_list_advanced_capabilities`

Verdict: **REVISE**

I reviewed `PROMPT.md`, `STATUS.md`, the Step 1-3 implementation state, and the existing `internal/tools`/`internal/mcp` registration plumbing. The Step 4 checklist states the right outcome, but it is not yet an implementation plan. Before coding, `STATUS.md` should pin how the tool will derive its catalog, how it will know the active toolset, and which tests will prevent drift.

## What looks good

- The step is correctly scoped to the discoverability tool, not docs or global `_meta.toolset` surfacing.
- The three stated requirements match the task prompt: the tool must be core, static/no-upstream, and still useful when `ICUVISOR_TOOLSET=full` is active.
- Step 2 already provides the source metadata needed for this tool (`Tool.Toolset`, `EffectiveToolset`, `Requirement`, and each tool's description), so this should not require a separate production name-to-tier map.

## Required plan adjustments

1. **Define the catalog-derivation mechanism.**
   - The plan must say how `icuvisor_list_advanced_capabilities` gets the full-only tools from the same registered catalog metadata instead of a second hand-maintained list.
   - A good shape is a small internal catalog collector/wrapping registrar inside `internal/tools` that records every `Tool` passed through `defaultRegistry.Register`, derives rows where `EffectiveToolset() == safety.ToolsetFull`, and forwards the tool to the real registrar so Step 3 filtering remains unchanged.
   - Derive the one-line summary from the tool's existing first description sentence. Do not introduce a separate summary table that can drift from tool descriptions.

2. **Pin where the tool is registered and how ordering works.**
   - The plan should say that `newListAdvancedCapabilitiesTool(...)` is added by the default registry after the existing catalog has been collected, so its output can include all full-only tools and exclude itself/core tools.
   - Mark it with `coreTool(...)`, `RequirementRead`, an empty/no-argument schema with `additionalProperties:false`, and a generic structured output schema.
   - Keep it available in both active tiers: core registers it as a core tool, and full includes core tools by Step 3 semantics.

3. **Propagate the active toolset to the handler without re-reading env.**
   - The handler has to say different things in `core` vs `full`, so the plan needs an explicit propagation path.
   - Prefer adding `Toolset safety.Toolset` to `tools.RegistryOptions`, defaulting empty/invalid to `core`, and pass the already-resolved `toolset` from `app.defaultStartServer` into `tools.NewRegistryWithOptions`.
   - Do not read `ICUVISOR_TOOLSET` from the tool and do not add any request argument that changes the catalog.

4. **Specify the response shape and exact enablement wording.**
   - Record a concrete terse shape in the plan, for example: `advanced_capabilities: [{name, summary, requirement}]`, `enable_instruction`, `current_toolset`, and `_meta.count`/`_meta.source`.
   - The response must include the exact string `ICUVISOR_TOOLSET=full` in both text and structured content, with wording that tells the user to set it in the MCP client/server environment and restart icuvisor.
   - When the active tier is already `full`, include an explicit status such as “full toolset is already enabled” while still returning the same full-only catalog.
   - Because delete-mode filtering is orthogonal, include either each tool's `requirement` or a short note that destructive tools may also require the existing delete-mode gate. Otherwise the tool can misleadingly imply `ICUVISOR_TOOLSET=full` alone exposes delete tools.

5. **Add the drift and behavior tests to the plan.**
   - Update the tier membership matrix to include `icuvisor_list_advanced_capabilities` as core and adjust the expected current core count/list.
   - Add handler tests proving the output includes known full-only tools such as `get_power_curves`, excludes core tools such as `get_athlete_profile`, excludes itself, has one-line summaries, includes `ICUVISOR_TOOLSET=full`, and works for both active `core` and `full`.
   - Add a no-upstream-call test using a fake/panic client or call-counting client so this remains a static catalog tool.
   - Add protocol/registration coverage that `tools/list` in default core includes `icuvisor_list_advanced_capabilities` while still hiding a known full-only tool, and that full mode also includes the discoverability tool.
   - Update existing default-registry tests that currently expect only `get_athlete_profile` for a profile-only client; after Step 4 they should expect the profile tool plus the discoverability tool.

6. **Mention schema/catalog checks affected by adding a new tool.**
   - If the toolcheck schema snapshot flow is expected in CI or `make lint`, add/update the `icuvisor_list_advanced_capabilities` input schema snapshot.
   - Ensure the confusable-description check still passes; its first sentence should be distinct and focused on discovering hidden full-toolset capabilities.

Once these choices and tests are recorded in `STATUS.md`, the Step 4 implementation should be straightforward and should compose cleanly with the Step 3 registration filter. As written, the current three checkboxes are acceptance criteria, not a sufficient implementation plan.
