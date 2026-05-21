# R014 code review — Step 1: Design deterministic contracts

Verdict: APPROVE

I reviewed the Step 1 contract/status updates against the prior Step 1 blockers. The deterministic-contract gaps called out in R009/R013 are addressed:

- `compute_baseline` now requires `baseline_end_date < current_start_date`, which makes the baseline/current membership and the single activity-candidate cap deterministic.
- `summary_window`, `activity`, and `derived_weekly` baseline sample grains are pinned, including the current-window weekly-bucket mean behavior.
- `compute_load_balance.training_load_total` now has per-row load-source priority and aggregation semantics.
- `compute_compliance_rate` now defines sport/event-type precedence, the auto-pair sport key when both filters are present, one-to-one linked reservation/conflict behavior, event truncation behavior, and the interval-evidence trigger/failure policy.

## Non-blocking follow-ups

1. `STATUS.md` still needs housekeeping before later steps advance: the reviews table stops at R012 even though R013 exists in the execution notes, and `Blockers` says `None` while later-step review artifacts such as R010/R012 are still `REVISE`. That is not a Step 1 contract blocker after this approval, but it can mislead the Step 2/3 workflow.
2. The generated-doc golden remains stale for the Step 3 registration changes. `go test ./cmd/gendocs` still fails with `generated catalog differs from golden` because `cmd/gendocs/testdata/tools.golden.json` lacks the four registered compute tools. Treat this as a Step 3/doc freshness blocker, not a Step 1 contract blocker.

## Verification performed

- `go test ./internal/tools ./internal/toolcatalog ./internal/toolchecks` — passes.
- `go test ./cmd/gendocs` — fails due to stale generated catalog golden, as noted above.
