# R012 Code Review — Step 2: Implement computations

**Verdict:** REVISE

The new computation helpers compile and `go test ./...` passes, but there are correctness issues in the analyzer math/source helpers that should be fixed before Step 2 is considered complete.

## Findings

1. **Daily activity pace aggregation does not follow the approved R006 semantics.**  
   `internal/tools/analyzer_sources.go:265-291` sends `pace_seconds_per_km` / `pace_seconds_per_mile` through `weightedActivityMean` (`internal/tools/analyzer_sources.go:294-318`). That computes a moving-time-weighted mean of per-activity paces, but R006 requires daily pace to be derived as `total moving seconds / total distance` for the day. These differ whenever activities have different paces; e.g. 1 km in 300 s plus 9 km in 1800 s should be 210 s/km, while the current weighted mean returns about 214.286 s/km. This will make trend/distribution/correlation results for daily pace wrong. Add explicit daily pace aggregation from summed usable distance/time, and label the assumption accordingly. Consider doing the same explicit total-distance/total-time derivation for average speed to match the contract rather than trusting/averaging upstream speed fields.

2. **Weekly trend slope ignores sample bucket indexes.**  
   `internal/analysis/trend.go:57-59` calls `olsSlope(values)`, and `olsSlope` always uses compressed `0..n-1` x values (`internal/analysis/trend.go:105-112`). The Step 2 numeric plan says daily trends use `0..n-1` but weekly trends use weekly bucket indexes. With a missing weekly bucket, samples for buckets 0 and 2 are treated as adjacent buckets 0 and 1, overstating the per-week slope. The trend computation should retain/use `NumericSample.Bucket` when `SampleGrainWeekly` is selected, while keeping the current behavior for daily/activity series.

3. **Correlation reports `n` before dropping invalid pairs.**  
   `internal/analysis/correlation.go:52-53` sets `N: len(input.Pairs)` and only then filters NaN/Inf pairs via `pairValues`. As a result, a response/meta layer that uses `result.N` can mark the sample as sufficient even though the coefficient/regression used fewer pairs. This is inconsistent with `Values`/`Stats` and the missing-data contract. Set `N` from the filtered pair count and ensure insufficient-sample checks use that same usable count.

## Tests run

- `go test ./...`
