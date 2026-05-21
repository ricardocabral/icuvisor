# R018 Code Review — Step 5: Testing & Verification

**Verdict:** REVISE

The requested quality gates pass, and the Step 4 findings are mostly addressed. I found one remaining regression in the invalid-value handling added for trend slope.

## Findings

1. **Daily trend slope now uses sparse original sample indexes after invalid values are skipped.**  
   `internal/analysis/trend.go:108-117` filters `NaN`/`Inf` values but still uses the original loop index as the x-value for non-weekly samples. The TP-091 contract says daily trend OLS uses dense usable-sample indexes `0..n-1`, while only weekly metrics use bucket indexes. Before this change, daily slopes were computed against `Values(samples)` with dense indexes; now a daily series like `[100, NaN, 120, 140]` is regressed as x=`[0,2,3]` instead of x=`[0,1,2]`, changing the slope from `20` to `12.857`. Build the daily x-value from the count of accepted finite samples (for example `float64(len(values))` before appending) and keep using `sample.Bucket` for weekly samples. Add a daily invalid-value regression test alongside the new weekly one.

## Tests run

- `go test ./internal/analysis ./internal/tools`
- `make test`
- `make build`
- `make lint`
