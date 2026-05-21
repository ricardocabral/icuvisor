# Review R001 — Plan Review for Step 1

**Verdict:** Needs revision before marking Step 1 complete.

I only found the high-level Step 1 checklist in `STATUS.md`/`PROMPT.md`; there is no concrete inventory plan or artifact yet. The direction is correct, but for this task the inventory is the foundation for the golden guards, so it needs to be explicit enough that Step 2 cannot miss a formula path.

## Required changes to the Step 1 plan

1. **Define the Step 1 inventory artifact.**  
   Record where the inventory will live before implementation proceeds. Prefer `STATUS.md` Notes/Discoveries or a small task-local note if the runner allows it; avoid broad docs churn. The inventory should include at least these columns:
   - canonical formula/ref ID
   - canonical definition text or current resource/golden hash target
   - implementation function(s)
   - tool adapter(s) that expose the result or `_meta.formula_ref`
   - existing test coverage
   - planned golden fixture/test case(s)
   - known gaps or resource-only/upstream-derived status

2. **Inventory ref propagation, not just math functions.**  
   Step 1 should explicitly trace formula-sensitive behavior through both `internal/analysis` and `internal/tools`:
   - HR drift / Pw:HR decoupling: `internal/analysis/segment_stats.go` and `compute_activity_segment_stats` response metadata.
   - Polarization index: `ComputeZoneBalance` and `compute_zone_time` metadata.
   - z-score: `ComputeBaselineStats`, `ComputeTrend` baseline z-score path, `compute_baseline`, and `analyze_trend` metadata.
   - EF / VI: document that the canonical refs exist in `analysis_formulas.go`; verify whether these are locally computed, upstream-mapped via `get_extended_metrics`/`analysis/metrics.go`, or currently resource-only. Do not silently skip them because there may be no local analyzer computation.

3. **Make the golden fixture layout decision concrete.**  
   The prompt scopes `testdata/analysis/**/*`, while `internal/resources/testdata/analysis_formulas.md` already exists. The plan should decide whether analyzer golden files live in repo-root `testdata/analysis/...` or package-local `testdata`, and how tests will load them without brittle relative paths. Also note whether the existing formulas markdown golden is reused or extended.

4. **Define expected outputs and boundary cases during inventory.**  
   The plan should require deterministic cases for the formulas named in the mission, including rounded numeric outputs and `_meta.formula_ref` values. Include boundary cases where the definition is easy to drift: zero/insufficient denominators for drift/decoupling, moderate/high-zero polarization states, sample standard deviation (`n-1`) for z-score, and the explicit handling/status for EF/VI if they are not locally computed.

## Non-blocking suggestions

- Use code search over `AnalysisFormulaRef`, `formula_ref`, `ComputeZoneBalance`, `ComputeBaselineStats`, and segment stat constants as part of the inventory and paste the resulting map into the artifact.
- Avoid updating `CHANGELOG.md` in Step 1 unless the inventory uncovers an actual user-visible behavior change; the prompt only requires it if behavior changes.

Once those details are added, the Step 1 plan should be sufficient to proceed to golden-file guard implementation.
