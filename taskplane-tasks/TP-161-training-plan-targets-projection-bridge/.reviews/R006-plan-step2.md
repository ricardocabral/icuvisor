# Plan Review R006 — Step 2

Verdict: APPROVED

The Step 2 plan is ready to implement. The approved Step 1 contract in `STATUS.md` plus the newly added failing tests give enough detail for the bridge implementation:

- Add `weekly_plan_targets` as a typed optional `get_fitness_projection` request field and pass normalized targets into `analysis.FitnessProjectionInput`.
- Keep the tool explicit/read-only: do not fetch `get_training_plan` from inside the projection handler.
- Validate weekly targets in the decoder: trimmed Monday `week_start_date`, duplicate normalized anchors rejected, overlap with projected days 1..horizon, and `training_load` bounded to `0..7*maxProjectionPlannedDailyLoad`.
- Implement deterministic `training_load/7` fill for matching future dates, with day 0 excluded and no redistribution for partial weeks.
- Preserve precedence: `planned_daily_loads` override exact-date weekly-target candidates; uncovered dates keep modeled ramp/recovery sources.
- Carry enough analyzer result metadata to report weekly target count, filled-day count, override count, source labels, assumptions, boundaries, and `source_tools` including `get_training_plan` when non-empty weekly targets are used.
- Refresh the schema snapshot and run the targeted tests before moving to Step 3.

Minor implementation note: treat `weekly_plan_targets: []` the same as no weekly targets for fill counts and `source_tools`; if you choose different behavior, add an explicit test and note it in `STATUS.md`.
