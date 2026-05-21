# Plan Review R002 — Step 1: Design shared meta structs

Verdict: **Approved**

The revised Step 1 notes in `STATUS.md` now provide a concrete contract and resolve the issues from R001. I checked the task prompt, current status, prior review, PRD analyzer rules, `internal/response/{shape,meta}.go`, and the TP-088 formula constants in `internal/resources/analysis_formulas.go`.

## What is now sufficient

- The planned `analysis.AnalyzerMeta` shape names every mandatory `_meta` field and leaves those fields non-`omitempty`, so `method`, `source_tools`, `n`, `missing_days`, `missing_action`, and `insufficient_sample` will be emitted even for zero/false values.
- `formula_ref` is explicitly optional and tied to the `icuvisor://analysis-formulas` registry instead of being treated as a mandatory field for every analyzer.
- Package ownership is clear: analyzer-domain structs/helpers live in `internal/analysis`; `internal/tools` should only assemble tool-specific envelopes.
- The plan correctly accounts for `response.Shape`: analyzer-owned `_meta` keys are preserved, while response-owned metadata remains added at the response boundary.
- Sampling semantics are defined for downstream tools: `n` means usable samples/pairs after skipped missing values, `missing_days` means skipped athlete-local daily buckets, and `missing_action` defaults to `skip` with no forward-fill default.
- The sample-threshold direction (`InsufficientSample(n, minN int)`, `MinBaselineSamples = 7`, `MinCorrelationSamples = 14`) matches PRD §7.2.C.
- Formula refs will import TP-088 exported constants from `internal/resources`, avoiding duplicated URI fragments.
- `source_tools` determinism is specified: trim blanks, dedupe, sort, and preserve an empty non-nil slice.

## Non-blocking implementation notes for Step 2

- Prefer making callers go through a builder/input helper rather than constructing `AnalyzerMeta{}` directly; that is the easiest way to guarantee non-nil `source_tools` and default `missing_action: "skip"`.
- Add tests that marshal and shape the response with `include_full: false` to prove `source_tools: []`, `n: 0`, `missing_days: 0`, and `insufficient_sample: false` survive terse shaping.
- Consider deciding in the helper implementation how to handle negative `n` or `missing_days` and blank/whitespace `method` values, either by validation errors or documented normalization.

No plan changes are required before proceeding to Step 2.
