# Code Review: Step 3 — Propagate to analyzers

## Verdict

Approve.

## Review scope

- Reviewed diff from `c400815e461201be64763488d4f78b163591a4ea..HEAD`.
- Read the TP-090 prompt/status and the changed analyzer metadata/test files.
- Checked the analyzer shaping path in `internal/tools/analyzer_common.go` and interval-source types in `internal/analysis/interval_source.go` for context.

## Findings

No blocking findings in the Step 3 changes.

The implementation adds the shared analyzer helpers in `internal/analysis`, keeps interval-specific analyzer metadata optional, appends/deduplicates `get_activity_intervals` through the existing `NewAnalyzerMeta` normalization path, and provides the policy reason `auto_lap_suspected` for downstream analyzer tasks. The placeholder tests cover positive propagation, non-interval omission, source-tool dedupe, and the execution-claim decline policy.

## Verification

Passed:

```sh
go test ./internal/analysis ./internal/tools
go test ./...
```

Quality-gate note:

```sh
golangci-lint run ./internal/analysis ./internal/tools
```

currently fails on `internal/analysis/interval_source.go:151` with staticcheck `QF1001` (De Morgan simplification). That file is outside the Step 3 diff, so I am not treating it as a blocker for this review, but it will need cleanup before the task's lint gate is marked complete.

## Non-blocking notes

- A future Step 4/TP-091 test for evaluated `IntervalSourceUnknown` propagation would make the “evaluated unknown vs. no interval evidence” contract even more explicit, but the current code path should already emit `_meta.interval_source: "unknown"` plus `_meta.auto_lap_suspected: false` when `ApplyIntervalSourceEvidence` is called with the classifier's unknown result.
