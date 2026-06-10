# Code Review R007 — Step 2

Verdict: REVISE

## Findings

1. **P2 — Generated tool-schema docs are stale, causing full tests to fail.**  
   `internal/tools/get_fitness_projection.go:380-381` changes the public `get_fitness_projection` input schema by updating `planned_daily_loads` and adding `weekly_plan_targets`, but the generated schema catalog goldens/web data were not refreshed. `cmd/gendocs/testdata/tool_schemas.golden.json` still has the old planned-load description and no `weekly_plan_targets` argument, and `web/data/tool_schemas.json` is stale too. This makes `go test ./...` fail in `cmd/gendocs` (`TestRunWritesGeneratedDocsGolden`) and leaves the published tool reference without the new bridge argument. Regenerate/update the gendocs outputs/golden after the schema change.

## Verification

- `go test ./internal/analysis ./internal/tools -run 'FitnessProjection|TrainingPlan'` — pass
- `go test ./internal/tools` — pass
- `go test ./internal/toolchecks` — pass
- `go test ./...` — fail: `cmd/gendocs TestRunWritesGeneratedDocsGolden` due stale generated docs golden
