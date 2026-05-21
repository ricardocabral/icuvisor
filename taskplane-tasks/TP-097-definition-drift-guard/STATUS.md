# TP-097: Definition-drift guard for canonical formulas — Status

**Current Step:** Step 6: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-05-20
**Review Level:** 2
**Review Counter:** 12
**Iteration:** 1
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no protected docs are changed without explicit approval

---

### Step 1: Inventory formula-sensitive code
**Status:** ✅ Complete

- [x] Define a Step 1 inventory artifact in STATUS.md Notes with formula/ref ID, canonical text/hash target, implementation functions, tool adapters, existing coverage, planned golden cases, and known gaps.
- [x] Trace formula-sensitive behavior through both `internal/analysis` and `internal/tools`, including HR drift/Pw:HR, polarization, z-score, EF, and VI.
- [x] Record stable formula IDs/refs and expected outputs, including boundary cases for zero denominators, polarization zero/high states, sample standard deviation z-score, and EF/VI resource-only or upstream-derived status.
- [x] Decide golden fixture layout, including whether to reuse/extend `internal/resources/testdata/analysis_formulas.md` and where analyzer golden files live.
- [x] R003: Correct polarization planned golden input/expected output so it matches zone bucket mapping.
- [x] R003: Move stray review rows out of the inventory artifact and record R002/R003 in the Reviews table.

---

### Step 2: Add golden-file guards
**Status:** ✅ Complete

- [x] Add formula resource stability tests pinning canonical refs, formulas, and a committed content hash/golden.
- [x] Add shared analyzer golden fixtures under `testdata/analysis/` for drift, decoupling, polarization, z-score, EF resource-only status, and VI upstream-derived status.
- [x] Add `internal/analysis` tests that compare formula-sensitive computations and boundary cases against the golden fixture.
- [x] Add `internal/tools` tests that compare analyzer/tool `_meta.formula_ref` or upstream-mapped outputs against the golden fixture.
- [x] Ensure tests fail loudly with update instructions when canonical formula refs or analyzer outputs drift.

---

### Step 3: Document breaking-change policy
**Status:** ✅ Complete

- [x] Add a concise code/test note near canonical formula definitions or golden guards stating formula ref/text changes are breaking definition-drift events.
- [x] Review contributor/tool catalog docs and update only if they lack guidance for analyzer formula drift.
- [x] Decide whether CHANGELOG.md needs an entry; update it only if this task changes user-visible behavior.

---

### Step 4: Verify
**Status:** ✅ Complete

- [x] Run targeted analysis/resource/toolcheck tests.
- [x] Run full quality gate.

---


### Step 5: Testing & Verification
**Status:** ✅ Complete

- [x] Targeted tests passing
- [x] FULL test suite passing: `make test`
- [x] Build passes: `make build`
- [x] Lint passes: `make lint`
- [x] All failures fixed or documented as pre-existing unrelated failures

---

### Step 6: Documentation & Delivery
**Status:** ✅ Complete

