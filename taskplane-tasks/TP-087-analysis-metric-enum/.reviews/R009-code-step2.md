# Code Review — TP-087 Step 2: Implement validation helpers

## Verdict: REVISE

The helper package is on the right path, but the previous blocking parser/schema inconsistency is still present in `HEAD`. `STATUS.md` marks the `_per_` fix as complete, but the code still rejects canonical metrics that are emitted by `MetricValues()` and the JSON schema enum.

## Findings

### Blocking

1. **Canonical schema values containing `_per_` still fail `ParseMetric`**  
   **File:** `internal/analysis/metrics.go:95`, `internal/analysis/metrics.go:206`, `internal/analysis/metrics.go:318`

   `ParseMetric` calls `looksLikeExpression(trimmed)` before consulting `aliasToMetric`. Because `looksLikeExpression` treats any `_per_` token as an expression, canonical catalog entries such as `pace_seconds_per_km` and `pace_seconds_per_mile` are included in `MetricValues()` / `MetricSchemaProperty()["enum"]` but rejected when parsed.

   Reproduction from this checkout:

   ```text
   pace_seconds_per_km => "" err=invalid analysis_metric: expressions are not supported; choose a supported metric
   pace_seconds_per_mile => "" err=invalid analysis_metric: expressions are not supported; choose a supported metric
   ctl => "ctl" err=<nil>
   ```

   This violates the closed enum contract: every canonical value advertised by the schema must be accepted by the parser. Fix by normalizing, checking `aliasToMetric` first, and only running expression detection for unknown values. Step 3 should include a round-trip test over all `MetricValues()` entries.

2. **Task status claims the blocking fix is complete when it is not**  
   **File:** `taskplane-tasks/TP-087-analysis-metric-enum/STATUS.md:59`

   The Step 2 checklist marks “Fix `ParseMetric` so canonical values containing `_per_` parse before expression rejection” as done, but the implementation has not changed accordingly. This can mislead subsequent workers/reviewers and hides the remaining blocker. Either apply the fix in code or leave this item unchecked until it actually lands.

## Verification

- Ran `git diff 74abeda..HEAD --name-only` and `git diff 74abeda..HEAD`.
- Read `PROMPT.md`, `STATUS.md`, and `internal/analysis/metrics.go`.
- Ran `go test ./internal/analysis ./...` — passed, but there are currently no `internal/analysis` tests.
- Ran a temporary in-module `go run` snippet to call `analysis.ParseMetric` for representative values; confirmed the `_per_` canonical metrics still fail.
