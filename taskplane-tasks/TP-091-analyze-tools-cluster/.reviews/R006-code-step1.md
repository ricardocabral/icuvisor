# R006 Code Review — Step 1: Design request/response contracts

**Verdict:** Changes requested

The R005 weekly-derived metric gap is now addressed well enough to implement. I found two remaining contract ambiguities that affect public results, source-window reads, `_meta.n`, and golden fixtures for activity-backed/lagged analyzers.

## Blocking findings

1. **Lagged correlation window boundaries are still undefined.**

   `lag_days` is defined as “x on day D pairs with y on day D+lag” (`STATUS.md:130`), but the contract does not say what happens at the requested window edges. For `window=2026-05-01..2026-05-14` and `lag_days=1`, should the analyzer fetch `metric_y` through 2026-05-15 so every x date in the requested window can pair, or should it only use y values inside the requested window and drop the final day? The same question applies to negative lags at the start of the window.

   Please define one deterministic rule before Step 2, and echo it in `_meta.assumptions`/`_meta.method` so `_meta.n` is interpretable. This also determines whether loaders need an internally expanded read window for one side of the correlation.

2. **Daily aggregation for activity-row rate metrics needs exact weighting semantics.**

   The source mapping says activity-row metrics are day-aggregated for trend and daily correlation, using “sums for additive fields ... and means for rates/intensities (pace, speed, HR, cadence)” (`STATUS.md:137`, `STATUS.md:141`). “Mean” is not specific enough for these metrics and an unweighted activity mean would be misleading on multi-activity days: pace should normally be distance-weighted/derived from total distance and time, speed time-weighted/derived from total distance and time, and HR/cadence are usually duration-weighted if moving time is available.

   Please either define the exact weighted formulas per rate family, or explicitly document that the analyzer uses an unweighted per-activity arithmetic mean and labels `aggregation` accordingly. This choice affects `analyze_trend`, `analyze_correlation pairing_grain=daily`, `_meta.n`, and golden fixtures for pace/speed/HR/cadence metrics.

## Non-blocking note

- The wording “derived weekly metrics use `FitnessClient.ListAthleteSummary` weekly buckets” (`STATUS.md:137`) is still a little confusing because the current client returns daily summary rows and the analyzer creates the weekly buckets. The R005 paragraph clarifies this, but tightening the earlier sentence would reduce implementation drift.

## Tests

Not run; reviewed the Step 1 contract/status diff only.
