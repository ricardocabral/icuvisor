# Plan Review: Step 3 — Test and document

**Verdict: Needs revision before implementation.**

The Step 3 plan lists the right broad categories, but it is too generic for the detailed contract already established in Step 1/R005. Before writing tests/docs, expand `STATUS.md` so the test matrix locks down the edge cases that are most likely to regress.

## Required fixes

1. **Make the handler test matrix explicit.** Include table-driven cases for:
   - terse default shape: no `full` payloads, exact `_meta.source_tools` (`get_athlete_profile`, `get_events`, `get_training_plan`, `get_fitness`), `read_only=true`, `writes_performed=false`, `planning_scope=context_only`, section counts, and window metadata;
   - `include_full:true` only widening event `full`, fitness row `full`, and raw training-plan assignment/nested payloads;
   - athlete-local clock behavior: default upcoming Monday, supplied mid-week `week_start` normalized back to Monday, and invalid `week_start` returning a short user error;
   - future `week_start` while the fitness call still uses the current as-of 7-day window (`as_of_date-6..as_of_date`), independent of the planning week;
   - event calls using `Limit: 500` for both week events and race scan, with the expected date ranges;
   - classification/filtering: `WORKOUT`, `RACE`/`RACE_*`, `NOTE`, and unknown/other categories, plus `upcoming_races` filtered from the race scan;
   - caveat codes for empty data and partial data, including `read_only_no_atp`, `no_week_events`, `no_week_workouts`, `no_active_training_plan`, `partial_training_plan_summary`, `no_fitness_rows`, `no_upcoming_races`, and both truncation caveats when returned rows reach the 500 limit.

2. **Treat catalog/registration tests as required, not “if needed.”** The plan should name the assertions to keep: registered as `safety.ToolsetFull`, grouped under `workout-library`, known in `internal/toolcatalog`, and athlete-scoped for ACL validation.

3. **Clarify documentation generation.** `CHANGELOG.md` is mandatory. For `README.md`, either update it or record why it has no tool list to change. Also account for generated catalog docs (`make docs-tools` / `web/data/tools.json`) if this repository expects the public tool catalog data to be regenerated for new tools.

Once those details are added to the Step 3 plan, the testing/documentation plan should be ready.
