# R016 Plan Review — Step 4: Tests and verification

**Verdict:** REVISE

The Step 4 section in `STATUS.md` is still only the original checklist, not an implementation plan. For a large public analyzer-tool cluster, the plan needs to name the deterministic fixtures, regression coverage, generated-doc/catalog updates, and exact verification commands before implementation proceeds. It also needs to reconcile the current review-history inconsistency: `STATUS.md` marks R012/R015 as approved, but the checked-in review files are `REVISE` and list concrete correctness/test failures.

## Required revisions

1. **Resolve the Step 2/Step 3 review-state contradiction before Step 4.**  
   `.reviews/R012-code-step2.md` and `.reviews/R015-code-step3.md` both say `REVISE`, while `STATUS.md` records them as approved and marks the steps complete. The Step 4 plan must either document that those findings have already been fixed in a later commit/re-review, or include explicit regression tests/fixes for them. At minimum, cover:
   - activity daily pace/speed aggregation from summed distance/time, not weighted mean of per-activity pace;
   - weekly trend slope using real weekly bucket indexes;
   - correlation `_meta.n` / result `n` after invalid-pair filtering;
   - generated tool catalog golden and safety static/adversarial catalog updates for the four new read tools;
   - `SourceDerivedWeekly` loading for `weekly_tss` / `weekly_hours`;
   - cancellation propagation on baseline/second-series loads;
   - distribution validation for mutually exclusive `bucket_count`/`buckets` and quantiles outside `0..1`.

2. **Specify analyzer math golden tests, not just “fixtures/golden tests.”**  
   Add a concrete table for `internal/analysis` tests covering each compute helper:
   - trend: daily rolling mean/slope/delta/z-score, missing samples, insufficient current/baseline, zero baseline/stddev boundaries, and weekly missing-bucket slope;
   - distribution: min/max/mean/sample stddev, R-7 quantiles, explicit buckets with below/above range, equal-value bucket expansion, invalid/insufficient samples;
   - correlation: Pearson, Spearman with ties, lag semantics, zero variance, invalid-pair filtering, and insufficient paired rows;
   - efforts delta: power, HR, metric pace, imperial pace, missing current/baseline buckets, activity-id propagation, percent-delta zero-baseline boundary, and `n` as comparable buckets only.

3. **Specify tool-adapter tests for the public contracts.**  
   The plan should list `internal/tools` tests using stubbed clients for all four tools. Required coverage includes strict request validation, unknown metric rejection through `analysis.ParseMetric`, source-tool metadata, mandatory analyzer `_meta` fields, terse omission of `series`, `include_full` opt-in, sport filter compatibility, daily-vs-activity grain metadata, lagged correlation y-window expansion, weekly metric support/rejection rules, and short user-facing errors. These are public MCP contracts, not only pure math behavior.

4. **Clarify the auto-lap item for `analyze_efforts_delta`.**  
   The approved Step 1/R008 contract says efforts delta uses only best-effort curve endpoints and does not consume intervals. The Step 4 plan should mark auto-lap propagation as not applicable under the current implementation, and preferably add a negative/stub assertion that no interval source is called. Do not add an interval dependency just to satisfy the checkbox.

5. **Plan generated docs/catalog verification explicitly.**  
   Step 3 deferred generated docs/CHANGELOG. Step 4 should name the artifacts and tests: run the docs generation path (`make docs-tools` or the repository-equivalent), update `web/content/reference/tools.md`, update generated catalog goldens, update any static safety catalog/matrix entries as `RequirementRead`, and add a `[Unreleased]` `CHANGELOG.md` entry.

6. **Disambiguate Step 4 vs Step 5 quality gates.**  
   The prompt duplicates full verification in Step 4 and Step 5. The plan should state which commands are run in Step 4 versus repeated/finalized in Step 5. A reasonable Step 4 gate is targeted packages plus affected golden/safety tests; Step 5 can be the final `make test`, `make build`, and `make lint`. If Step 4 intends to run the full gate too, list the exact commands and require failures to be fixed or recorded in `STATUS.md`.

## Suggested Step 4 plan shape

- Add pure computation tests in `internal/analysis/{trend,distribution,correlation,efforts_delta}_test.go` with hand-calculated expected values.
- Add public tool adapter tests in `internal/tools/analyze_*_test.go` with deterministic stub clients and no network access.
- Add regression tests for every R012/R015 finding that is not already covered by the above.
- Update catalog/generated docs/safety fixtures and `CHANGELOG.md`.
- Run targeted verification such as:
  - `go test ./internal/analysis ./internal/tools ./internal/toolcatalog ./internal/safety ./cmd/gendocs`
  - `make docs-tools` if needed before golden checks
  - optionally `make test` here if Step 4 keeps the full-gate checkbox literal.
- Record results and any pre-existing unrelated failures in `STATUS.md`.

## Tests run

Not run; this was a plan review only. Reviewed `PROMPT.md`, `STATUS.md`, and prior review files for the current task.
