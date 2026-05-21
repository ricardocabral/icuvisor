# TP-098: Analyzer toolset placement and core-promotion gate — Status

**Current Step:** Step 6: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-05-20
**Review Level:** 1
**Review Counter:** 8
**Iteration:** 1
**Size:** S

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no protected docs are changed without explicit approval

---

### Step 1: Audit analyzer tier placement
**Status:** ✅ Complete

- [x] List all analyzer-family tools and current registered tiers.
- [x] Record analyzer-family source of truth (`analyzerFamilyCatalogNames`, analyzer catalog group cases, and `get_fitness_projection` activation-test inclusion).
- [x] Audit constructor wrapping and effective registered tiers for each analyzer-family tool in STATUS.md.
- [x] Confirm all are `full` before benchmark evidence.
- [x] Record TP-100/KR5 evidence and promotion eligibility for only `analyze_trend`, `compute_zone_time`, and `compute_baseline`.
- [x] Define core-promotion acceptance note tied to KR5 benchmark results.

---

### Step 2: Enforce placement in tests
**Status:** ✅ Complete

- [x] Add/adjust catalog-tier tests so analyzer family defaults to `full`.
- [x] Add a dedicated analyzer-family default-tier test in `internal/tools/catalog_tiers_test.go` using `analyzerFamilyCatalogNames()`.
- [x] Add a promotion-candidate policy test for only `analyze_trend`, `compute_zone_time`, and `compute_baseline`, tied to `docs/kr5-benchmark.md` evidence.
- [x] Ensure `icuvisor_list_advanced_capabilities` advertises hidden analyzer capabilities clearly.
- [x] Strengthen `internal/tools/list_advanced_capabilities_test.go` so every effective full-only analyzer-family tool is advertised with clear summary text.
- [x] Run targeted Step 2 tool tests and record the command/result.

---

### Step 3: Apply promotion if evidence exists
**Status:** ✅ Complete

- [x] If TP-100 benchmark results are present and positive, promote only `analyze_trend`, `compute_zone_time`, and `compute_baseline` to `core`.
- [x] Change only the three eligible constructors (`analyze_trend`, `compute_zone_time`, `compute_baseline`) from `fullTool(...)` to `coreTool(...)`.
- [x] Update tier-policy tests so the three benchmark-gated candidates are `core` and all non-candidate analyzer-family tools remain `full`.
- [x] If evidence is absent/negative, leave all analyzers in `full` and document why.
- [x] Confirm promoted core analyzers are no longer advertised as hidden full-only capabilities while remaining full-only analyzers still are.
- [x] Regenerate catalog docs data with `make docs-tools` and review generated/manual tool docs.
- [x] Update docs accordingly.

---

### Step 4: Verify
**Status:** ✅ Complete

- [x] Run catalog/toolset tests and full quality gate.
- [x] Update CHANGELOG.md.

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
| R001 | Plan | Step 1 | REVISE | .reviews/R001-plan-step1.md |
| R002 | Plan | Step 1 | APPROVE | .reviews/R002-plan-step1.md |
| R003 | Plan | Step 2 | REVISE | .reviews/R003-plan-step2.md |
| R004 | Plan | Step 2 | APPROVE | .reviews/R004-plan-step2.md |
| R005 | Plan | Step 3 | REVISE | .reviews/R005-plan-step3.md |
| R006 | Plan | Step 3 | APPROVE | .reviews/R006-plan-step3.md |
| R007 | Plan | Step 4 | APPROVE | .reviews/R007-plan-step4.md |
| R008 | Plan | Step 5 | APPROVE | .reviews/R008-plan-step5.md |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| No out-of-scope discoveries; TP-100 evidence was already present and positive for the three intended promotion candidates. | Recorded in Step 1/3 notes; no follow-up needed. | `docs/kr5-benchmark.md`, `STATUS.md` |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-20 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-20 23:47 | Task started | Runtime V2 lane-runner execution |
| 2026-05-20 23:47 | Step 0 started | Preflight |
| 2026-05-21 00:14 | Worker iter 1 | done in 1622s, tools: 159 |
| 2026-05-21 00:14 | Task complete | .DONE created |

