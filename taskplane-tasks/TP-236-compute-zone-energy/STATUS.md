# TP-236: Add deterministic power-zone energy analysis — Status

**Current Step:** Step 4: Formula resource, generated data, and docs
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 2
**Review Counter:** 12
**Iteration:** 4
**Size:** L

---

### Step 0: Preflight

**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Stream, time-base, and zone semantics reviewed

---

### Step 1: Lock the numerical and response contract

**Status:** ✅ Complete

- [x] Integration and boundary rules defined
- [x] Range limits and statuses defined
- [x] Terse/full response contract defined
- [x] Mechanical versus metabolic boundary explicit
- [x] R001 pure input, diagnostics, invalid-sample, and rounding contract locked
- [x] R001 configured-zone validation, selection, and mixed-configuration policy locked
- [x] R001 deterministic range cap and exact status state machine locked
- [x] R001 concrete JSON/meta contract and coherent formula/test checkpoints locked
- [x] R002 operational-versus-absence stream error classification locked
- [x] R002 tool-specific metadata emission path and exact metadata shapes locked
- [x] R002 deterministic mixed-zone identity, ordering, shares, counts, and audit enums locked
- [x] R002 pure validation, mismatch, and short-input result semantics locked
- [x] R003 serialized analysis-units and shared analyzer-meta semantics locked
- [x] R003 insufficient-reason precedence and zone-config absence semantics locked
- [x] R003 literal interpretation and ordered boundary text locked

---

### Step 2: Implement and test the pure calculation

**Status:** ✅ Complete

- [x] Timestamp-weighted zone integration implemented
- [x] Invalid and missing samples reported explicitly
- [x] Stable totals, shares, and rounding implemented
- [x] Numerical boundary tests added

---

### Step 3: Add and register the MCP analyzer

**Status:** ✅ Complete

- [x] Range, profile, and stream orchestration implemented
- [x] Partial activity coverage reported
- [x] Full-toolset registration and schemas added
- [x] Catalog, safety, protocol, and tier coverage added
- [x] R008 catalog source-of-truth and analyzer-family activation surfaces updated
- [x] R008 MCP read-only annotation mapped and asserted through tools/list
- [x] R008 rounded cross-activity aggregation and response-wide share reconciliation tested

---

### Step 4: Formula resource, generated data, and docs

**Status:** 🟨 In Progress

- [x] Formula resource updated
- [x] PRD and roadmap aligned
- [x] Generated website data and cookbook updated
- [x] Changelog updated
- [x] R012 below-first-boundary and open-final-zone formula semantics documented

---

### Step 5: Testing & Verification

**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Race suite passing
- [ ] Lint passing
- [ ] All failures fixed
- [ ] Build passes
- [ ] Generated docs clean

---

### Step 6: Documentation & Delivery

**Status:** ⬜ Not Started

- [ ] Must Update docs modified
- [ ] Check If Affected docs reviewed
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |
| R002 | Plan | 1 | REVISE | `.reviews/R002-plan-step1.md` |
| R003 | Plan | 1 | REVISE | `.reviews/R003-plan-step1.md` |
| R004 | Plan | 1 | APPROVE | `.reviews/R004-plan-step1.md` |
| R005 | Code | 1 | APPROVE | `.reviews/R005-code-step1.md` |

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Tier-2 `taskplane-tasks/CONTEXT.md` is absent; required implementation paths and authoritative PRD/roadmap context are available. | Proceed using task packet and repository guidance. | Preflight |

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-07-10 | Task staged | PROMPT.md and STATUS.md created |
| 2026-07-10 16:11 | Task started | Runtime V2 lane-runner execution |
| 2026-07-10 16:11 | Step 0 started | Preflight |
| 2026-07-10 16:30 | Worker iter 1 | done in 1177s, tools: 86 |
| 2026-07-10 16:53 | Worker iter 2 | done in 1334s, tools: 16 |
| 2026-07-10 16:53 | Step 2 started | Implement and test the pure calculation |
| 2026-07-10 17:08 | Worker iter 3 | done in 951s, tools: 59 |
| 2026-07-10 17:08 | Step 3 started | Add and register the MCP analyzer |

## Blockers

*None*

## Notes

### Step 1 revised contract plan (R001)

