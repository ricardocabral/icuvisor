# Code Review — TP-087 Step 2: Implement validation helpers

## Verdict: REVISE

The new `internal/analysis` package is generally shaped correctly for the shared enum/schema contract, but there is one blocking validation bug: some canonical enum values emitted by `MetricValues()` and the schema are rejected by `ParseMetric()` as expressions.

## Findings

### Blocking

1. **Canonical `*_per_*` metrics are rejected as expressions**  
   **File:** `internal/analysis/metrics.go:95`, `internal/analysis/metrics.go:314`

   `ParseMetric` calls `looksLikeExpression` before checking the canonical/alias map. `looksLikeExpression` treats any input containing `_per_` as an expression, but the catalog intentionally includes canonical metrics such as:

   - `pace_seconds_per_km`
   - `pace_seconds_per_mile`

   As a result, those metrics appear in `MetricValues()` / `MetricSchemaProperty()["enum"]`, but a tool receiving the exact canonical value will return `invalid analysis_metric: expressions are not supported; choose a supported metric`.

   This breaks the closed enum contract and the Step 1 design record, which explicitly included those pace metrics while rejecting unsupported formula-like names such as `tss_per_hour`.

   Suggested fix: check the canonical/alias map before expression rejection, then run expression detection only for unknown values. Make `_per_` detection case-insensitive on the normalized input, and add Step 3 coverage that every `MetricValues()` entry successfully round-trips through `ParseMetric`.

## Verification

- Ran `go test ./internal/analysis` — passed (`[no test files]`).
- Ran `go test ./...` — passed.

## Notes

- Defensive copying for `MetricSources` and canonical-only schema enum behavior look good.
- The current `ValidateMetricCatalog` sortedness check is effectively redundant because `MetricValues()` sorts before returning. That is not blocking for Step 2, but Step 3 tests should prefer asserting deterministic `MetricValues()` output and parse round-tripping rather than relying on this helper to catch catalog-order drift.
