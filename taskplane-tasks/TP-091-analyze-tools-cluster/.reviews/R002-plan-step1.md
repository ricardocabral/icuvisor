# R002 Plan Review — Step 1: Design request/response contracts

**Verdict:** Revise

The updated Step 1 notes are a large improvement over R001: they now define per-tool inputs, defaults, terse/full response behavior, analyzer meta usage, and a no-raw-stream source strategy. I would still hold Step 2 until a few public-contract ambiguities are resolved, because these affect schema shape, fixture design, and user-visible interpretation.

## Blocking findings

1. **`analyze_correlation` response is missing PRD-required regression fields.**
   PRD §7.2.C defines `analyze_correlation` as returning Pearson/Spearman, `n`, **slope**, and **intercept**. The Step 1 response sketch only includes coefficient/direction/strength and optional p-value. Add `slope` and `intercept` (or explicitly document a deliberate PRD deviation before implementation). Also state whether slope/intercept are computed on raw paired values for both methods, only for Pearson, or omitted for Spearman with a separate reason field.

2. **Correlation source/grain selection is still ambiguous and internally conflicting.**
   The plan says `pairing_grain` defaults from source compatibility, but the common source loader chooses the first supported metric source in a fixed priority. Those rules conflict for metrics with multiple sources. Example: `training_load` has activity-row and training-summary sources; with `hrv` vs `training_load`, a daily correlation should select/derive daily `training_load`, not pick activity rows and then fail or pair incorrectly. Before implementation, define deterministic rules for:
   - requested `pairing_grain: daily` vs `activity`;
   - mixed daily/activity metrics, including whether activity metrics are aggregated to athlete-local days or rejected;
   - sport filter behavior when only one side of a correlation is sport/category-backed;
   - tie-breaking when both metrics have several compatible sources.

3. **Minimum sample rules are not concrete enough for all four tools.**
   The plan references baseline sufficiency and mandatory `_meta.insufficient_sample`, but it does not specify exact `MinSamples` for trend, distribution, efforts-delta, or each correlation mode. PRD gives `n>=14` for correlation and `n>=7` for baseline analyzers; the implementation needs exact rules for:
   - current-window trend slope/rolling mean;
   - baseline delta/z-score in trend;
   - distribution stats/histogram/quantiles;
   - efforts buckets when current or baseline buckets are missing;
   - what `n` means for efforts-delta (comparable buckets? current curve points? activities are not available from curve summaries).

4. **Efforts-delta bucket validation and defaults need one more pass.**
   The request sketch says power/HR use `duration_seconds` defaults, while pace requires `distance_meters` and does not use sport defaults. That is mostly clear, but the public contract should explicitly say:
   - `duration_seconds` is invalid for `pace` and `distance_meters` is invalid for `power|heart_rate`;
   - pace has no default distance buckets, or define the default if one is intended;
   - max bucket count / max bucket value, or intentionally no cap;
   - how missing current vs missing baseline buckets appear in `result` and `_meta.insumufficient_sample`.

5. **Silent rolling-window clipping should be replaced with reject-or-report semantics.**
   `rolling_window_days` is described as being clipped to the current window length by validation. Silent mutation of a caller's requested formula makes `_meta.method` and tests harder to reason about. Prefer rejecting values larger than the window, or explicitly return the effective value in `_meta.assumptions` and method text so the LLM cannot describe the wrong calculation.

## Non-blocking suggestions

- Add a small response-field table to `STATUS.md` for each tool before coding. The prose is good, but exact field names such as `window_mean`, `rolling_latest_mean`, `percent_delta`, `histogram_buckets`, `coefficient`, etc. will prevent churn in golden tests.
- For `analyze_distribution`, explicitly state whether `buckets` are bucket boundaries or full `{min,max}` ranges, and whether the intervals are `[lower, upper)` except the final bucket.
- For `formula_ref`, consider saying "empty unless the exact anchor exists in `icuvisor://analysis-formulas`" and use the resource constants from `internal/resources` rather than hard-coded strings when implemented.
- For sport filters, define case/whitespace normalization to match existing read tools and avoid analyzer-specific surprises.

Once the correlation contract, sample-size rules, and efforts bucket semantics are tightened, the plan should be ready for implementation.