- **Pure integration contract:** `ZoneEnergyInput` carries elapsed `TimestampsSeconds`, `PowerWatts`, and one validated `PowerZoneConfig`; `ComputeZoneEnergy` applies the left-endpoint rule `work_j += power[i] * (time[i+1]-time[i])`. The final sample contributes nothing without a following timestamp. A length mismatch rejects the whole activity (no aligned-prefix guess) while reporting `misaligned_samples = abs(len(power)-len(time))`. Intervals are evaluated in deterministic precedence: non-finite endpoint timestamp; duplicate timestamp; reversed timestamp; gap over 60 seconds; non-finite left power; negative left power. Invalid timestamps therefore invalidate each adjacent interval that uses them; invalid power invalidates only its own left interval. Zero watts is valid coasting time and is zoned, contributing zero joules. Diagnostics are exactly `input_samples`, `aligned_samples`, `usable_intervals`, `skipped_intervals`, `misaligned_samples`, `skipped_non_finite_timestamp`, `skipped_duplicate_timestamp`, `skipped_reversed_timestamp`, `skipped_large_gap`, `skipped_non_finite_power`, and `skipped_negative_power`.
- **Numerical presentation:** accumulate unrounded float64 seconds and joules, convert with `1 kJ = 1000 W*s`, then emit seconds/kJ to 3 decimals and shares to 4 decimals. Headline totals are sums of displayed zone totals; shares use displayed totals and deterministically adjust the last nonzero zone so nonzero shares sum to exactly `1.0000`. All-zero work emits `energy_share: 0` for every zone while retaining valid coasting seconds/time shares.
- **Configured zones:** boundaries are ordered lower bounds in watts and are never sorted. The first may be zero; all others must be finite, strictly increasing, and positive. Duplicate, descending, negative, or non-finite boundaries reject the configuration. Names pair by boundary index; missing/blank names become `Zone N`, and extras are ignored. Inclusion is `lower <= watts < upper`, final zone open-ended. If the first boundary is above zero, an explicit ordinal-0 `Below <first zone>` bucket `[0, first)` captures coasting/sub-boundary power. Sport settings are selected in activity `Type`, then `SubType`, order, matching either `SportSettings.Type` or any `Types` value case-insensitively. Distinct setting IDs/boundary/name vectors remain distinct configuration groups; rows are never silently merged by ordinal across configurations.
- **Bounded orchestration/state machine:** accept an inclusive athlete-local range of 1..366 days. Request 201 activity candidates, sort by local start (fallback UTC start) then activity ID, retain 200, and mark truncation only when candidate 201 exists. Optional sport filter is exact case-insensitive activity type/subtype. Operational profile/list/cancellation errors return tool errors. Per-activity stream fetch errors are data-coverage skips. Status is exactly `ok` (one or more usable activities, no skip/invalid/truncation), `partial` (usable data plus any skipped activity, skipped invalid interval, incompatible/missing config, stream-fetch failure, or truncation), or `insufficient` (no usable activity): reason `no_activities`, `missing_power_zones`, or `no_usable_power_streams`. Zero-power-only usable data is `ok` unless another partial condition applies.
- **Concrete response:** terse `result` always has `status`, optional `insufficient_reason`, `start_date`, `end_date`, optional `sport`, `activity_count`, `usable_activity_count`, `skipped_activity_count`, `invalid_interval_count`, `truncated_activity_candidates`, `total_seconds`, `total_kj`, `zones`, and `interpretation`. Each zone row has `zone_key`, `sport_setting_id`, `sport`, `zone`, `name`, `lower_watts`, optional `upper_watts`, `seconds`, `kj`, `time_share`, and `energy_share`. `interpretation` is fixed text stating this is power-derived mechanical work, not metabolic energy, calorie expenditure, or food calories. With `include_full`, `series[]` adds activity audit rows with `activity_id`, local `date`, `sport`, `status`, optional `reason`, `sport_setting_id`, totals, usable/skipped interval counts, and diagnostics; IDs/reasons are omitted from terse output. Raw samples never appear.
- **Pinned metadata:** `_meta.method = "left_endpoint_power_timestamp_integration"`; `_meta.source_tools = ["get_activities","get_activity_streams","get_athlete_profile"]`; `_meta.formula_ref = "icuvisor://analysis-formulas#power_zone_mechanical_work"`; `_meta.units = {"power":"W","time":"s","integration_work":"J","work":"kJ"}`; `_meta.assumptions` includes 60-second maximum interval, left-endpoint/final-sample rules, activity cap, and disclosed configured-zone sources; `_meta.boundaries` repeats the mechanical/metabolic distinction and no interpolation/raw-sample policy.
- **Step coherence:** Step 1 adds contract constants/types plus a real `TestZoneEnergyContract` compile/definition test and runs the filtered checkpoint. The formula ref constant, rendered paragraph, BIPM SI citation, golden resource, and resource tests move together in Step 4, avoiding intentional golden drift. The PRD update will replace the stale claim that only `compute_activity_segment_stats` reads streams.

