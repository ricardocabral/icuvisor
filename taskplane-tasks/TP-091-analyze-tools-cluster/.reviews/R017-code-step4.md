# R017 Code Review — Step 4: Tests and verification

**Verdict:** REVISE

The targeted and full Go test suites pass, and the previously identified R012/R015 regressions are mostly addressed. I found two remaining issues that should be fixed before Step 4 is considered complete.

## Findings

1. **Weekly trend slope can panic when samples contain invalid numeric values.**  
   `internal/analysis/trend.go:105-114` builds `values := Values(samples)`, which filters out `NaN`/`Inf`, then allocates `xs` with `len(values)`. For weekly grain it loops over the original `samples` slice and writes `xs[i] = ...`. If any sample was filtered out, `len(samples) > len(values)` and the loop can index past the end of `xs`. The other analysis helpers intentionally tolerate invalid numeric inputs, so `ComputeTrend` should return an unavailable slope/boundary rather than panic. Build the x/y pairs in one pass, appending a bucket index only for finite sample values (or otherwise keep `xs` and `values` aligned), and add a regression test with a weekly `NaN`/`Inf` sample.

2. **`analyze_trend` reports the unevaluated rolling-window default in `_meta.assumptions`.**  
   `internal/tools/analyze_trend.go:94-98` sets `assumptions["rolling_window_days"] = args.RollingWindowDays`. When callers omit the field, the handler computes with the default `rolling = 7`, but the public analyzer metadata reports `rolling_window_days: 0`. This violates the accepted contract that the effective rolling window is echoed in metadata and makes default responses self-contradictory. Store the effective day value before converting weekly windows to bucket counts and report that value; keep `rolling_bucket_count` for weekly metrics.

## Tests run

- `go test ./internal/analysis ./internal/tools ./internal/safety ./internal/toolcatalog ./cmd/gendocs`
- `make test`
