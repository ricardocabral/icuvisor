# R004 Plan Review — Step 1: Design request/response contracts

**Verdict:** Approved for Step 2

The Step 1 notes in `STATUS.md` now cover the public request shapes, baseline/window rules, analyzer meta contract, source-client mapping, sample-size thresholds, correlation grain/lag semantics, efforts bucket validation, and activity pagination/completeness rules. The R001–R003 blockers are substantively resolved, and the plan is detailed enough to start implementation without inventing broad public behavior in code.

## What is now ready

- Each of the four analyzer tools has a request contract with required/optional fields, defaults, validation bounds, and terse/full behavior.
- The plan consistently routes metrics through `analysis.ParseMetric`, `analysis.MetricSchemaProperty`, and `analysis.NewAnalyzerMeta`/`shapeAnalyzerResponse`.
- Unsupported source families are explicitly rejected rather than silently fan-out fetching raw streams or one-activity extended metrics.
- Correlation now has `coefficient`, `slope`, `intercept`, `regression_method`, deterministic daily/activity pairing rules, and `lag_days=0` for activity-grain pairing.
- Activity-backed analyzers now require full-window pagination and fail closed on pagination-boundary/partial-read conditions.
- Efforts-delta is separated from `analysis_metric`, uses the curve endpoints, and has closed bucket validation.

## Non-blocking implementation cautions

Please pin these during Step 2/Step 4 so golden tests do not encode accidental choices:

1. **Short trend windows vs default rolling window.** The plan says `rolling_window_days` defaults to 7 and values greater than the current window length are rejected. That means a trend request for a window shorter than 7 days is invalid unless a smaller rolling window is supplied, while the sample-size rule also allows `_meta.insufficient_sample=true`. Implement one behavior consistently and cover it in validation tests.

2. **Formula conventions.** For deterministic golden tests, explicitly choose and test the exact variants for OLS slope, quantile interpolation/nearest-rank behavior, histogram boundary inclusion, correlation strength labels, and sample vs population standard deviation. The z-score formula should use the existing `analysis-formulas` sample-standard-deviation definition.

3. **Pace efforts output.** Make the pace-family result fields unit-explicit, following the existing curve tools: upstream elapsed seconds plus athlete-preferred `pace_seconds_per_km` or `pace_seconds_per_mile` where applicable. Keep `absolute_delta`/`percent_delta` semantics clear for lower-is-better elapsed/pace values.

4. **Lag window boundaries.** Add tests for non-zero `lag_days` at the start/end of the requested window so it is clear whether pairs outside the requested range are ignored or fetched via an expanded internal read. Do not silently change `_meta.n` interpretation.

5. **Source tool names in `_meta.source_tools`.** For efforts-delta, report the user-facing read-tool surface (`get_power_curves`, `get_hr_curves`, or `get_pace_curves`) rather than only internal client method names.

These are not blockers to beginning Step 2, but they should be resolved in code and fixtures before the task is marked complete.
