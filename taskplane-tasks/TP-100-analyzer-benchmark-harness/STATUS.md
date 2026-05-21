# TP-100: Extend KR5 benchmark harness for analyzer family — Status

**Current Step:** Step 6: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-05-20
**Review Level:** 2
**Review Counter:** 22
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

### Step 1: Find and extend benchmark harness
**Status:** ✅ Complete

- [x] Locate the v0.4 prompt set and token/response measurement code.
- [x] Define fair analyzer-enabled/disabled comparison modes that hold all non-analyzer tools constant.
- [x] Pin concrete v2 fixture/result schema, per-mode token metrics, and validation rules for per-prompt/per-mode call-plan coverage.
- [x] Document that the harness is call-plan based: identical prompt text maps to analyzer vs fetch-and-reduce call plans, not autonomous model tool choice.
- [x] Add analyzer-enabled and analyzer-disabled benchmark modes.
- [x] Upgrade the result schema to v2 with benchmark modes, per-call mode/source-tool usage, and stream-pull metrics.
- [x] Define deterministic raw-stream pull counting from harness ToolCall rows, with diagnostics log parsing explicitly avoided for fixture mode.
- [x] R005: compare canonical non-analyzer catalog payloads across analyzer modes, not just tool names.
- [x] R005: compute response token metrics from audited/stripped payloads instead of benchmark-only redaction_audit metadata.

---

### Step 2: Add analyzer prompt shapes
**Status:** ✅ Complete

- [x] Add trend, zone-time distribution, baseline, correlation/compliance, and single-activity histogram prompt cases scoped to the analyzer benchmark fixture.
- [x] Add harness support for prompt `server_scope` filtering so analyzer-only prompts do not require legacy fixture coverage.
- [x] Add validation for prompt-level expected source-tool usage by mode.
- [x] Add an `icuvisor-analyzer-family` fixture with mode-specific catalogs and paired enabled/disabled call rows for each analyzer prompt.
- [x] Ensure prompts are identical across with/without-analyzer runs except tool availability and call-plan rows.
- [x] Record expected source-tool usage for analyzer-enabled calls and fallback source-tool plans for analyzer-disabled calls.
- [x] R010: make analyzer-enabled fixture catalog include the full analyzer family with non-stub descriptions and input schemas.

---

### Step 3: Run and record results
**Status:** ✅ Complete

- [x] Run the fixture benchmark and update `scripts/benchmark/results/kr5-results.json` with v2 output.
- [x] Record analyzer enabled/disabled token deltas and raw-stream pull counts in `docs/kr5-benchmark.md`.
- [x] Calculate per-candidate net response-token savings for `analyze_trend`, `compute_zone_time`, and `compute_baseline` using paired prompt rows.
- [x] Flag whether `analyze_trend`, `compute_zone_time`, and `compute_baseline` meet net-savings criteria for TP-098.
- [x] R014: correct TP-098 candidate catalog-token costs to use incremental catalog deltas.
- [x] R014: fix R013 execution-history verdict to ACCEPT.

---

### Step 4: Verify
**Status:** ✅ Complete

- [x] Add no-network Python unittest coverage for analyzer-mode validation and v2 result invariants.
- [x] Run `python3 -m unittest discover -s scripts/benchmark -p '*_test.py'`.
- [x] Run deterministic fixture freshness check and compare temp output to `scripts/benchmark/results/kr5-results.json`.
- [x] Defer full project `make test`/`make build`/`make lint` gates to Step 5 and record any benchmark-specific failures now.
- [x] Update CHANGELOG.md for benchmark/docs visibility.

---


### Step 5: Testing & Verification
**Status:** ✅ Complete

