# Code Review — TP-087 Step 2: Implement validation helpers

## Verdict: APPROVE

The Step 2 helper implementation now satisfies the closed `analysis_metric` parser/schema contract. The previous `_per_` blocker is fixed: canonical schema values are checked through the alias map before expression rejection, so `pace_seconds_per_km` and `pace_seconds_per_mile` round-trip successfully while unknown formula-like values such as `tss_per_hour` are still rejected as expressions.

## Findings

No blocking findings.

## Notes

- `MetricValues()` returns a deterministic, canonical-only enum and `MetricSchemaProperty()` uses that list directly.
- `MetricSources()` returns defensive copies and supports aliases via `ParseMetric`, which keeps future analyzer routing reusable without MCP/tool coupling.
- `ValidateMetricCatalog()`'s sortedness check is effectively redundant because it calls `MetricValues()`, which sorts its output before returning. This is not blocking for Step 2, but Step 3 tests should include the more important invariants: every `MetricValues()` entry parses back to itself, aliases parse to canonical values, and expression/unknown hints remain concise.

## Verification

- Ran `git diff 74abeda..HEAD --name-only` and `git diff 74abeda..HEAD`.
- Read `PROMPT.md`, `STATUS.md`, and `internal/analysis/metrics.go`.
- Ran `go test ./internal/analysis ./...` — passed.
- Ran a temporary in-module `go run` check over all `analysis.MetricValues()` entries; all canonical values parsed successfully. Spot checks confirmed `pace_seconds_per_km` parses and `ctl/atl` / `tss_per_hour` are rejected with the expression hint.
