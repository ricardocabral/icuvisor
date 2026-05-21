# Code Review R006 — Step 2: Implement resource

**Verdict:** REVISE

## Findings

### 1. Polarization-index boundary handling was not fully carried into the resource

`internal/resources/analysis_formulas.go:33` publishes the canonical `polarization_index` entry, but it drops the Step 1 boundary decision for `high_share == 0`. The Step 1 source table in `STATUS.md:129` says that if high share is zero, PI is undefined for polarized classification and bucket shares should drive a non-polarized label. The rendered resource only calls out the `moderate_share == 0` divide-by-zero case, and the invariant test pins only that wording (`internal/resources/analysis_formulas_test.go:45-49`).

This resource is intended to be the locked formula contract that future analyzers cite via `_meta.formula_ref`; omitting this edge case leaves implementers free to compute `log10(0)`/`-Inf` or to invent a different behavior for no-high-intensity blocks.

**Fix:** Add the high-share-zero boundary behavior from `STATUS.md` to the polarization paragraph, update `testdata/analysis_formulas.md`, and pin the wording in `TestAnalysisFormulasMarkdownPinsRequiredFormulaRefs` (for example with a boundary check for `high share is zero` or equivalent).

## Tests run

- `go test ./internal/resources ./internal/mcp`
- `git diff --check`