- [x] Targeted Python benchmark unittest passing: `python3 -m unittest discover -s scripts/benchmark -p '*_test.py'`
- [x] Deterministic fixture freshness check matches committed `scripts/benchmark/results/kr5-results.json`
- [x] FULL test suite passing: `make test`
- [x] Build passes: `make build`
- [x] Lint passes: `make lint`
- [x] All failures fixed or documented as pre-existing unrelated failures
- [x] R021: correct R020 execution-history verdict to ACCEPT.
- [x] R021: add Step 5 command/outcome rows to the Execution Log.

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
| R001 | plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |
| R002 | plan | 1 | REVISE | `.reviews/R002-plan-step1.md` |
| R003 | plan | 1 | REVISE | `.reviews/R003-plan-step1.md` |
| R004 | plan | 1 | UNAVAILABLE | _reviewer unavailable_ |
| R005 | code | 1 | REVISE | `.reviews/R005-code-step1.md` |
| R006 | code | 1 | UNAVAILABLE | _reviewer unavailable_ |
| R007 | plan | 2 | REVISE | `.reviews/R007-plan-step2.md` |
| R008 | plan | 2 | REVISE | `.reviews/R008-plan-step2.md` |
| R009 | plan | 2 | ACCEPT | `.reviews/R009-plan-step2.md` |
| R010 | code | 2 | REVISE | `.reviews/R010-code-step2.md` |
| R011 | code | 2 | ACCEPT | `.reviews/R011-code-step2.md` |
| R012 | plan | 3 | REVISE | `.reviews/R012-plan-step3.md` |
| R013 | plan | 3 | ACCEPT | `.reviews/R013-plan-step3.md` |
| R014 | code | 3 | REVISE | `.reviews/R014-code-step3.md` |
| R015 | code | 3 | ACCEPT | `.reviews/R015-code-step3.md` |
| R016 | plan | 4 | REVISE | `.reviews/R016-plan-step4.md` |
| R017 | plan | 4 | ACCEPT | `.reviews/R017-plan-step4.md` |
| R018 | code | 4 | ACCEPT | `.reviews/R018-code-step4.md` |
| R019 | plan | 5 | REVISE | `.reviews/R019-plan-step5.md` |
| R020 | plan | 5 | ACCEPT | `.reviews/R020-plan-step5.md` |
| R021 | code | 5 | REVISE | `.reviews/R021-code-step5.md` |
| R022 | code | 5 | ACCEPT | `.reviews/R022-code-step5.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Full Go gate exposed stale `compute_baseline` helper names colliding with shared analyzer helpers. | Fixed by prefixing baseline-local helpers and aligning catalog summary tests. | `internal/tools/compute_baseline.go`, `internal/tools/catalog_test.go` |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-20 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-20 21:54 | Task started | Runtime V2 lane-runner execution |
| 2026-05-20 21:54 | Step 0 started | Preflight |
| 2026-05-20 23:31 | Targeted Python benchmark unittest | PASS: `python3 -m unittest discover -s scripts/benchmark -p '*_test.py'` (7 tests) |
| 2026-05-20 23:32 | Fixture freshness check | PASS: fixed-timestamp benchmark output matched `scripts/benchmark/results/kr5-results.json` |
| 2026-05-20 23:35 | Full test suite | PASS: `make test` |
| 2026-05-20 23:36 | Build gate | PASS: `make build` |
| 2026-05-20 23:36 | Lint gate | PASS: `make lint` |
| 2026-05-20 23:43 | Worker iter 1 | done in 6556s, tools: 298 |
| 2026-05-20 23:43 | Task complete | .DONE created |

---

## Blockers

*None*

---

## Notes

- R001 suggestions: keep fixture mode deterministic and credential-free; validate analyzer coverage per mode; keep prompt text identical across modes; store expected source-tool usage in fixture/config metadata without live athlete payloads.
- Step 1 current design: add a synthetic analyzer benchmark server family with two mode-specific catalogs: `analyzers_enabled` is the same base catalog plus analyzer-family tools, and `analyzers_disabled` is the same base catalog minus analyzer-family tools. Validation compares the catalogs to prove non-analyzer tool names are identical and disabled mode exposes no analyzer tools. Fixture calls gain required `mode` and optional `source_tool_usage` as a list of source tool names/counts while prompt IDs/text remain shared. Result JSON moves to `schema_version: kr5-benchmark-result-v2` with top-level `benchmark_modes`, per-call `mode` and `source_tool_usage`, and per-server `mode_summaries` keyed by mode containing catalog description token count, response token count, median response bytes, median response tokens, and `raw_stream_pull_count`. Raw stream pulls are LLM-visible harness rows whose tool name is exactly `get_activity_streams` or a configured reference alias such as `icu_get_activity_streams`/`reference:get_activity_streams`, excluding `unavailable:*` placeholders and excluding analyzer `_meta.source_tools`. Validation must require every analyzer prompt/intent to have rows for both modes, disabled rows to avoid `analyze_*`, `compute_*`, `get_activity_histogram`, and `get_fitness_projection`, and enabled rows to call expected analyzer-family tools unless marked unavailable/error.
- R001/R002/R003 suggestions retained: keep fixture mode deterministic and credential-free; validate analyzer coverage per mode; keep prompt text identical across modes; store expected source-tool usage in fixture/config metadata without live athlete payloads; do not overclaim autonomous model tool-selection behavior.
- R005 suggestion: allow explicit unavailable/error rows to satisfy expected analyzer-tool validation when fixtures intentionally model an unavailable analyzer case.
- Step 5 plan: rerun targeted Python benchmark unittest, rerun fixed-timestamp fixture freshness check without `--allow-approx-tokenizer`, then run `make test`, `make build`, and `make lint`. Fix failures when in scope; if a full project gate fails due to pre-existing unrelated Go compile/lint state or missing local tooling, record the exact command and error in STATUS before checking the failure-documentation box.
- Step 4 plan: add a no-network stdlib unittest file under `scripts/benchmark/` that exercises scoped analyzer prompt validation, analyzer catalog fairness, expected `source_tool_usage`, raw-stream pull counting, and committed v2 analyzer result invariants. Run `python3 -m unittest discover -s scripts/benchmark -p '*_test.py'`, then run the fixed-timestamp fixture benchmark to `/tmp/kr5-results-check.json` without `--allow-approx-tokenizer` and compare it to the committed result. Step 4 will not run the full Go `make test`/`make build`/`make lint` gates; Step 5 is the final project quality gate. Add a concise `[Unreleased]` changelog bullet for the analyzer benchmark/report and do not claim core promotion.
- Step 3 plan: run `python3 scripts/benchmark/kr5_benchmark.py --mode fixtures --prompt-set scripts/benchmark/prompts/kr5_shared_prompts.json --fixture-dir scripts/benchmark/testdata/fixtures --output scripts/benchmark/results/kr5-results.json --generated-at 2026-05-20T00:00:00Z` without `--allow-approx-tokenizer` for committed evidence. Update `docs/kr5-benchmark.md` methodology for v2 `benchmark_modes`, `mode_summaries`, `response_tokens`, scoped analyzer prompts, and raw-stream pull counts while preserving the historical KR5 comparison separately from the analyzer fixture. For TP-098, compute candidate gross savings from paired prompt rows using response tokens: `KR5-A01` disabled `get_activities` + `get_activity_streams` vs enabled `analyze_trend`; `KR5-A02` disabled `get_activity_streams` vs enabled `compute_zone_time`; `KR5-A03` disabled `get_activities` + `get_fitness` vs enabled `compute_baseline`. Record net verdict as gross response-token savings minus the candidate analyzer tool's incremental catalog-description tokens; `meets` means net savings > 0 and enabled mode has no LLM-visible raw-stream pull for that prompt. Also report aggregate mode context (`analyzers_disabled` vs `analyzers_enabled`) but do not use the full analyzer-family aggregate alone as the per-candidate promotion verdict.
- Step 2 plan: add analyzer prompts to `kr5_shared_prompts.json` with prompt-level `benchmark_modes` and `server_scope: ["icuvisor-analyzer-family"]`, while unscoped prompts continue to apply to every fixture. Update validation so scoped-out prompts do not require coverage for that measurement and result `prompt_count` remains the shared prompt-set size while per-server call counts reflect applicable rows. Add prompt-level `expected_source_tools_by_mode` as `{mode: {called_tool: [source tools...]}}` and validate that each matching call's `source_tool_usage` contains exactly those tool names/counts unless the prompt/mode has an explicit unavailable row. Add a new synthetic `icuvisor-analyzer-family` fixture whose `mode_catalogs` share identical non-analyzer payloads, with `analyzers_enabled` adding only analyzer-family tools and `analyzers_disabled` using fetch-and-reduce source tools; keep prompt text single-copy per prompt and put mode differences only in fixture call rows and expected metadata. Planned cases: `KR5-A01` trend (`analyze_trend` vs `get_activities`+`get_activity_streams`), `KR5-A02` zone-time distribution for TP-098 (`compute_zone_time` vs `get_activity_streams` raw binning), `KR5-A03` baseline (`compute_baseline` vs wellness/fitness/activity pulls), `KR5-A04` correlation/compliance (`analyze_correlation` + `compute_compliance_rate` vs wellness/events/activity pulls), and `KR5-A05` histogram (`get_activity_histogram` vs `get_activity_streams`).
| 2026-05-20 21:57 | Review R001 | plan Step 1: REVISE |
| 2026-05-20 22:00 | Review R002 | plan Step 1: REVISE |
| 2026-05-20 22:03 | Review R003 | plan Step 1: REVISE |
| 2026-05-20 21:59 | Review R002 | plan Step 1: REVISE |
| 2026-05-20 22:01 | Review R003 | plan Step 1: REVISE |
| 2026-05-20 22:10 | Review R005 | code Step 1: REVISE |
| 2026-05-20 22:26 | Review R007 | plan Step 2: REVISE |
| 2026-05-20 22:29 | Review R008 | plan Step 2: REVISE |
| 2026-05-20 22:31 | Review R009 | plan Step 2: ACCEPT |
| 2026-05-20 22:40 | Review R010 | code Step 2: REVISE |
| 2026-05-20 22:46 | Review R011 | code Step 2: ACCEPT |
| 2026-05-20 22:49 | Review R012 | plan Step 3: REVISE |
| 2026-05-20 22:51 | Review R013 | plan Step 3: ACCEPT |
| 2026-05-20 23:04 | Review R014 | code Step 3: REVISE |
| 2026-05-20 23:07 | Review R015 | code Step 3: ACCEPT |
| 2026-05-20 23:10 | Review R016 | plan Step 4: REVISE |
| 2026-05-20 23:12 | Review R017 | plan Step 4: ACCEPT |
| 2026-05-20 23:24 | Review R018 | code Step 4: ACCEPT |
| 2026-05-20 23:26 | Review R019 | plan Step 5: REVISE |
| 2026-05-20 23:27 | Review R020 | plan Step 5: ACCEPT |
| 2026-05-20 23:37 | Review R021 | code Step 5: REVISE |
| 2026-05-20 23:41 | Review R022 | code Step 5: UNKNOWN |
