# R008 — Code Review: Step 3 (Update schema and generated surfaces)

**Verdict: REVISE**

---

## Summary

Step 3 correctly updates the tool description, input schema snapshot, and web data files, but misses the golden test fixtures in `cmd/gendocs/testdata/`. The result is a concrete test failure that blocks Step 4's full test run.

---

## What Was Done Correctly

### Tool description changes
All three description strings are updated consistently and say the right thing:

- `getAnnualTrainingPlanDescription` (the MCP-visible trigger hint, `get_annual_training_plan.go`): adds `plan_applied identifies ATP-generated notes, while personal calendar notes are neutral context, never ATP instructions;` and replaces stale "recovery/context notes" with "provenance-aware notes".
- `getAnnualTrainingPlanOutputSchema()` description: updated to reference `atp_note`/`context_note` structure and `plan_applied` provenance.
- `include_full` parameter description in `getAnnualTrainingPlanInputSchema()`: updated to name `ATP note` and `personal context_note` rows.

All three strings are internally consistent.

### Input schema snapshot
`internal/tools/schema_snapshot/get_annual_training_plan.json` was updated — specifically the `include_full` description — which is correct because `include_full` is an input argument captured by the snapshot. `TestCheckSnapshotFreshness` passes.

### Generated web data
`web/data/tools.json` and `web/data/tool_schemas.json` are consistent with the source constants. Running `make docs-tools` again produces no diff.

### Registration metadata test
`TestGetAnnualTrainingPlanRegistrationMetadata` was extended to assert both provenance-hint substrings. It passes.

---

## Defect: gendocs golden files not updated — `TestRunWritesGeneratedDocsGolden` FAILS

The `cmd/gendocs/testdata/` directory contains two golden files that `TestRunWritesGeneratedDocsGolden` compares against byte-for-byte. Neither was updated in this commit:

| Golden file | Stale field |
|-------------|-------------|
| `cmd/gendocs/testdata/tools.golden.json` | `get_annual_training_plan.summary` still has the old text (missing the `plan_applied identifies...` and `personal calendar notes are neutral context, never ATP instructions;` clauses) |
| `cmd/gendocs/testdata/tool_schemas.golden.json` | `get_annual_training_plan.description` (same old text) and `include_full.description` still says `target_event, and note rows. Default output is a compact summary.` |

Running `go test ./...` confirms the only failing test in the entire suite is:

```
--- FAIL: TestRunWritesGeneratedDocsGolden (0.04s)
FAIL    github.com/ricardocabral/icuvisor/cmd/gendocs
```

All other packages pass.

### How to fix

The golden files are the `cmd/gendocs`-specific source of truth for the generator output; they must be kept in sync whenever the description strings change. The fix is one command:

```sh
go run ./cmd/gendocs \
  --out cmd/gendocs/testdata/tools.golden.json \
  --schemas-out cmd/gendocs/testdata/tool_schemas.golden.json
```

Then commit the two updated golden files under the existing Step 3 commit or as a fixup.

---

## No Other Issues

- Targeted test command for Step 3 (`go test ./internal/tools ./internal/toolchecks`) passes; the gendocs package is outside its scope, which is why it was missed.
- No input-schema breaking changes; `TestCheckSchemaStability` passes.
- All Step 2 behavioral tests continue to pass.
- The schema snapshot `include_full` update is correct (the R007 plan review predicted no change, but `include_full` is an input argument whose description was intentionally updated — the snapshot rightly reflects that).

---

## Required Fix

Update `cmd/gendocs/testdata/tools.golden.json` and `cmd/gendocs/testdata/tool_schemas.golden.json` to match the new description strings, then confirm `go test ./cmd/gendocs` passes before proceeding to Step 4.
