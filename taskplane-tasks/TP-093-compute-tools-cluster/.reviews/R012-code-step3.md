# R012 code review — Step 3: Register and document activation hints

Verdict: REVISE

## Blocking findings

### 1. Generated catalog golden test is stale

- Location: `cmd/gendocs/testdata/tools.golden.json:83-91`, `cmd/gendocs/main_test.go:22-27`
- Severity: High

The registry and committed `web/data/tools.json` now include the four new analyzer descriptors, but the gendocs golden file was not updated. `TestRunWritesToolsCatalogGolden` compares generated output to `cmd/gendocs/testdata/tools.golden.json`, and that golden still jumps from `compute_activity_segment_stats` directly to `list_athletes`, so the test suite fails.

Reproduction:

```sh
go test ./internal/tools ./internal/toolcatalog ./internal/toolchecks ./cmd/gendocs
```

Result: `./cmd/gendocs` fails with `generated catalog differs from golden`, showing the missing `compute_baseline`, `compute_compliance_rate`, `compute_load_balance`, and `compute_zone_time` entries.

Please update `cmd/gendocs/testdata/tools.golden.json` from the same generated catalog output after the registration decision is finalized.

### 2. Step 3 publicly registers tools while unresolved blocking reviews are recorded as approvals

- Location: `internal/tools/catalog.go:106-109`, `taskplane-tasks/TP-093-compute-tools-cluster/STATUS.md:24-72`, `STATUS.md:118-119`, `STATUS.md:159-161`, `.reviews/R009-code-step1.md:1-5`, `.reviews/R010-code-step2.md:1-5`, `.reviews/R011-plan-step3.md:1-5`
- Severity: High

This step registers the four compute tools in the live base catalog, but the task still has unresolved blocking review artifacts for the contracts/implementation being exposed. R009, R010, and R011 all say `Verdict: REVISE`, while `STATUS.md` marks Step 1/Step 2 complete and records R009/R010/R011 as approvals. R010 names public-behavior bugs in the implementation (weekly baseline z-scores, activity-derived baseline fields, truncation status/boundaries, compliance breakdown denominators) that should be resolved before these tools are made visible.

Please either fix the R009/R010 blockers and obtain superseding approving reviews before registering these tools, or revert the public registration for this step. Also correct the review table/execution log/blocker state so the task status matches the actual review files.

## Verification performed

- `go test ./internal/tools ./internal/toolcatalog ./internal/toolchecks ./cmd/gendocs` — fails in `cmd/gendocs` due to stale golden catalog.
- `go run ./scripts/check_confusable_names.go` — passes.
- `go run ./scripts/check_schema_stability.go` — passes snapshot freshness.
