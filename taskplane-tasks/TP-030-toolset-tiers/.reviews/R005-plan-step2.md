# R005 plan review — Step 2: Per-tool tier membership

Verdict: **APPROVE**

I reviewed `PROMPT.md`, the revised `STATUS.md`, R004, and the current registry/safety/MCP plumbing. The Step 2 plan now addresses the blocking gaps from R004 and is scoped appropriately for a membership-only step.

## Why this is ready

- The metadata API is explicit: add `tools.Tool.Toolset safety.Toolset` and an effective-tier helper that treats an empty field as `full`. This avoids the Step 1 `Toolset("").String() == "core"` trap.
- Unknown non-empty in-code tier values are planned to fail validation, which is the right behavior for a developer/catalog bug.
- Tier membership remains self-declared by each tool constructor. The plan explicitly avoids a production name-to-tier map.
- The exact current catalog split is recorded in `STATUS.md`: 16 existing `core` tools, 21 existing `full`-only tools, plus the planned Step 4 `icuvisor_list_advanced_capabilities` core tool for a total target of 17 core tools.
- The core/full choices are consistent with PRD §7.2.E: daily activity/fitness/wellness/event reads and non-destructive event/wellness/message writes are core; heavy raw streams, specialist training-plan/workout/custom-item/sport-settings surfaces, and destructive deletes are full.
- The planned drift test has the right property: every current registered tool must be accounted for, and unexpected newly registered tools should fail the test until consciously classified.
- Step boundaries are preserved: no `tools/list` filtering, no skip-count logging, and no implementation of `icuvisor_list_advanced_capabilities` in Step 2.

## Implementation notes to keep during coding

- Add tests for both helper/default behavior and invalid non-empty tier validation. In particular, assert that `tools.Tool{}` (or an otherwise unmarked tool) has effective tier `safety.ToolsetFull`, and that a tool with `Toolset: safety.Toolset("surprise")` is rejected by the same validation path used for MCP registration.
- Prefer extending the existing adversarial catalog matrix or using the same registration fixture style, so the tier table and existing requirement table do not drift independently. A test-only expected map/table is fine; production code should stay self-declared.
- When comparing/validating a tool's tier, do not use `safety.ParseToolset` or `Toolset.String()` for unknown values, because those deliberately normalize unknowns to `core` for env/config safety. Validation should compare the raw field against `""`, `ToolsetCore`, and `ToolsetFull`.
- Do not rely on safe-mode registration when building the membership matrix; it must observe delete/write tools too. Use a collecting registrar or full capability so all current tools are checked.

With those notes, the plan is sufficiently specific to implement Step 2.
