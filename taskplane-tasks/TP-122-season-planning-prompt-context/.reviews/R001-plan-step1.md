# Plan Review — Step 1

Verdict: **Changes requested before marking Step 1 complete.**

The task direction is sound, but `STATUS.md` does not yet record the required design decision or non-goals. Before implementation, add a concrete discovery entry that states whether this will enhance existing prompts or add `season_planning`, and list the exact deterministic tools/fallbacks the prompt guidance will cite.

Required plan refinements:

1. If adding a new prompt, explicitly include registry/golden-test fallout: new name constant, `NewRegistry()` ordering, `TestNewRegistryRegistersSixPrompts` count update/rename, new fixture, and PRD prompt-catalog/doc updates. The current PRD documents six curated prompts, so a seventh prompt is user-visible scope.
2. If enhancing existing prompts, say which ones and why. A safe default seems to strengthen `weekly_planning` for broader planning context and possibly `race_week_taper`/`weekly_review` for race-priority and compliance context, without creating an ATP writer.
3. Treat `get_training_plan` and `compute_compliance_rate` as potentially unavailable full-toolset capabilities. Keep `icuvisor_list_advanced_capabilities` fallback guidance and make clear the workflow can proceed from `get_events`, `get_fitness`, and `get_training_summary` when advanced tools are absent.
4. Make the safety guardrails explicit in the plan: gather race date/priority, active plan, planned events, current load, recent completion/compliance, then return a reviewed proposal and questions; do not automatically fill the calendar, create ATP notes, or call write/delete tools without user approval of exact changes.
5. Keep the prompt terse enough to satisfy `TestPromptResourceCitationsStayTerse`; avoid embedding long workout syntax or physiology rules in prompt text.

Baseline check run during review: `go test ./internal/prompts` passes.
