# TP-094: `compute_activity_segment_stats` raw-stream analyzer — Status

**Current Step:** Step 6: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-05-20
**Review Level:** 2
**Review Counter:** 20
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

### Step 1: Define segment selection and metrics
**Status:** ✅ Complete

- [x] Define request fields for activity ID, metric, time range and/or distance range.
- [x] Specify validation for ambiguous or out-of-range segments.
- [x] Define supported stats and formula refs for decoupling/drift/NP/IF.
- [x] Address R001 plan review: add IF FTP schema, align metric enum to canonical keys, resolve single-stat metadata semantics, NP/IF formula-ref policy, time-based NP eligibility, and stricter split-half eligibility.
- [x] Address R002 plan review: allow zero watts in NP/IF, define `_meta.n` by stat, and specify deterministic 30-second rolling-window averaging.

Step 1 design notes:
- Request shape will be `activity_id` plus exactly one segment selector: either `start_seconds`/`end_seconds` or `start_distance_m`/`end_distance_m`, with optional `include_full` defaulting false. The schema uses one `stat` (not `stats[]`) from a closed enum (`mean`, `median`, `p90`, `decoupling`, `drift`, `np`, `if`) so top-level analyzer `_meta.formula_ref`, `n`, and `insufficient_sample` describe exactly one calculation. `metric` is required only for `mean`/`median`/`p90` and must be a canonical stream key enum (`watts`, `heart_rate`, `cadence`, `velocity_smooth`, `distance`, `time`); derived stats reject non-empty/incompatible `metric`. `ftp_watts` is required for `if` only, must be positive/finite, and is rejected for every other `stat`.
- Validation will reject blank `activity_id`, unknown metric/stat, missing segment selector, both time and distance selectors at once, partial ranges, non-finite/negative bounds, end <= start, ranges outside stream coverage, missing required streams, and mismatched stream lengths. Distance selectors are normalized to meters, time selectors to seconds. Segment inclusion is intentionally `start <= axis <= end`; adjacent segment calls may share boundary samples, and that is documented rather than hidden.
- Supported calculations: mean/median/p90 use the requested metric stream with `_meta.formula_ref` empty; `decoupling` uses Pw:HR over first/second elapsed-time halves with `resources.AnalysisFormulaRefPwHRDecoupling`; `drift` uses HR drift over first/second elapsed-time halves with `resources.AnalysisFormulaRefHRDrift`; `np` uses a 30-second rolling average power fourth-power calculation over elapsed seconds with `_meta.formula_ref` empty because TP-088 currently has no NP ref; `if` uses that NP divided by caller-provided `ftp_watts` with `_meta.formula_ref` empty because TP-088 currently has no IF ref. Formula policy is explicit: do not reuse EF/VI refs for NP/IF; method text records the exact NP/IF formula until a future formula-resource expansion adds dedicated refs.
- Eligibility: split-half calculations require the time stream plus required value streams, at least two finite paired samples per half, positive average HR denominators, and for Pw:HR positive average power and first-half ratio denominator. NP/IF require time and watts streams spanning at least 30 elapsed seconds inside the selected segment; watts samples must be finite and non-negative (zero-watt coasting samples are valid and included, while negative/non-finite values are skipped from windows and counted out of `_meta.n`). NP/IF rolling windows are deterministic simple averages of finite non-negative watts samples whose timestamps satisfy `window_end-30 < t <= window_end`; `_meta.n` is the number of valid rolling windows for NP/IF, the number of finite metric samples for mean/median/p90, and the number of finite paired samples across both halves for decoupling/drift. `include_full:false` never returns raw samples; `include_full:true` returns only sliced canonical inputs/calculation points needed to audit the stat, not unrelated activity streams.

---

### Step 2: Implement stream slicing and stats
**Status:** ✅ Complete

