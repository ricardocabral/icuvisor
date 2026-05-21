# R023 plan review — Step 5: Testing & Verification

Verdict: REVISE

The Step 5 checklist points at the right final gates, but it is too generic for this task's current state. Step 5 is where known quality-gate failures must be resolved, not merely discovered, and the plan should name the exact targeted command set and the known lint blockers carried forward from R022.

## Blocking findings

### 1. Known TP-093 lint blockers are not called out as work to fix

- Location: `STATUS.md` Step 5 checklist; `.reviews/R022-code-step4.md`
- Severity: High

R022 explicitly recorded `golangci-lint run ./internal/tools` failures in the compute-tool implementation files:

- `internal/tools/compute_baseline.go`: `unparam` for `metric` in `collectWellnessBaseline` and `collectActivityBaseline`
- `internal/tools/compute_zone_time.go`: `unparam` for `metric` in `loadValueForZoneMetric`, plus unused `_compileComputeZoneFmtUse`
- `internal/tools/compute_compliance_rate.go`: unused `linkedActivityForEvent`

These are not unrelated external failures for Step 5 purposes; they are in this task's implementation surface and will block `make lint`. The plan should explicitly include fixing these known issues before or while running the full lint gate, and it should not leave room to document them as "pre-existing unrelated failures."

### 2. The targeted verification command is underspecified

- Location: `STATUS.md` Step 5 checklist
- Severity: Medium

"Targeted tests passing" is not reviewable enough for this L-sized analyzer cluster. Step 4 settled on a concrete affected-package command, and Step 5 should repeat it so the worker records an exact command/outcome before running the full suite:

```sh
go test -count=1 ./internal/analysis ./internal/tools ./internal/toolcatalog ./internal/toolchecks ./internal/safety ./cmd/gendocs
```

If any fixes are made for lint or test failures during Step 5, rerun this targeted command after the fix, then run `make test`, `make build`, and `make lint`.

## Suggested plan adjustment

Revise Step 5 to include:

1. Fix the known R022 lint blockers in TP-093 compute-tool files.
2. Run and record the targeted command above.
3. Run and record `make test`.
4. Run and record `make build`.
5. Run and record `make lint`.
6. Fix all failures in changed/task-owned code; document only genuinely pre-existing unrelated failures with command output and rationale.

## Non-blocking suggestion

Because this repository requires gofmt/goimports cleanliness and generated docs/catalog outputs have changed in earlier steps, consider also recording a formatting/dirty-diff check after fixes, e.g. `make fmt-check` and a focused generated-doc diff check if `make docs-tools` is rerun.
