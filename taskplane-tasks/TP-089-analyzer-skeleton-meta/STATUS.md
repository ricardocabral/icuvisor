# TP-089: Analyzer skeleton and mandatory `_meta` contract — Status

**Current Step:** Step 6: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-05-20
**Review Level:** 2
**Review Counter:** 12
**Iteration:** 3
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no protected docs are changed without explicit approval

---

### Step 1: Design shared meta structs
**Status:** ✅ Complete

- [x] Define typed meta structs/helpers for method/source_tools/n/missing/insufficient-sample fields.
- [x] Decide default missing action (`skip`) and minimum-sample helper shape.
- [x] Map formula refs to the resource added in TP-088.
- [x] Record exact non-omitempty analyzer meta JSON contract and non-nil `source_tools` rule.
- [x] Record package ownership and `response.Shape` interaction.
- [x] Record deterministic `source_tools`, `n`, `missing_days`, and sample-threshold semantics.
- [x] Record formula-ref ownership using TP-088 exported constants without duplicated URI fragments.
- [x] Fix R003 audit table file paths so R002/R003 point at concrete review files.
- [x] Move review execution-log rows from Notes into the Execution Log table.

---

### Step 2: Implement skeleton helpers
**Status:** ✅ Complete

- [x] Implement `internal/analysis` API: `AnalyzerMetaInput`, `NewAnalyzerMeta`, `NormalizeSourceTools`, `InsufficientSample`, `MissingActionSkip`, `MinBaselineSamples`, and `MinCorrelationSamples`.
- [x] Enforce documented normalization defaults: trim method, clamp negative counts to zero, normalize source tools, default missing action to `skip`, and pass caller-provided formula refs through unchanged.
- [x] Add a small `internal/tools` analyzer response envelope/helper that omits heavy series data unless `include_full` is true and preserves analyzer `_meta` through `response.Shape`.
- [x] Add post-shape golden fixtures/test utilities for analyzer responses.
- [x] Keep helpers small and internal.

---

### Step 3: Add a no-op/demo test path
**Status:** ✅ Complete

- [x] Create tests that prove a sample analyzer response contains all mandatory meta fields.
- [x] Assert terse/full behavior and missing-day handling at helper level.
- [x] Run targeted tests.

---

### Step 4: Verify
**Status:** ✅ Complete

- [x] Run full quality gate.
- [x] Update CHANGELOG.md if public scaffolding is visible.
- [x] Record conventions in STATUS.md for downstream tasks.

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
| R001 | plan | 1 | REVISE | .reviews/R001-plan-step1.md |
| R002 | plan | 1 | APPROVE | .reviews/R002-plan-step1.md |
| R003 | code | 1 | REVISE | .reviews/R003-code-step1.md |
| R004 | plan | 2 | REVISE | .reviews/R004-plan-step2.md |
| R005 | plan | 2 | APPROVE | .reviews/R005-plan-step2.md |
| R006 | code | 2 | APPROVE | .reviews/R006-code-step2.md |
| R007 | plan | 3 | APPROVE | .reviews/R007-plan-step3.md |
| R008 | code | 3 | APPROVE | .reviews/R008-code-step3.md |
| R009 | plan | 4 | APPROVE | .reviews/R009-plan-step4.md |
| R010 | code | 4 | APPROVE | .reviews/R010-code-step4.md |
| R011 | plan | 5 | APPROVE | .reviews/R011-plan-step5.md |
| R012 | code | 5 | APPROVE | .reviews/R012-code-step5.md |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| No out-of-scope discoveries. | None required | TP-089 Step 6 |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-20 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-20 15:23 | Task started | Runtime V2 lane-runner execution |
| 2026-05-20 15:23 | Step 0 started | Preflight |
| 2026-05-20 15:26 | Review R001 | plan Step 1: REVISE |
| 2026-05-20 15:31 | Review R002 | plan Step 1: APPROVE |
| 2026-05-20 15:32 | Review R003 | code Step 1: REVISE |
| 2026-05-20 15:42 | Worker iter 1 | done in 1123s, tools: 66 |
| 2026-05-20 15:42 | Step 2 started | Implement skeleton helpers |
| 2026-05-20 15:43 | Review R004 | plan Step 2: REVISE |
| 2026-05-20 15:44 | Review R005 | plan Step 2: APPROVE |
| 2026-05-20 16:14 | Worker iter 2 | done in 1919s, tools: 38 |
| 2026-05-20 16:18 | Review R006 | code Step 2: APPROVE |
| 2026-05-20 16:20 | Review R007 | plan Step 3: APPROVE |
| 2026-05-20 16:25 | Review R008 | code Step 3: APPROVE |
| 2026-05-20 16:27 | Review R009 | plan Step 4: APPROVE |
| 2026-05-20 16:31 | Review R010 | code Step 4: APPROVE |
| 2026-05-20 16:34 | Review R011 | plan Step 5: APPROVE |
| 2026-05-20 16:39 | Review R012 | code Step 5: APPROVE |
| 2026-05-20 16:42 | Worker iter 3 | done in 1700s, tools: 129 |
| 2026-05-20 16:42 | Task complete | .DONE created |

---

## Blockers

*None*

