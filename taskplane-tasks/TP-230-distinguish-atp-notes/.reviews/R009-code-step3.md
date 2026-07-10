# R009 — Code Review: Step 3 (Update schema and generated surfaces) — Iteration 2

**Verdict: APPROVE**

---

## Summary

This iteration addresses the sole R008 defect: the two `cmd/gendocs/testdata/` golden files were not updated in the original Step 3 commit. The fix commit (`05b9f6e fix(TP-230): refresh gendocs golden fixtures`) resolves it. All targeted tests now pass, and all generated surfaces are consistent with the source strings.

---

## R008 Defect: Resolved

`TestRunWritesGeneratedDocsGolden` was failing because `cmd/gendocs/testdata/tools.golden.json` and `cmd/gendocs/testdata/tool_schemas.golden.json` still carried the old description strings. Both files are now updated in commit `05b9f6e` and match the source constants:

| File | Updated field | Correct |
|------|---------------|---------|
| `cmd/gendocs/testdata/tools.golden.json` | `get_annual_training_plan.summary` | ✓ |
| `cmd/gendocs/testdata/tool_schemas.golden.json` | `get_annual_training_plan.description` and `include_full.description` | ✓ |

`go test ./cmd/gendocs` passes. `TestRunWritesGeneratedDocsGolden` passes.

---

## Consistency Verified Across All Surfaces

All four generated/committed files carry identical description strings:

| Surface | Consistent |
|---------|------------|
| `internal/tools/get_annual_training_plan.go` source constants | source of truth |
| `internal/tools/schema_snapshot/get_annual_training_plan.json` | ✓ |
| `web/data/tools.json` | ✓ |
| `web/data/tool_schemas.json` | ✓ |
| `cmd/gendocs/testdata/tools.golden.json` | ✓ |
| `cmd/gendocs/testdata/tool_schemas.golden.json` | ✓ |

`diff web/data/tools.json cmd/gendocs/testdata/tools.golden.json` is empty; same for the `tool_schemas` pair.

---

## Tests

All three targeted packages pass:

```
ok  github.com/ricardocabral/icuvisor/cmd/gendocs
ok  github.com/ricardocabral/icuvisor/internal/tools
ok  github.com/ricardocabral/icuvisor/internal/toolchecks
```

Full suite (`go test ./...`) passes with no failures across all packages.

Specific tests confirmed:
- `TestRunWritesGeneratedDocsGolden` — passes (was failing before fix)
- `TestCheckSnapshotFreshness` — passes
- `TestGetAnnualTrainingPlanRegistrationMetadata` — passes (asserts both provenance-hint substrings)

---

## Description Changes Are Correct

All three description strings in `get_annual_training_plan.go` are internally consistent and meet the Step 3 requirements:

- **`getAnnualTrainingPlanDescription`**: adds `plan_applied identifies ATP-generated notes, while personal calendar notes are neutral context, never ATP instructions;` and replaces "recovery/context notes" with "provenance-aware notes".
- **`getAnnualTrainingPlanOutputSchema()` description**: explicitly states ATP provenance rule and that personal notes are neutral context never counted as recovery conclusions.
- **`getAnnualTrainingPlanInputSchema()` `include_full` description**: names `ATP note` and `personal context_note` rows.

---

## Minor Observation (Non-blocking)

The R007 plan review recommended removing "recovery weeks" from the trigger clause. The phrase remains in the updated description:

> `...weekly load/TSS targets, recovery weeks, taper context, or periodization summary...`

R008 did not require this change, and it was not listed as a defect. Keeping "recovery weeks" as a trigger is defensible: the tool does return periodization structure that may include recovery phases. The behavioral correction (removal of keyword-based recovery classification) was completed in Step 2. This is noted for awareness only.

---

## Commit Hygiene

Three commits since baseline, all conforming to the TP-230 convention:
- `4695dd3 fix(TP-230): complete Step 3 — update generated tool surfaces`
- `53fe378 hydrate: TP-230 add R008 revision item to Step 3`
- `05b9f6e fix(TP-230): refresh gendocs golden fixtures`

All messages use lowercase subjects, imperative mood, and include TP-230.

---

## Ready for Step 4

Step 4 (`make test`, `make test-race`, `make lint`, `make build`, `make docs-tools && git diff --check`) can proceed.
