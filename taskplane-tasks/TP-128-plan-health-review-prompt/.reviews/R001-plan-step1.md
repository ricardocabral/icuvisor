# Review R001 — Plan review for Step 1

**Verdict:** Changes requested

The Step 1 checklist points at the right files and tools, but it is not yet a complete "prompt contract" plan. Before implementing, tighten the design deliverable so Step 2 cannot drift into an ad hoc weekly-review variant or a black-box plan-health score.

## Required changes

1. **Define the prompt contract explicitly.** Step 1 should record the chosen name/approach (`plan_health_review` vs extending `weekly_review`), arguments, default date windows, and output sections. The contract should cover planned window, completed lookback, optional race date/horizon, and how race risk is reported when no race event is found.

2. **Include formula-transparency resources and fallbacks.** The tool sequence should name `get_fitness_projection`, `compute_compliance_rate`, `compute_load_balance`, wellness reads, events/training plan, and when to call `icuvisor_list_advanced_capabilities` if full-tool analyzers are unavailable. Because the mission is formula-transparent, decide whether the prompt cites `icuvisor://analysis-formulas` and requires answers to quote analyzer `_meta.method` / assumptions instead of inventing a score.

3. **Preserve product-scope boundaries.** If a new prompt is added, Step 1 must note that PRD §7.2.G currently lists six prompts and that docs/tests will need to be updated or the PRD impact explicitly deferred. The design should also state that this is a review workflow only: no plan-filler behavior, no autonomous coaching/physiology model, and no calendar writes without reviewing the exact proposal first.

4. **Make the Step 1 artifact concrete.** Add an acceptance criterion that `STATUS.md` Discoveries contains a short design summary: decision, arguments, tool order, output shape, caveats for deload/recovery weeks and missing wellness/readiness data, and test implications (`TestNewRegistryRegistersSixPrompts` count/golden fixtures if adding a prompt). Running `go test ./internal/prompts` is useful, but the design record is the main Step 1 deliverable.
