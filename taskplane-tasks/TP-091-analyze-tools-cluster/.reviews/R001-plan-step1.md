# R001 Plan Review — Step 1: Design request/response contracts

**Verdict:** Changes requested

The task context is clear, but the current Step 1 "plan" is only the original checklist in `STATUS.md`; it does not define the request/response contracts it says will be designed. For a review-level 3 task adding four public MCP tools, this needs a concrete contract sketch before Step 2 implementation starts.

## Blocking findings

1. **No concrete public schemas are specified.**
   Step 1 must define the exact input properties for each tool, not just list categories. At minimum, capture required/optional fields, defaults, validation bounds, and JSON Schema descriptions for:
   - `analyze_trend`: metric, current window, baseline window/default, rolling window, sport filter, grain/source behavior, `include_full`.
   - `analyze_distribution`: metric, window, bucket strategy/histogram vs quantiles, sport filter, whether time-in-zone is supported only from precomputed fields, `include_full`.
   - `analyze_correlation`: metric_x/metric_y, window, method (`pearson`/`spearman`), pairing grain (`daily`/`activity`), lag direction and bounds, sport filter, `include_full`.
   - `analyze_efforts_delta`: effort type/bucket schema (duration seconds and/or distance meters), current/baseline windows, sport, unit behavior, `include_full`.

2. **Metric/source compatibility rules are not planned.**
   The implementation must use `internal/analysis.ParseMetric`, `MetricSchemaProperty`, and `MetricSources`, but the plan does not say which source families are valid per analyzer. This matters because metrics span daily wellness/fitness, per-activity rows, training-summary windows, extended metrics, and interval rows. The plan should define deterministic source selection and rejection behavior for unsupported combinations instead of allowing ad-hoc field extraction.

3. **Response contracts are missing.**
   The plan needs the output shape for each tool: headline `result` fields, what appears in `series` only with `include_full:true`, and all mandatory analyzer `_meta` fields. Include decisions for `_meta.method`, `_meta.source_tools`, `_meta.n`, `_meta.missing_days`, `_meta.missing_action`, `_meta.insufficient_sample`, `formula_ref`, units, and scale legends for subjective metrics. This is part of the public API and should be reviewed before math is implemented.

4. **Window semantics are underspecified.**
   Define whether callers pass `start_date`/`end_date`, `window_days`, or both; whether ranges are inclusive; max window sizes; how baseline windows default; and that dates are interpreted in the athlete timezone at the presentation/request boundary. Missing-day counting depends on this contract.

5. **Correlation and lag semantics need explicit design.**
   The plan must state how `lag_days` is applied (e.g. positive lag means metric_x on day D is paired with metric_y on D+lag), how pairs are built for daily vs per-activity metrics, and how minimum `n>=14` is enforced. Otherwise tests can pass while user-facing interpretation is ambiguous.

6. **Efforts-delta needs a separate effort contract.**
   Best-effort durations/distances are intentionally not general `analysis_metric` values; `ParseMetric` hints users toward `analyze_efforts_delta` for these. Step 1 should define a closed effort-bucket schema and source clients (`get_best_efforts`, power/HR/pace curves as applicable) rather than trying to squeeze efforts into `analysis_metric`.

## Required plan additions before proceeding

Please update `STATUS.md` notes or add a Step 1 design note with:

- A per-tool request struct sketch and JSON Schema property list.
- A per-tool response struct sketch, including terse vs `include_full` behavior.
- Source-client mapping for every analyzer, preferring existing aggregate reads and explicitly avoiding raw streams for this task.
- Metric compatibility/rejection rules using the closed enum and one-line user-facing hints.
- Window/baseline/lag definitions, defaults, limits, and sample-size rules.
- Unit and scale metadata decisions, especially for preferred pace/distance units and subjective scales.

Once those contracts are captured, the implementation plan will be reviewable and should reduce churn in Step 2/Step 3.