- [x] Fetch canonical stream keys only for requested metrics.
- [x] Slice by time or distance using normalized units.
- [x] Compute mean/median/p90 and formula-based derived stats with minimum samples.
- [x] Address R005 plan review: define percentile/scalar minimums, deterministic error-vs-insufficient behavior, and exact canonical stream fetch/mapping rules.
- [x] Address R007 code review: keep top-level analyzer `_meta.insufficient_sample` aligned with calculator insufficiency for per-half/denominator failures.
- [x] Address R007 code review: add decoupling audit series for `include_full:true` without unrelated raw streams.
- [x] Address R007 code review: validate bounds and `ftp_watts` compatibility before stream fetch and remove the `round` unparam lint issue.
- [x] Address R008 code review: omit `time` from scalar distance stream requirements unless the requested metric is `time`.
- [x] Address R008 code review: expose missing required stream keys in short public errors.
- [x] Address R008 code review: treat in-coverage ranges with no samples as insufficient analyzer results, not out-of-range errors.
- [x] Address R009 code review: reject any provided `ftp_watts` for non-IF stats before fetching streams.
- [x] Address R009 code review: return a specific public out-of-coverage segment range error at the tool boundary.
- [x] Address R010 code review: add JSON tags for public segment response fields.
- [x] Address R010 code review: split time-selected drift/decoupling halves at requested elapsed segment endpoints instead of observed sample endpoints.
- [x] Address R011 code review: anchor NP/IF 30-second windows to requested elapsed start for time-selected segments.
- [x] Address R011 code review: require finite selected time samples for distance-selected split-half and rolling-window derived stats.

Step 2 implementation notes:
- Add `internal/analysis/segment_stats.go` with a pure calculator API that accepts canonical stream slices keyed by canonical stream key. Keep intervals.icu client access in `internal/tools`; the analysis package only validates/slices samples and computes results.
- Required canonical streams by `stat`: mean/median/p90 fetch selector axis plus requested metric; drift fetch time + heart_rate plus distance if distance-selected; decoupling fetch time + heart_rate + watts plus distance if distance-selected; NP/IF fetch time + watts plus distance if distance-selected.
- The slicer validates axis/value lengths before calculation, normalizes distance bounds in meters and time bounds in seconds, and filters to inclusive bounds. Invalid arguments, missing required streams, mismatched stream lengths, and ranges outside selector-axis coverage return short user-facing errors. Valid slices with too few usable samples return an analyzer payload with `_meta.insufficient_sample:true`, `_meta.n` set to the usable sample/window count, and no fabricated numeric value.
- Scalar stats use finite values only, require at least one sliced sample, set `_meta.n` to the finite count, compute mean as arithmetic average, median by sorting and averaging the two middle values for even `n`, and p90 by nearest-rank (`ceil(0.90*n)`, 1-indexed, clamped to n). These deterministic conventions are covered by pure analysis tests.
- Derived stat methods follow Step 1 semantics: elapsed-time halves for drift/decoupling, finite paired sample counts for `_meta.n`, simple 30-second elapsed-time rolling windows for NP/IF, finite non-negative watts with zero values included, and positive `ftp_watts` for IF. Tool fetch code dedupes required canonical keys, converts them to explicit upstream `types` spellings (`time`, `distance`, `watts`, `heartrate`, `cadence`, `velocity_smooth`), calls `GetActivityStreams` without broad default expansion, canonicalizes returned `type`/`name` through `internal/streams`, and errors if any required canonical key is absent.

---

### Step 3: Register tool and tests
**Status:** ✅ Complete

- [x] Register in `full` and describe as the raw-stream exception.
- [x] Add fixtures for time segment, distance segment, missing stream, and insufficient sample.
- [x] Assert mandatory meta and source_tools.
- [x] Address R013 plan review: update shared `internal/toolcatalog` known/athlete-scoped names, catalog grouping/tier invariants, and public catalog/full-only assertions.
- [x] Address R015 code review: add `compute_activity_segment_stats` to the static safety catalog matrix as a read tool.
- [x] Address R015 code review: regenerate/update the generated tool catalog golden JSON for the new analyzer group/tool.

