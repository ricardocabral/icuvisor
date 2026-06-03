# Plan Review: Step 2 — Implement `get_planning_context`

**Verdict: Approved.**

The revised Step 2 plan addresses the previous integration gaps and is specific enough to implement safely.

## What is now clear

- The tool remains **read-only** and composes only `GetAthleteProfile`, `ListEvents`, `GetTrainingPlan`, and `ListAthleteSummary`.
- Tier placement is **full** because it exposes training-plan-derived context.
- Catalog placement is recorded as `workout-library`, alongside `get_training_plan` / `apply_training_plan`.
- Registration drift is accounted for: `registryBaseTools`, `toolCatalogGroup`, `internal/toolcatalog` constants, `athleteScopedToolNames`, and `catalog_tiers_test.go` are all called out.
- Deterministic clock injection is planned via a `newGetPlanningContextToolWithClock`-style constructor, which should make week anchoring, as-of metadata, current fitness window, and race scan tests stable.
- The Step 1 contract is detailed enough for implementation: week window, current fitness window, race scan limit/truncation caveats, event classification, caveat codes, `include_full` scope, and `_meta.source_tools` are all defined.

## Implementation reminders

- Keep `get_planning_context` as `fullTool(...)`; do not promote it to core without a separate toolset policy decision.
- Treat `len(events) >= 500` as `*_may_be_truncated` consistently for both week and race scans, since upstream applies the limit.
- Ensure the output schema explicitly says `planning_scope: context_only`, `read_only: true`, and no ATP/calendar writes.
- Step 3 should still add handler tests for future `week_start` with current fitness window, truncation caveats, empty-data caveats, and `include_full` raw payload gating.

No further plan changes are required before coding Step 2.
