# Plan Review R003 — Step 1

Verdict: APPROVED

The Step 1 contract in `STATUS.md` now resolves the prior blocking gaps and is specific enough to encode as failing tests before implementation.

What is ready:

- Defines an explicit `weekly_plan_targets` input shape with no implicit `get_training_plan` fetch.
- Specifies Monday/ISO week anchoring, athlete-local dates, day-0 exclusion, and partial-week behavior.
- Defines the deterministic conversion as `training_load / 7` with no redistribution.
- Establishes precedence: exact-date `planned_daily_loads` override weekly-target candidates, while unrelated dates keep existing modeled ramp/recovery behavior.
- Handles mid-week current-week targets via projected-day overlap validation.
- Defines validation bounds for week dates, duplicates, finite non-negative weekly load, and max weekly load tied to the existing daily cap.
- Specifies daily source labels and `_meta.source_tools`/assumptions expectations.

Proceed with Step 1 failing tests. Please make the tests lock down the edge cases already captured in the status note, especially:

1. current-week Monday target with `start_date` mid-week;
2. explicit daily override of a weekly-target date, with no redistribution;
3. dates outside weekly coverage falling back to `modeled_ramp`/`modeled_recovery_week`;
4. duplicate normalized week anchors, invalid/non-Monday dates, no-overlap targets, and out-of-range/non-finite loads;
5. `_meta.assumptions`, `_meta.source_tools`, and `training_load_source` values.

Minor implementation note for later: define whether an empty `weekly_plan_targets: []` counts as “supplied” for `_meta.source_tools`; tests should make that behavior explicit.
