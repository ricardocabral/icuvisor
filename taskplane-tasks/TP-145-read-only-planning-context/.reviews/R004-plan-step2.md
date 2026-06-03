# Plan Review: Step 2 — Implement `get_planning_context`

**Verdict: Needs revision before implementation.**

The response contract from Step 1 is now strong enough to implement, but the Step 2 implementation plan still misses a few repository-specific integration details that will otherwise cause catalog/test drift.

## Required fixes

1. **Choose and record the catalog group.**
   - The tier is now clear: `fullTool`, because the tool exposes `get_training_plan`-derived data.
   - The group is not specified. `internal/tools/catalog_test.go` only allows existing groups, so decide whether this belongs under `events` or `workout-library` and record that choice before coding. If a new `planning` group is desired, the plan must explicitly include the catalog test/docs updates for that new group.

2. **Expand the implementation/test file scope for tier registration.**
   - Adding a registered tool will fail `internal/tools/catalog_tiers_test.go::TestRegisteredToolTierMembership` unless the expected tier map is updated.
   - Add `internal/tools/catalog_tiers_test.go` to the Step 2/3 touched files and assert `get_planning_context` is `safety.ToolsetFull`.

3. **Plan deterministic clock injection.**
   - The contract depends on athlete-local `as_of`, default Monday week anchoring, and current fitness/race windows.
   - Follow the existing `newGetTodayToolWithClock` / `newGetEventsToolWithClock` pattern and include a `newGetPlanningContextToolWithClock` (or equivalent internal handler clock parameter) so tests do not depend on wall-clock time.

## Implementation notes

- Register the tool in all relevant catalogs: `registryBaseTools`, `toolCatalogGroup`, `internal/toolcatalog` constants and `athleteScopedToolNames`, plus the tier test above.
- Keep the composite client interface read-only and limited to `GetAthleteProfile`, `ListEvents`, `GetTrainingPlan`, and `ListAthleteSummary`; do not route through write tools or add any create/update/delete dependency.
- The input schema should make `week_start` normalization and `include_full` scope explicit, and the output schema/description should state `planning_scope: context_only`, `read_only: true`, and no ATP/calendar writes.

Once those details are captured in `STATUS.md`, the implementation plan is ready.
