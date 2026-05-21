# Plan Review R004 — Step 2: Implement skeleton helpers

Verdict: **REVISE**

I read `PROMPT.md`, the current `STATUS.md`, prior reviews, the existing `internal/analysis` package, `internal/response/{shape,meta}.go`, the TP-088 formula constants, and the PRD analyzer rules. Step 1 now records a good contract, but Step 2 still only has the generic checklist items. For this foundational helper step, the plan needs a concrete implementation shape before coding proceeds.

## Required plan clarifications

1. **Name the helper API and package split.**
   The plan should state the exact small helpers to add, not just "builders". For example, record whether `internal/analysis` will expose something like `AnalyzerMetaInput`, `NewAnalyzerMeta(input) AnalyzerMeta`, `NormalizeSourceTools`, `InsufficientSample`, and constants `MissingActionSkip`, `MinBaselineSamples`, `MinCorrelationSamples`. If `internal/tools/analyzer_common*.go` is needed, state what it owns and keep it limited to tool response envelopes; do not let it become a second source of analyzer-domain semantics.

2. **Specify validation/normalization behavior.**
   Step 1 documented desired semantics but not the implementation policy for bad inputs. The Step 2 plan should say how helpers handle:
   - blank/whitespace `method`;
   - negative `n` or `missing_days`;
   - blank, duplicated, or unordered `source_tools`;
   - missing `missing_action` (default to `skip`);
   - optional `formula_ref` values imported by callers from `internal/resources` constants rather than copied strings.

   Either returning an error or clamping/normalizing can be acceptable if documented, but the choice must be explicit because downstream analyzer tasks will rely on it.

3. **Do not rely on `response.Shape` to enforce terse/full analyzer payloads.**
   The existing shaper strips nulls and adds common `_meta`; it does not automatically drop a non-null `series[]`. The plan must state how the skeleton response helper will omit heavy per-bucket/per-sample fields unless `include_full: true` was requested. Golden coverage should prove both paths after `response.Shape`, including that mandatory analyzer `_meta` fields survive alongside response-owned keys.

4. **Define the golden-test strategy and fixture locations.**
   The plan should identify where test-only helpers and fixtures live, such as `internal/analysis/meta_test.go` plus `internal/analysis/testdata/*.golden.json`, or tool-level fixtures if the response envelope belongs in `internal/tools`. It should also state whether goldens are pre-shape or post-shape. Prefer post-shape fixtures with deterministic `response.Options` so the contract pins what MCP clients see.

5. **Pin the mandatory `_meta` edge cases in Step 2/3 planning.**
   The upcoming tests need to cover the risky cases from the contract:
   - `source_tools` is `[]`, not `null`, when omitted;
   - `n: 0`, `missing_days: 0`, and `insufficient_sample: false` are present in terse JSON;
   - `missing_action` defaults to `"skip"`;
   - source tools are trimmed, deduped, and sorted deterministically;
   - formula refs are passed through from TP-088 constants without duplicating URI fragments.

## Why this blocks the plan

This task is the contract all v0.6 analyzers build on. If Step 2 starts with only the generic checklist, it is too easy to produce helpers that technically satisfy one demo response but leave downstream tasks guessing about construction, invalid inputs, and terse/full behavior. Updating `STATUS.md` with the concrete helper API and test/golden strategy should be enough for approval; no product-scope change is needed.