---

## Notes

### Step 1 design decisions

Analyzer `_meta` contract lives in `internal/analysis/meta.go` as `AnalyzerMeta` with typed fields and JSON tags:

```go
type AnalyzerMeta struct {
    Method             string   `json:"method"`
    SourceTools        []string `json:"source_tools"`
    N                  int      `json:"n"`
    MissingDays        int      `json:"missing_days"`
    MissingAction      string   `json:"missing_action"`
    InsufficientSample bool     `json:"insufficient_sample"`
    FormulaRef         string   `json:"formula_ref,omitempty"`
}
```

`method`, `source_tools`, `n`, `missing_days`, `missing_action`, and `insufficient_sample` are mandatory and must never use `omitempty`; they are emitted even when empty, zero, or false. `formula_ref` is optional and uses `omitempty` because not every analyzer maps to a canonical TP-088 formula. Builders must initialize `source_tools` to a non-nil slice so response JSON does not produce `null` or allow terse shaping to strip the mandatory key.

Package ownership: analyzer-domain types/helpers stay small in `internal/analysis`; `internal/tools` may wrap them only to assemble tool-specific response envelopes. Analyzer helpers do not add response-owned keys (`server_version`, `catalog_hash`, `delete_mode`, `toolset`, `units`). `response.Shape` remains the response boundary that preserves analyzer `_meta` keys while adding common response-owned metadata.

Sampling semantics: `n` is the count of usable samples after skipped missing values; pairwise analyzers report usable pairs. `missing_days` counts athlete-local daily buckets skipped because required inputs were missing; non-daily analyzers still emit `0`. `missing_action` defaults to constant `MissingActionSkip = "skip"`; forward-fill is never the default. Step 2 should add `InsufficientSample(n, minN int) bool` plus small threshold constants/guidance for PRD minima (`MinBaselineSamples = 7`, `MinCorrelationSamples = 14`), while allowing stricter tool-specific rules.

`source_tools` normalization should be deterministic: trim blanks, dedupe, and sort tool names before assigning them to `_meta.source_tools`; preserve an empty but non-nil slice when no source tools are supplied in tests/demo paths.

Formula refs: analyzers must import exported TP-088 constants from `internal/resources` (`AnalysisFormulaRefHRDrift`, `AnalysisFormulaRefPwHRDecoupling`, `AnalysisFormulaRefPolarization`, `AnalysisFormulaRefEfficiencyFactor`, `AnalysisFormulaRefVariabilityIndex`, `AnalysisFormulaRefZScore`) instead of hand-copying URI fragments. Step 2 may expose an analysis helper such as `AnalyzerMetaInput.FormulaRef` but should not duplicate the string values.

### Step 2 implementation plan

`internal/analysis` owns analyzer-domain semantics and exposes only small helpers: `AnalyzerMetaInput`, `NewAnalyzerMeta(input) AnalyzerMeta`, `NormalizeSourceTools`, `InsufficientSample`, plus `MissingActionSkip`, `MinBaselineSamples`, and `MinCorrelationSamples`. `internal/tools` may own a small analyzer response envelope/helper for MCP-facing responses and include-full behavior, but it must not duplicate analyzer-domain meta normalization.

Normalization policy: `NewAnalyzerMeta` trims whitespace from `method` and `missing_action`; negative `n` and `missing_days` are clamped to zero; blank/duplicate source tools are removed and the remaining names are sorted; a blank `missing_action` defaults to `MissingActionSkip`; `formula_ref` is trimmed and passed through unchanged so callers import TP-088 constants rather than duplicating fragments. `source_tools` must always be non-nil, including the omitted-source path.

Terse/full policy: the tool-level skeleton helper owns optional heavy fields such as per-bucket/per-sample `series`; it must set them only when `include_full` is true before calling `response.Shape`. Tests should pin post-shape JSON using deterministic `response.Options` so analyzer mandatory `_meta` keys survive alongside response-owned meta fields.

Golden strategy: helper/unit tests live near the code (`internal/analysis/meta_test.go` and `internal/tools/analyzer_common_test.go`) with post-shape fixtures in `internal/tools/testdata/*.golden.json`. Edge cases to pin include `source_tools: []` instead of `null`, zero-valued mandatory keys present in terse JSON, `missing_action: "skip"` defaulting, deterministic source-tool normalization, and formula refs passed from `internal/resources` constants.

### Downstream analyzer conventions

Future analyzer tools should build `_meta` with `analysis.NewAnalyzerMeta` and use `shapeAnalyzerResponse`/`encodeAnalyzerResponse` at the MCP boundary when returning analyzer-style results. Put heavy per-sample or per-bucket data in `Series`; it is omitted from terse responses and included only when `include_full` is true. Analyzer tests should compare post-shape JSON with `assertAnalyzerGolden` so mandatory analyzer meta and response-owned meta are pinned together.

### Documentation review

CHANGELOG.md was updated because the scaffolding is visible in the unreleased codebase. README.md, web/content/reference/tools.md, and docs/prd/PRD-icuvisor.md were reviewed; no updates were needed because no registered MCP tool, generated tool reference, setup flow, or product analyzer contract changed beyond implementing the planned internal scaffold.