### Step 1 second revision (R002)

- **Stream errors:** context cancellation/deadline and every non-absence client error abort the analyzer with the short tool error. This includes `intervals.ErrUnauthorized`, `ErrRateLimited`, `ErrUpstream`, unknown transport errors, and malformed-response/decode errors; iteration stops immediately. Only `intervals.ErrNotFound` becomes `streams_not_found`, an advertised `StreamTypes` set lacking `watts` or `time` becomes `required_streams_not_advertised`, and a successful response lacking usable canonical arrays becomes a closed data-coverage skip. Step 3 tests will cover unauthorized, rate-limit, upstream, unknown/decode, cancellation, not-found, unadvertised, and empty/misaligned response classes.
- **Metadata implementation:** `compute_zone_energy` will use a tool-specific payload/encoder, following `get_activity_histogram`, instead of `encodeAnalyzerResponse`. `zoneEnergyMeta` embeds normalized `analysis.AnalyzerMeta` and adds exact `units`, `zone_sources`, and `coverage` fields. `units` is `map[string]string`. `zone_sources[]` is `{zone_key:string,sport_setting_id:int,sport:string,boundaries_watts:[]float64,names:[]string}`. `coverage` is `{fetched_candidate_count:int,retained_candidate_count:int,sport_matched_activity_count:int,usable_activity_count:int,skipped_activity_count:int}`. Assumption keys/types are exactly `integration_rule:string("left_endpoint")`, `timestamp_unit:string("s")`, `final_sample_duration_seconds:int(0)`, `max_interval_seconds:int(60)`, `activity_cap:int(200)`, `candidate_fetch_limit:int(201)`, and `interpolation:bool(false)`.
- **Zone identity/order/shares:** canonical setting sport is trimmed `SportSettings.Type`, else the first trimmed `Types` entry that matched the activity candidate. A configuration fingerprint is the first 12 lowercase hex characters of SHA-256 over canonical JSON `{sport,boundaries_watts,names}`. Positive IDs use `zone_key = "setting-<id>-<fingerprint>"`; zero IDs use `zone_key = "config-<fingerprint>"`. Identity is `(positive ID, fingerprint)` or fingerprint alone for ID zero, so duplicate IDs with different definitions remain separate and duplicate identical definitions coalesce. Groups sort by case-folded canonical sport, numeric setting ID, then `zone_key`; rows sort by zone ordinal. Row `sport` is canonical setting sport, never an arbitrary activity type. Time/energy-share denominators span the whole response; independent rounding remainders go to the last output-ordered row with positive displayed seconds/kJ respectively.
- **Coverage/audit:** fetch at most 201 range candidates, sort all fetched candidates, retain the first 200 before sport filtering, and set truncation only for candidate 201. `activity_count` means sport-matched activities among the retained set and exactly equals usable plus skipped; metadata separately exposes all fetched/retained/matched counts. Per-activity status is closed to `usable`, `partial`, or `skipped`. `partial` has reason `invalid_intervals_skipped`. `skipped` reason is one of `no_matching_power_zone_config`, `invalid_power_zone_config`, `required_streams_not_advertised`, `streams_not_found`, `missing_power_stream`, `missing_time_stream`, `misaligned_streams`, `insufficient_stream_samples`, or `no_usable_intervals`.
- **Pure rejection/result:** `ComputeZoneEnergy(input) (ZoneEnergyResult, error)` validates `PowerZoneConfig` internally and returns `ErrInvalidPowerZoneConfig` for invalid boundaries; callers do not pre-certify it. Mismatch and short streams are data results, not errors. `input_samples = max(len(power),len(timestamps))`, `aligned_samples = min(...)`. On mismatch, `misaligned_samples = abs(...)`, `usable_intervals = 0`, `skipped_intervals = max(input_samples-1,0)`, and all per-invalid-reason counters remain zero. Equal-length input shorter than two has zero usable/skipped intervals. For aligned length N>=2, `usable_intervals + skipped_intervals = N-1`. `TestZoneEnergyContract` will assert config validation, mismatch, short input, diagnostic names, and these counter equations rather than merely compiling.

