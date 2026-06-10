# Code Review R008 — Step 2

Verdict: REVISE

## Findings

1. **P2 — `weekly_plan_targets` accepts an omitted `training_load` as a zero-load week.**  
   `internal/tools/get_fitness_projection.go:62-64` models `TrainingLoad` as a plain `float64`, and `validateProjectionWeeklyPlanTargets` only range-checks the decoded value at `internal/tools/get_fitness_projection.go:361-363`. Because `DecodeStrict` does not enforce JSON Schema `required`, a request like `{"start_date":"2026-06-03","horizon_days":3,"weekly_plan_targets":[{"week_start_date":"2026-06-01"}]}` decodes successfully with `TrainingLoad == 0` and then projects the overlapping days as `weekly_plan_targets` zero load. The schema marks `training_load` as required, and silently converting a malformed weekly target into a recovery/drop week undermines the bridge. Make the field presence-checkable (for example `*float64` or custom raw validation) and reject missing `weekly_plan_targets[].training_load` before distribution.

## Verification

- `go test ./internal/analysis ./internal/tools -run 'FitnessProjection|TrainingPlan'` — pass
- `go test ./cmd/gendocs ./internal/tools ./internal/toolchecks` — pass
- `go test ./...` — pass
- Ad hoc decoder check confirmed a weekly target missing `training_load` is currently accepted as `0`.
