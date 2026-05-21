# Plan Review R007 — Step 3: Add a no-op/demo test path

Verdict: **REVISE**

I read `PROMPT.md`, `STATUS.md`, the prior Step 2 reviews, and the current analyzer helper files. The only Step 3 "plan" recorded in `STATUS.md` is the original checklist, with no concrete file/test strategy. For this task, that is not enough because Step 3 is the contract-pinning step for the mandatory analyzer `_meta` shape.

## Findings

1. **No concrete Step 3 plan is recorded.**  
   `STATUS.md` marks Step 3 as in progress, but it does not specify which tests will be added or what cases they will cover beyond the prompt's generic bullets. Before implementation, record a short plan that names the target files and assertions. At minimum it should identify:
   - `internal/analysis/meta_test.go` for helper-level contract/normalization tests.
   - `internal/tools/analyzer_common_test.go` for a demo/no-op shaped analyzer response that exercises the existing `internal/tools/testdata/analyzer/*.golden.json` fixtures.
   - The targeted test command, likely `go test ./internal/analysis ./internal/tools`.

2. **The plan must explicitly cover the mandatory-zero-value `_meta` contract.**  
   Step 1 documented that `method`, `source_tools`, `n`, `missing_days`, `missing_action`, and `insufficient_sample` are mandatory and non-`omitempty`, including zero/false/empty values. The Step 3 plan should include a test that marshals or shapes a minimal analyzer response and proves these keys remain present, especially `source_tools: []` rather than `null` and `insufficient_sample: false` rather than omitted.

3. **The plan must pin terse/full behavior at the post-shape boundary.**  
   Step 2 added `shapeAnalyzerResponse`/goldens, but there is not yet a `Test*` that calls them. The Step 3 plan should state that terse output omits `series`, full output includes `series`, and both outputs preserve analyzer `_meta` after `response.Shape` merges response-owned keys.

4. **The plan should include missing-day/default-missing-action and insufficient-sample cases.**  
   The prompt requires missing-day handling, and Step 1/2 established `missing_action` defaults to `skip` and `n` controls `insufficient_sample`. The plan should name concrete table cases such as `missing_days > 0`, blank `MissingAction` defaulting to `skip`, negative counts clamped to zero, and `n < MinSamples` yielding `insufficient_sample: true`.

5. **STATUS review table is inconsistent with the actual R006 verdict.**  
   `STATUS.md` lists R006 as `APPROVE`, but `.reviews/R006-code-step2.md` says `Verdict: REVISE`. If the R006 finding has been fixed, the status should still preserve the actual review verdict and add a follow-up execution-log/status note for the fix, rather than rewriting the historical review result. Please reconcile this before relying on Step 2 as complete.

## Suggested revised Step 3 plan

- Add table-driven tests in `internal/analysis/meta_test.go` for:
  - mandatory zero-value JSON keys, including non-nil empty `source_tools`;
  - source-tool trim/dedupe/sort;
  - default `missing_action: "skip"` and positive `missing_days`;
  - negative count clamping;
  - insufficient-sample threshold behavior.
- Add `TestShapeAnalyzerResponseTerseAndFull` (or equivalent) in `internal/tools/analyzer_common_test.go` that builds a demo `analyzerResponseInput`, imports the TP-088 formula ref constant from `internal/resources`, shapes with deterministic response metadata, and compares against `testdata/analyzer/demo_terse.golden.json` and `demo_full.golden.json`.
- Assert directly, in addition to goldens if helpful, that terse output has no `series` and full output has `series`.
- Run `go test ./internal/analysis ./internal/tools` and record the result in `STATUS.md`.

Once those details are recorded, the Step 3 implementation should be low risk.
