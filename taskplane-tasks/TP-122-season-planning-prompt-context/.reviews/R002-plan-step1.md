# Plan Review — Step 1

Verdict: **Approved.**

The revised Step 1 discovery entries now record a concrete approach: enhance the existing prompt surface, primarily `weekly_planning`, with supporting context in `weekly_review` and `race_week_taper`, instead of adding a new `season_planning` prompt. This fits the PRD’s current six-prompt catalog and avoids unnecessary registry/doc churn.

The plan also correctly limits the workflow to deterministic reads/analyzers (`get_events`, `get_training_plan`, `get_fitness`, `get_training_summary`, `compute_compliance_rate`, `icuvisor_list_advanced_capabilities`) and explicitly treats `get_training_plan`/`compute_compliance_rate` as possibly unavailable, with fallback guidance. The non-goals are clear: no ATP writer, no automatic calendar filling/ATP-note creation, and no write/delete calls before user approval of exact changes.

Implementation cautions for Step 2:

- Keep prompt text terse enough for `TestPromptResourceCitationsStayTerse` (currently guards against verbose/schema-like prompts and >25 lines).
- Make the strengthened `weekly_planning` guidance explicitly gather race date/priority, active plan, planned events, current load, recent completion/compliance, and available write capabilities before proposing edits.
- If Step 2 ends up adding a new prompt despite this plan, revisit PRD/catalog/golden-test updates as user-visible scope.

Verification run during review: `go test ./internal/prompts` passes.
