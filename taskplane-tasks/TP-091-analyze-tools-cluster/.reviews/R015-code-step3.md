# R015 Code Review — Step 3: Register tools and descriptions

**Verdict:** REVISE

The four analyzer tools are registered in the full toolset and their descriptions have the requested activation hints, but the step currently leaves failing repository tests and a functional weekly-metric regression in the new adapters.

## Findings

1. **Full test suite fails after registering the new tools.**  
   `go test ./...` fails in `cmd/gendocs` (`TestRunWritesToolsCatalogGolden`) because the generated catalog golden has not been updated, and in `internal/safety` (`TestAdversarialStaticCatalogMatrix`) because the static safety catalog/counts do not include the four new read tools. This is directly caused by Step 3 registration, not an unrelated failure. Update the generated tool catalog artifacts/golden and add `analyze_trend`, `analyze_distribution`, `analyze_correlation`, and `analyze_efforts_delta` to the safety adversarial catalog with `RequirementRead`.

2. **`weekly_tss` / `weekly_hours` are registered as supported analyzers but never load data.**  
   `loadAnalyzerSeries` selects `analysis.SourceDerivedWeekly` for derived weekly metrics when `allowWeekly=true`, but the switch at `internal/tools/analyze_common.go:54-69` only handles `SourceFitnessDaily`/`SourceTrainingSummary` before checking `metric == "weekly_tss" || metric == "weekly_hours"`. Because `SourceDerivedWeekly` has no case, the function returns an empty series with no error. As a result `analyze_trend` and `analyze_distribution` for the explicitly approved weekly metrics produce insufficient/empty results instead of fetching summary rows and building weekly buckets. Add a `SourceDerivedWeekly` case (or include it in the summary-loading case) and set the weekly assumptions there.

3. **Some cancellation errors are converted into user-facing fetch errors.**  
   The first series load in `analyze_correlation` preserves `context.Canceled`/`DeadlineExceeded`, but the second load at `internal/tools/analyze_correlation.go:90-92` wraps every error in `NewUserError`. `analyze_trend` does the same for the baseline load at `internal/tools/analyze_trend.go:82-84`. Repository guidance requires honoring cancellation; mirror the cancellation check used for the first/current load before wrapping these errors.

4. **`analyze_distribution` does not enforce its request contract for buckets/quantiles.**  
   The approved contract says explicit `buckets` are mutually exclusive with `bucket_count`, and quantiles must be in `0..1`. The handler defaults/validates `bucket_count` at `internal/tools/analyze_distribution.go:40-48`, then passes both `BucketCount` and `Buckets` through at line 77, while invalid quantiles are silently skipped by `ComputeDistribution`. Please reject `bucket_count` when `buckets` is supplied and return an invalid-argument user error for out-of-range quantiles so callers do not get partial, surprising results.

## Tests run

- `go test ./internal/tools ./internal/toolcatalog ./internal/analysis` — pass
- `go test ./...` — fail (`cmd/gendocs` catalog golden, `internal/safety` static catalog matrix)
