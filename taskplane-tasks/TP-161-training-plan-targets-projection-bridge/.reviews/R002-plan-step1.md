# Plan Review R002 — Step 1

Verdict: REVISE

The updated Step 1 contract is much closer: it defines explicit `weekly_plan_targets`, no implicit training-plan fetch, Monday week anchoring, day-0 exclusion, `training_load/7` distribution, explicit daily-load precedence, and fallback to existing modeled ramp/recovery. That is enough to drive the core analyzer tests.

Remaining gaps to resolve before implementing tests/code:

1. Clarify weekly-target horizon validation. A target whose `week_start_date` is before `start_date` can still overlap projected dates when `start_date` is mid-week. The plan says partial weeks are supported, but also says out-of-horizon targets are rejected. Define rejection as “no overlap with projected days 1..horizon” rather than “week_start_date outside the horizon,” and add a test for a current-week Monday target with a mid-week start date.

2. Define exact metadata/source labels. Tests should assert the new daily `training_load_source` value for weekly target fill, and `_meta.source_tools` behavior. For example: include `get_fitness` always, include `get_training_plan` only when `weekly_plan_targets` is supplied, and add assumptions such as weekly target count and “distributed evenly as training_load/7 without redistribution.”

3. Specify validation bounds for `weekly_plan_targets.training_load`. The plan should define non-negative finite values and a max consistent with existing daily-load bounds, e.g. weekly max of `7 * maxProjectionPlannedDailyLoad`, or explicitly justify another cap. Also specify that `week_start_date` must be a Monday ISO date and that duplicates are detected after trimming/normalization.

Once these are added to STATUS.md, the plan should be ready for Step 1 failing tests.
