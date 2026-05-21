# R009 Code Review — Step 1: Design request/response contracts

**Verdict:** Approved for Step 2

The R008 follow-up is resolved: `analyze_efforts_delta` now has a deterministic family-specific `_meta.source_tools` contract matching the curve endpoints it will read (`get_power_curves`, `get_hr_curves`, or `get_pace_curves`). The Step 1 notes now cover the public request/response shapes, window/baseline defaults, source compatibility, activity pagination completeness, weekly-derived metrics, lag semantics, weighted aggregation, efforts units, terse/full behavior, and mandatory analyzer metadata well enough to begin implementation.

## Blocking findings

None.

## Non-blocking implementation cautions

- Pin exact source-selection priority in code/tests for metrics with multiple daily-capable sources (for example `ctl`/`atl` via fitness vs wellness, and activity-vs-summary metrics). The current contract is sufficient if implementation consistently follows catalog order or another documented priority, but golden tests should make the chosen `_meta.source_tools` deterministic.
- Pin exact formula variants in analyzer tests before marking the task complete: OLS slope convention, Pearson/Spearman tie handling, quantile interpolation, histogram boundary inclusion, and distribution sample-vs-population standard deviation.
- Clean up process hygiene when advancing the step: update `Review Counter`, add the R009 row under `## Reviews`, mark Step 1 complete if proceeding, and move the stray `| 2026-05-20 19:39 | Review R008 | ... |` line out of `## Notes` into `## Execution Log`.

## Tests

Not run; reviewed the Step 1 contract/status diff only.