### Step 3 plan revision (R008)

- Register `newComputeZoneEnergyTool(...)` through `registryBaseTools` in `internal/tools/catalog.go`, classify it in `toolCatalogGroup`, and update `internal/tools/catalog_test.go` analyzer activation coverage so generated metadata and advanced capability discovery remain complete.
- Map `RequirementRead` to the SDK tool annotation in `internal/mcp/registrar_tools.go` and assert `annotations.readOnlyHint: true` through the `tools/list` protocol response.
- Aggregate the pure calculator's displayed 3-decimal per-activity zone seconds/kJ so aggregate rows, headline totals, and full audit rows reconcile exactly. Recompute response-wide shares from displayed aggregate totals with deterministic remainder assignment to the last positive output-ordered row, and test multiple activities/configuration groups.

### Step 1 third revision (R003)

- **Serialized units path:** use `_meta.analysis_units` (not response-owned `_meta.units`) with exact value `{power:"W",time:"s",integration_work:"J",work:"kJ"}` so `response.Shape` preserves it while continuing to add ordinary catalog/common metadata. The Step 3 serialized-response and schema tests assert `analysis_units` after shaping. No response-shaping package change is needed.
- **Shared analyzer semantics:** `zoneEnergyMeta` embeds `analysis.AnalyzerMeta`. Its `n` is usable activity count (both `usable` and `partial` activity audit rows), `missing_days` is always `0` because rest days are not missing observations and coverage is activity-count based, `missing_action` is `"skip"`, and `insufficient_sample` is computed with `MinSamples: 1`, exactly matching whether usable activity count is zero and the result status is `insufficient`.
- **Insufficient decision table:** `no_activities` means zero retained sport-matched candidates after optional sport filtering. A zero-length boundary list is absence and yields activity reason `no_matching_power_zone_config`, not invalidity. Nonempty malformed boundaries yield `invalid_power_zone_config`. When no activity is usable: all `no_matching_power_zone_config` => `missing_power_zones`; all reasons zone-related with at least one invalid configuration => `invalid_power_zones`; any activity that reached stream eligibility but had unavailable/misaligned/invalid stream data, including a mixture with zone-related skips => `no_usable_power_streams`. Thus precedence is `no_activities`, then all-missing zones, then all-zone-related with invalid, else no usable streams.
- **Literal model-visible text:** `result.interpretation` is exactly `"Power-derived kJ is external mechanical work only; it is not metabolic energy, calorie expenditure, or food calories."` The ordered `_meta.boundaries` array is exactly: (1) `"Mechanical work from recorded power is not metabolic energy, calorie expenditure, or food calories."`; (2) `"Left-endpoint integration; the final sample contributes no duration or work."`; (3) `"Intervals longer than 60 seconds and invalid samples are skipped; missing power is not interpolated."`; (4) `"Raw stream samples are never returned."`.
| 2026-07-10 16:20 | Review R002 | plan Step 1: REVISE |
| 2026-07-10 16:24 | Review R003 | plan Step 1: REVISE |
| 2026-07-10 16:27 | Review R004 | plan Step 1: APPROVE |
| 2026-07-10 16:35 | Review R005 | code Step 1: APPROVE |
| 2026-07-10 16:38 | Review R006 | plan Step 2: APPROVE |
| 2026-07-10 17:01 | Review R008 | plan Step 3: REVISE |
| 2026-07-10 17:04 | Review R009 | plan Step 3: APPROVE |
| 2026-07-10 17:21 | Review R010 | code Step 3: APPROVE |
| 2026-07-10 17:28 | Review R012 | code Step 4: REVISE |