---

## Blockers

*None*

---

## Notes

### Step 1 analyzer-family audit

Source of truth: `internal/tools/catalog_test.go` `analyzerFamilyCatalogNames()` plus `internal/tools/catalog.go` analyzer `toolCatalogGroup` cases. `get_activity_histogram` and `get_fitness_projection` are included because analyzer activation tests treat them as analyzer-family tools even though their catalog groups are `activities` and `fitness`, respectively. Verification command: `go test ./internal/tools -run 'TestCatalogIncludesFullAnalyzers|TestCatalogAnalyzerActivationHints|TestRegisteredToolTierMembership'` (passed).

| Tool | Family source | Catalog group | Constructor/effective tier | Promotion eligibility/evidence |
| ---- | ------------- | ------------- | -------------------------- | ------------------------------ |
| `analyze_trend` | analyzerFamilyCatalogNames + analyzer group | `analyzers` | `fullTool(...)` / `full` | Eligible: TP-100 KR5 net +149 tokens, raw pulls `1 -> 0`, gate meets |
| `analyze_distribution` | analyzerFamilyCatalogNames + analyzer group | `analyzers` | `fullTool(...)` / `full` | Not eligible: no TP-098 per-candidate promotion evidence |
| `analyze_correlation` | analyzerFamilyCatalogNames + analyzer group | `analyzers` | `fullTool(...)` / `full` | Not eligible: no TP-098 per-candidate promotion evidence |
| `analyze_efforts_delta` | analyzerFamilyCatalogNames + analyzer group | `analyzers` | `fullTool(...)` / `full` | Not eligible: no TP-098 per-candidate promotion evidence |
| `compute_zone_time` | analyzerFamilyCatalogNames + analyzer group | `analyzers` | `fullTool(...)` / `full` | Eligible: TP-100 KR5 net +30 tokens, raw pulls `1 -> 0`, gate meets |
| `compute_load_balance` | analyzerFamilyCatalogNames + analyzer group | `analyzers` | `fullTool(...)` / `full` | Not eligible: no TP-098 per-candidate promotion evidence |
| `compute_baseline` | analyzerFamilyCatalogNames + analyzer group | `analyzers` | `fullTool(...)` / `full` | Eligible: TP-100 KR5 net +219 tokens, raw pulls `0 -> 0`, gate meets |
| `compute_compliance_rate` | analyzerFamilyCatalogNames + analyzer group | `analyzers` | `fullTool(...)` / `full` | Not eligible: no TP-098 per-candidate promotion evidence |
| `compute_activity_segment_stats` | analyzerFamilyCatalogNames + analyzer group | `analyzers` | `fullTool(...)` / `full` | Not eligible: no TP-098 per-candidate promotion evidence |
| `get_activity_histogram` | analyzerFamilyCatalogNames activation-test inclusion | `activities` | `fullTool(...)` / `full` | Not eligible: no TP-098 per-candidate promotion evidence |
| `get_fitness_projection` | analyzerFamilyCatalogNames activation-test inclusion | `fitness` | `fullTool(...)` / `full` | Not eligible: no TP-098 per-candidate promotion evidence |

Core-promotion acceptance note: before Step 3, all analyzer-family tools are `full`. TP-100's `docs/kr5-benchmark.md` evidence makes only `analyze_trend`, `compute_zone_time`, and `compute_baseline` eligible for core promotion. All other analyzer-family tools must stay `full` unless future benchmark evidence is added and reviewed.

### Step 2 targeted verification

