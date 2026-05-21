# R019 plan review — Step 5: Catalog-cache caveat + Tests

Verdict: **APPROVE**

The revised Step 5 plan now addresses the prior R017/R018 blockers. It pins `select_athlete.allowed_tools` and `_meta.requires_new_conversation` to an `internal/mcp.safeRegistrar.visibleToolNamesForAthlete` helper over the post-registration/post-gate catalog, adds the hidden-gate metadata regressions that catch delete-mode/toolset leaks, and requires exact structured comparisons across `tools/list`, `select_athlete`, and `icuvisor_list_advanced_capabilities`. The planned fake-client routing checks and two-session isolation coverage are also appropriate for protecting the Step 3/4 coach-mode invariants.

Implementation reminders, not blockers:

- The truth-table tests should keep the two call-time outcomes distinct: tools hidden by delete-mode/toolset should follow the normal unknown-tool path, while tools registered for another athlete but coach-denied for the selected/per-call athlete should return the existing enumeration-safe target error.
- For read-only write/delete denial in the fake-client scenario, assert the fake intervals client was not called, not just that an error was returned.
- When replacing substring checks with structured parsing, include the existing `list_athletes` `_meta.source` / `active_athlete_id` assertions as structured JSON as well, since that test currently uses string matching.
- Document the catalog-cache caveat in `docs/coach-mode.md` during this step, with the new-conversation/reconnect guidance and TP-040 notification note, even though Step 6 will expand documentation.

With those details carried into implementation, the plan is ready.
