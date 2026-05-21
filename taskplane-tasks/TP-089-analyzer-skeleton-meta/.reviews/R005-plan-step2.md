# Plan Review R005 — Step 2: Implement skeleton helpers

Verdict: **APPROVE**

I read `PROMPT.md`, the updated `STATUS.md`, the prior R004 plan review, the existing `internal/analysis` package, `internal/response/{shape,meta}.go`, the TP-088 formula constants, and the PRD analyzer response rules. The Step 2 plan now has the concrete helper API, package split, normalization policy, terse/full ownership, and golden-test strategy that were missing in R004.

## What is now clear enough to implement

- `internal/analysis` owns the analyzer-domain contract via `AnalyzerMetaInput`, `NewAnalyzerMeta`, `NormalizeSourceTools`, `InsufficientSample`, and constants for `MissingActionSkip`, `MinBaselineSamples`, and `MinCorrelationSamples`.
- The normalization behavior is explicit: trim method/missing action, clamp negative counts to zero, normalize source tools deterministically, default missing action to `skip`, keep `source_tools` non-nil, and avoid duplicating TP-088 formula URI fragments.
- The plan correctly does **not** rely on `response.Shape` to drop heavy analyzer payloads. Tool-level response assembly will omit `series` unless `include_full` is true before shaping.
- Goldens are planned as post-shape fixtures, which is the right boundary because it proves what MCP clients see after common `_meta` is merged.
- The risky mandatory `_meta` edge cases are listed for coverage: empty source tools as `[]`, zero/false mandatory fields preserved, default missing action, deterministic source tool order, and formula refs imported from `internal/resources` constants.

## Non-blocking implementation notes

- Keep the `internal/tools` helper intentionally small. It should only assemble analyzer response envelopes and include-full behavior; all analyzer meta semantics should remain in `internal/analysis`.
- When implementing `FormulaRef`, be consistent with the status wording. Trimming surrounding whitespace is fine as input normalization, but do not validate, rewrite, or duplicate the TP-088 URI constants in `internal/analysis`.
- Make the golden `response.Options` deterministic enough to avoid churn from response-owned `_meta` such as catalog/version/delete-mode/toolset defaults.

This plan is sufficient for Step 2 to proceed.