`go test ./internal/tools -run 'TestRegisteredToolTierMembership|TestAnalyzerFamilyDefaultsToFullToolset|TestAnalyzerCorePromotionCandidatesAreBenchmarkGated|TestCatalogIncludesFullAnalyzers|TestCatalogAnalyzerActivationHints|TestListAdvancedCapabilitiesOutputFromCatalog'` passed after adding analyzer-specific tier, promotion-candidate, and advanced-capabilities coverage.

### Step 3 promotion notes

TP-100 evidence was present and positive, so the absent/negative-evidence branch was not taken. Promoted only `analyze_trend`, `compute_zone_time`, and `compute_baseline` to `core`; all non-candidate analyzer-family tools remain `full`. Targeted verification command after promotion: `go test ./internal/tools -run 'TestRegisteredToolTierMembership|TestNonCandidateAnalyzerFamilyRemainsFullToolset|TestAnalyzerCorePromotionCandidatesAreBenchmarkGated|TestCatalogIncludesAnalyzerFamilyPlacement|TestCatalogAnalyzerActivationHints|TestListAdvancedCapabilitiesOutputFromCatalog'` (passed). Ran `make docs-tools`, which updated `web/data/tools.json`; reviewed `web/content/reference/tools.md` as generated-data backed with no manual edit needed. Updated `CHANGELOG.md` under `[Unreleased]` for the core toolset behavior change.

### Step 4 verification notes

Catalog/toolset targeted command passed after refreshing the gendocs golden: `go test ./cmd/gendocs ./internal/tools -run 'TestRunWritesToolsCatalogGolden|TestRegisteredToolTierMembership|TestNonCandidateAnalyzerFamilyRemainsFullToolset|TestAnalyzerCorePromotionCandidatesAreBenchmarkGated|TestCatalogIncludesAnalyzerFamilyPlacement|TestCatalogAnalyzerActivationHints|TestListAdvancedCapabilitiesOutputFromCatalog'`. Full quality gate command `make test` passed after updating `cmd/gendocs/testdata/tools.golden.json`. `CHANGELOG.md` already contains the `[Unreleased]` core-promotion entry.

### Step 5 quality gate notes

Targeted tests passed: `go test ./cmd/gendocs ./internal/tools ./internal/safety ./internal/toolcatalog -run 'TestRunWritesToolsCatalogGolden|TestRegisteredToolTierMembership|TestNonCandidateAnalyzerFamilyRemainsFullToolset|TestAnalyzerCorePromotionCandidatesAreBenchmarkGated|TestCatalogIncludesAnalyzerFamilyPlacement|TestCatalogAnalyzerActivationHints|TestListAdvancedCapabilitiesOutputFromCatalog|TestToolEffectiveToolsetDefaultsEmptyToFull|TestCatalog'`. Full suite `make test`, build `make build`, and lint `make lint` all passed. No remaining failures.

### Step 6 delivery notes

Must-update docs completed: `CHANGELOG.md` records the user-visible core promotion and `STATUS.md` records execution evidence. Check-if-affected docs reviewed: `web/content/reference/tools.md` is generated-data backed and needed no manual edit after `make docs-tools`; `README.md` and `docs/prd/PRD-icuvisor.md` do not need changes because the public tool catalog remains generated and behavior stays within analyzer roadmap scope.

| 2026-05-20 23:49 | Review R001 | plan Step 1: REVISE |
| 2026-05-20 23:51 | Review R002 | plan Step 1: APPROVE |
| 2026-05-20 23:55 | Review R003 | plan Step 2: REVISE |
| 2026-05-20 23:56 | Review R004 | plan Step 2: APPROVE |
| 2026-05-21 00:01 | Review R005 | plan Step 3: REVISE |
| 2026-05-21 00:03 | Review R006 | plan Step 3: APPROVE |
| 2026-05-21 00:08 | Review R007 | plan Step 4: APPROVE |
| 2026-05-21 00:11 | Review R008 | plan Step 5: APPROVE |
