# Plan Review — TP-096 Step 4

**Verdict:** Changes requested

The Step 4 plan is directionally right, but it is too vague for the current tree. Step 3 did more than regenerate catalog artifacts: it changed `internal/tools/compute_baseline.go` to clear the generator build blocker. Verification therefore needs to cover both the new catalog/toolcheck guardrails and the production compute-baseline behavior touched during Step 3.

I sampled the current tree with a targeted command that includes the touched production path:

```sh
go test ./internal/tools -run 'TestCatalog|TestComputeBaseline|TestComputeActivitySegmentStats|TestGetActivityHistogram|TestGetFitnessProjection' -count=1
```

It currently fails in `TestComputeBaselineWellnessZScoreIncludesFormulaAndInterpretation` and `TestComputeBaselineStatusGoldensAndValidation`, with wellness baseline rows coming back as `n_baseline:0` / `not_enough_baseline_samples` instead of the expected baseline/status results. That means the Step 4 plan should not limit the targeted phase to only toolcheck/catalog tests; it must explicitly include the compute-baseline regression tests introduced by the Step 3 fix.

## Required plan changes

1. **Make targeted commands explicit.** Include at least:

   ```sh
   go test ./internal/tools -run 'TestCatalog' -count=1
   go test ./internal/toolchecks -run 'TestCheckConfusableCatalog|TestFirstDescriptionSentence|TestGenerateToolCatalogUsesCallerContext' -count=1
   go test ./cmd/gendocs -run TestGenerateToolsGolden -count=1
   ```

2. **Add targeted coverage for production code touched in Step 3.** Because `compute_baseline.go` was modified to resolve the build blocker, run and require the compute baseline tests before broader gates:

   ```sh
   go test ./internal/tools -run 'TestComputeBaseline' -count=1
   ```

   Keep `TestComputeActivitySegmentStats`, `TestGetActivityHistogram`, and `TestGetFitnessProjection` in the targeted set if their descriptions or catalog grouping changed and you want a quick sanity check of those analyzer-family tools.

3. **Define “full quality gate” concretely.** The plan should name the command(s), not just the phrase. For this repository that can be:

   ```sh
   make check
   ```

   Note that `make check` covers fmt-check, vet, lint, and race tests, but not `make build`; Step 5 still needs to run the explicit completion-criteria commands (`make test`, `make build`, `make lint`) unless the task owner decides Step 4 should absorb Step 5.

4. **State the failure policy.** Step 4 should say that any failure is fixed before proceeding, unless it is clearly pre-existing and unrelated, in which case the exact command/output and rationale are recorded in `STATUS.md`. The current compute-baseline failures are in touched code and should be treated as in-scope, not documented away.

With these additions, Step 4 will verify the catalog-quality changes and guard against regressions from the build-blocker repair performed during Step 3.