Step 3 implementation notes:
- Wire `newComputeActivitySegmentStatsTool` into `registryBaseTools` as a `full` read tool next to `get_activity_streams`, map it to an `analyzers` catalog group (including allowed-group expectations), and remove only `compute_activity_segment_stats` from the analyzer ghost assertions now that it is intentionally registered.
- Add `compute_activity_segment_stats` to the shared `internal/toolcatalog` known/athlete-scoped read-tool catalog so `defaultRegistry.Register` accepts it and coach ACL/catalog invariants continue to pass.
- Extend tool tests to cover time-selected and distance-selected successful calls through the handler, missing stream user errors, insufficient-sample analyzer payloads, terse/full behavior, and narrow upstream stream types. Terse success must have no `series`/raw samples; `include_full:true` must include only sliced audit inputs.
- Assert analyzer `_meta` includes mandatory `method`, `source_tools:["get_activity_streams"]`, `n`, `missing_days`, `missing_action`, `insufficient_sample`, and formula refs where applicable. Cover drift/decoupling formula refs and NP/IF empty formula-ref policy.
- Lock observable full-only placement through registry and catalog tests: `EffectiveToolset()==full`, `Catalog()` tier is `full`, and the raw-stream-exception summary is visible via catalog/advanced-capabilities surfaces.

---

### Step 4: Verify
**Status:** ✅ Complete

- [x] Run stream canonicalization tests plus full suite/build/lint.
- [x] Update docs/reference and CHANGELOG.md.
- [x] Record performance/token considerations in STATUS.md.

Step 4 verification notes:
- Run focused stream canonicalization/analyzer/catalog tests first, then `make test`, `make build`, and `make lint` for the step-level verification gate.
- Update the generated/public tool reference and CHANGELOG `[Unreleased]` for the newly registered full-only analyzer.
- Record token/performance behavior: terse responses omit raw samples; `include_full` includes only sliced audit inputs; upstream fetches are narrowed to required canonical streams.

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

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| No out-of-scope discoveries; implementation stayed within TP-094 analyzer/catalog/docs scope. | Logged for delivery | STATUS.md |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-20 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-20 17:20 | Task started | Runtime V2 lane-runner execution |
| 2026-05-20 17:20 | Step 0 started | Preflight |
| 2026-05-20 18:38 | Worker iter 1 | done in 4667s, tools: 286 |
| 2026-05-20 18:38 | Task complete | .DONE created |

---

## Blockers

*None*

---

## Notes

- Performance/token considerations: `compute_activity_segment_stats` narrows upstream stream fetches to the selector axis plus stat-required canonical streams, uses `IncludeDefaults:false`, omits raw stream samples in terse mode, and exposes only sliced/calculation audit inputs through `include_full:true`.

*Reserved for execution notes*
| 2026-05-20 17:24 | Review R001 | plan Step 1: REVISE |
| 2026-05-20 17:26 | Review R002 | plan Step 1: REVISE |
| 2026-05-20 17:28 | Review R003 | plan Step 1: APPROVE |
| 2026-05-20 17:30 | Review R004 | code Step 1: APPROVE |
| 2026-05-20 17:32 | Review R005 | plan Step 2: REVISE |
| 2026-05-20 17:35 | Review R006 | plan Step 2: APPROVE |
| 2026-05-20 17:42 | Review R007 | code Step 2: UNKNOWN |
| 2026-05-20 17:50 | Review R008 | code Step 2: UNKNOWN |
| 2026-05-20 17:56 | Review R009 | code Step 2: UNKNOWN |
| 2026-05-20 18:01 | Review R010 | code Step 2: UNKNOWN |
| 2026-05-20 18:07 | Review R011 | code Step 2: UNKNOWN |
| 2026-05-20 18:14 | Review R012 | code Step 2: APPROVE |
| 2026-05-20 18:16 | Review R013 | plan Step 3: REVISE |
| 2026-05-20 18:18 | Review R014 | plan Step 3: APPROVE |
| 2026-05-20 18:23 | Review R015 | code Step 3: REVISE |
| 2026-05-20 18:26 | Review R016 | code Step 3: APPROVE |
| 2026-05-20 18:28 | Review R017 | plan Step 4: APPROVE |
| 2026-05-20 18:31 | Review R018 | code Step 4: APPROVE |
| 2026-05-20 18:32 | Review R019 | plan Step 5: APPROVE |
| 2026-05-20 18:35 | Review R020 | code Step 5: APPROVE |
