# R004 plan review — Step 2: Per-tool tier membership

Verdict: **changes requested before implementation**

I reviewed `PROMPT.md`, the current `STATUS.md`, and the existing registry/safety plumbing. The Step 2 checklist has the right high-level goals, but it is not yet specific enough to implement safely. In particular, the plan does not pin the exact `core` list, does not say how the default-`full` behavior will be represented without conflicting with `safety.Toolset.String()`, and does not identify the drift-catching test shape.

## What looks good

- Keeping toolset policy in `internal/safety` remains the right boundary from Step 1.
- Step 2 is correctly scoped to membership metadata and catalog tests, leaving actual registration filtering/count logging for Step 3.
- The plan recognizes the important product rule: tools opt in to `core`; unmarked/new tools should be `full` by default.

## Required plan adjustments

1. **Specify the concrete metadata API before editing every tool.**
   - Add a membership field to `tools.Tool`, e.g. `Toolset safety.Toolset` or `Tier safety.Toolset`, and set it in tool constructors.
   - Add a helper such as `func (t Tool) EffectiveToolset() safety.Toolset` / `Tier()` that returns `safety.ToolsetFull` when the field is empty.
   - This helper is important because Step 1 intentionally made `safety.Toolset("").String()` render as `core`; using `tool.Toolset.String()` directly would accidentally make unmarked tools core, violating the Step 2 requirement.
   - Decide whether invalid in-code tier values fail validation or normalize to `full`, and include a test for that behavior. I recommend validation failure for unknown non-empty values, while empty remains the documented `full` default.

2. **Keep membership self-declared, not a parallel production name map.**
   - The source of truth should be the `Tool` returned by each `new*Tool` constructor.
   - Avoid putting a production `map[string]Toolset` in `internal/safety` or the registrar; that would fork the catalog and undermine the “each tool self-declares” requirement.
   - A test-only expected table is fine and should be used to catch drift.

3. **Record the exact `core` set in `STATUS.md` before implementation.**
   - The current plan says “target ~17 tools” but does not say which tools. Step 2 should not proceed until the list is pinned.
   - The list should explicitly classify all current tools as `core` or `full`, and note that `icuvisor_list_advanced_capabilities` will be `core` when implemented in Step 4.
   - Include a short rationale for borderline tools. For example, raw/heavy or specialist surfaces (`get_activity_streams`, workout-library CRUD, custom-item CRUD, sport-settings updates, training-plan application, deletes) likely belong in `full`; daily-use activity/fitness/wellness/event reads and event/wellness/message writes belong in `core`.

4. **Define the catalog drift test explicitly.**
   - Add a table-driven test that registers the full current catalog with a real/stub `intervals.Client` (similar to the existing adversarial catalog test) and asserts every registered tool has the expected effective tier.
   - The test must fail on both missing expected tools and unexpected newly registered tools, so new tools do not silently inherit `full` without a conscious test/table update.
   - Also add a small unit test for the empty-tier default: `Tool{}` (or a tool without the field set) has effective tier `full`.

5. **Preserve step boundaries.**
   - Do not filter `tools/list` in Step 2; that belongs to Step 3.
   - Do not implement the `icuvisor_list_advanced_capabilities` tool in Step 2 unless you intentionally merge Step 4 scope. If it remains Step 4, record its planned `core` membership in `STATUS.md` and extend the matrix when the tool is added.
   - Do not add startup registered/skipped counts per gate yet; those are Step 3.

Once `STATUS.md` includes the exact tier table and the plan names the `Tool` metadata/helper plus the drift tests above, Step 2 should be ready to implement.