- [x] "Must Update" docs modified
- [x] "Check If Affected" docs reviewed
- [x] Discoveries logged
- [x] Final commit includes task ID

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | Step 1 | REVISE | `.reviews/R001-plan-step1.md` |
| R002 | Plan | Step 1 | APPROVE | `.reviews/R002-plan-step1.md` |
| R003 | Code | Step 1 | REVISE | `.reviews/R003-code-step1.md` |
| R004 | Code | Step 1 | APPROVE | `.reviews/R004-code-step1.md` |
| R005 | Plan | Step 2 | APPROVE | `.reviews/R005-plan-step2.md` |
| R006 | Code | Step 2 | APPROVE | `.reviews/R006-code-step2.md` |
| R007 | Plan | Step 3 | APPROVE | `.reviews/R007-plan-step3.md` |
| R008 | Code | Step 3 | APPROVE | `.reviews/R008-code-step3.md` |
| R009 | Plan | Step 4 | APPROVE | `.reviews/R009-plan-step4.md` |
| R010 | Code | Step 4 | APPROVE | `.reviews/R010-code-step4.md` |
| R011 | Plan | Step 5 | APPROVE | `.reviews/R011-plan-step5.md` |
| R012 | Code | Step 5 | APPROVE | `.reviews/R012-code-step5.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| EF canonical formula currently has no local analyzer/tool output; it is pinned as resource-only until a product decision adds output semantics. | Captured in Step 1 inventory and `testdata/analysis/formula_golden.json`; guards require intentional update for future EF output. | `taskplane-tasks/TP-097-definition-drift-guard/STATUS.md`, `testdata/analysis/formula_golden.json` |
| VI is upstream-derived from `icu_variability_index`, not locally recomputed. | Guarded with get_extended_metrics golden tests, including missing-upstream omission. | `internal/tools/formula_golden_test.go` |
| `compute_baseline.go` had package-level helper names colliding with shared analyzer source helpers. | Renamed baseline-local helpers without behavior change so package tests compile. | `internal/tools/compute_baseline.go` |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-20 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-20 21:54 | Task started | Runtime V2 lane-runner execution |
| 2026-05-20 21:54 | Step 0 started | Preflight |
| 2026-05-20 22:33 | Worker iter 1 | done in 2325s, tools: 210 |
| 2026-05-20 22:33 | Task complete | .DONE created |

---

## Blockers

*None*

---

## Notes

Step 3 CHANGELOG decision: no CHANGELOG.md entry was added because this task added tests/contributor policy only and did not change runtime/user-visible behavior.

Step 6 documentation review: STATUS.md was kept current; CONTRIBUTING.md was updated with analyzer formula drift policy; CHANGELOG.md was reviewed and intentionally left unchanged. README.md, web/content/reference/tools.md, and docs/prd/PRD-icuvisor.md were reviewed for applicability; no update was needed because runtime tool behavior, generated catalog descriptions, and product scope did not change.

### Step 1 revised inventory plan

Inventory will be recorded in this Notes section as the Step 1 artifact to avoid broad docs churn. Columns/fields per formula: canonical formula/ref ID; canonical definition text or resource/golden hash target; implementation function(s); tool adapter(s) exposing result or `_meta.formula_ref`; existing test coverage; planned golden fixture/test case(s); known gaps or resource-only/upstream-derived status.

Step 1 code search targets: `AnalysisFormulaRef`, `formula_ref`, `ComputeZoneBalance`, `ComputeBaselineStats`, `ComputeTrend`, `ComputeSegmentStats`, and tool adapters in `internal/tools`.

Golden layout decision: keep the existing resource markdown golden in `internal/resources/testdata/analysis_formulas.md`; add cross-package analyzer golden JSON under repo-root `testdata/analysis/` so `internal/analysis` and `internal/tools` tests can load the same expected outputs via `../../testdata/analysis/...` without duplicating fixtures.

### Step 1 inventory artifact

| Formula | Ref/hash target | Canonical definition target | Implementation functions | Tool adapters / metadata | Existing coverage | Planned guard cases | Gaps/status |
|---|---|---|---|---|---|---|---|
| HR drift | `icuvisor://analysis-formulas#hr_drift`; formula markdown golden hash/content | `100 * (avg_hr_second_half - avg_hr_first_half) / avg_hr_first_half`, positive HR denominators, insufficient instead of divide-by-zero | `internal/analysis/segment_stats.go`: `ComputeActivitySegmentStats` -> `computeDrift` | `internal/tools/compute_activity_segment_stats.go` emits result and `_meta.formula_ref` from `SegmentStatsResult.FormulaRef` | `internal/resources/analysis_formulas_test.go`; `internal/analysis/segment_stats_test.go` verifies insufficient n/ref and requested-bound split value 100 | Analyzer golden: value 100, ref, method, details, audit halves; boundary: first/second HR <= 0 returns `insufficient_sample` with ref | Locally computed |
| Pw:HR decoupling | `icuvisor://analysis-formulas#pw_hr_decoupling`; formula markdown golden hash/content | `100 * (ratio_first - ratio_second) / ratio_first`, ratios from avg power / avg HR across elapsed-time halves, positive denominators | `internal/analysis/segment_stats.go`: `ComputeActivitySegmentStats` -> `computeDecoupling` | `compute_activity_segment_stats` emits `_meta.formula_ref` and optional full audit | `segment_stats_test.go` verifies value 10 and ref; tool test only checks non-empty formula ref | Analyzer golden: value 10, exact ref, ratio details; boundary: zero HR or zero first power returns `insufficient_sample` with ref | Locally computed |
| Polarization index | `icuvisor://analysis-formulas#polarization_index`; formula markdown golden hash/content | `log10((low_share / moderate_share) * (high_share / moderate_share) * 100)` with explicit undefined states for zero moderate/high | `internal/analysis/compute.go`: `ComputeZoneBalance`, `classifyZoneBalance` | `internal/tools/compute_zone_time.go`: `compute_zone_time` and `compute_load_balance` via `zoneAnalyzerMeta`, `_meta.formula_ref` | Resource markdown test; `compute_tools_test.go` covers zone-source selection but not exact formula ref/index | Analyzer golden: zones `[700,100,100,200]` -> low/mod/high shares `0.727273/0.090909/0.181818`, index `3.20412`, state `ok`, classification `polarized`, formula ref; boundaries: moderate zero and high zero states with nil index | Locally computed from upstream zone buckets/precomputed activity fields |
| Efficiency factor (EF) | `icuvisor://analysis-formulas#efficiency_factor`; formula markdown golden hash/content | `normalized_power / avg_hr` with positive avg HR and NP required | No local analyzer implementation found. `get_extended_metrics` maps upstream activity fields but no `efficiency_factor` field; `segment_stats.go` has `if` = intensity factor (`normalized_power / ftp_watts`), not EF | No current tool emits EF or EF formula ref | Resource markdown test pins ref/formula/citation/boundary | Guard resource text/ref; add analyzer golden documenting `efficiency_factor` as resource-only/no local output so future additions must deliberately add tool output + formula ref | Resource-only canonical definition; upstream/local analyzer gap recorded |
| Variability index (VI) | `icuvisor://analysis-formulas#variability_index`; formula markdown golden hash/content | `normalized_power / avg_power`, positive avg power and NP required | No local formula implementation; upstream activity field `icu_variability_index` mapped in `internal/tools/get_extended_metrics.go`; metric alias `vi` in `internal/analysis/metrics.go` sources `get_extended_metrics.variability_index` | `get_extended_metrics` emits `metrics.variability_index`; no `_meta.formula_ref` because value is upstream-mapped | Resource markdown test; `get_extended_metrics_test.go` verifies omitted upstream VI is not zero-filled | Guard resource text/ref and upstream-mapped fixture value `1.07`; boundary: missing upstream `icu_variability_index` omitted, not recomputed/substituted | Upstream-derived value, not locally recomputed |
| z-score | `icuvisor://analysis-formulas#z_score`; formula markdown golden hash/content | `(current_value - baseline_mean) / sample_standard_deviation`; sample std dev denominator `n-1`; min baseline samples; zero variance omitted/insufficient | `internal/analysis/compute.go`: `ComputeBaselineStats`; `internal/analysis/trend.go`: `ComputeTrend` baseline delta path using `Stats.StdDev` | `compute_baseline` always emits z-score formula ref; `analyze_trend` emits formula ref only when `ZScore != nil` | Resource markdown test; `compute_tools_test.go` exact baseline mean/stddev/z-score/ref; `analyzer_math_test.go` checks trend z-score presence | Analyzer golden: `ComputeBaselineStats([50,60],[70],2,false)` -> mean 55, sample stddev 7.071068, z-score 2.12132; tool output rounded 2.1213/ref; trend z-score ref when available; boundary: zero baseline variance suppresses z-score/ref path as designed | Locally computed in baseline and trend paths |
| 2026-05-20 22:05 | Review R004 | code Step 1: APPROVE |
| 2026-05-20 22:08 | Review R005 | plan Step 2: APPROVE |
| 2026-05-20 22:17 | Review R006 | code Step 2: APPROVE |
| 2026-05-20 22:19 | Review R007 | plan Step 3: APPROVE |
| 2026-05-20 22:21 | Review R008 | code Step 3: APPROVE |
| 2026-05-20 22:23 | Review R009 | plan Step 4: APPROVE |
| 2026-05-20 22:25 | Review R010 | code Step 4: APPROVE |
| 2026-05-20 22:27 | Review R011 | plan Step 5: APPROVE |
| 2026-05-20 22:30 | Review R012 | code Step 5: APPROVE |
