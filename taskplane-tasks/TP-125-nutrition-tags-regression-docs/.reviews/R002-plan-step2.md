# Plan Review: Step 2 — Fill regression or docs gaps

Verdict: Needs changes

The Step 2 plan is directionally right, but it inherits an audit gap that would leave the task's core requirement under-tested: activity tags must be present in normal/terse reads without requiring `include_full:true`.

Blocking plan gap:
- `get_activities` tag coverage currently exercises `TestGetActivitiesTagEdgeCases` with `"include_full":true`, so it does not prove tags survive the default terse response.
- `get_activity_details` tag coverage likewise uses `"include_full":true` in `TestGetActivityDetailsTagsPreserveOrderAndFullPayload`.
- `STATUS.md` only records a missing empty-tags regression for `get_activity_details`, but Step 2 should also explicitly add default/no-`include_full` regressions for present tags and explicit empty tag arrays for both `get_activities` and `get_activity_details` (or refactor existing table tests to cover both default and include-full paths).

Required adjustment before execution:
- Expand Step 2's regression-test plan/discoveries to cover terse/default tag preservation separately from raw `full` payload preservation.
- Keep the existing nutrition tests focused on default reads; they already appear to cover `carbs_ingested_g` and `carbs_used_g` without `include_full:true`.
- Include the user-facing docs update and `CHANGELOG.md` update in the Step 2/Step 4 handoff so the Documentation Requirements are not missed.

Once those items are in the plan, the rest of the step scope is appropriate: add regression tests only, keep upstream raw names inside `full`, retain the disambiguated `_g` fields in terse output, and run `go test ./internal/tools`.
