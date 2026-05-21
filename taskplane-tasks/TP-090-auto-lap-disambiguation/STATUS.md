# TP-090: Auto-lap disambiguation on `get_activity_intervals` — Status

**Current Step:** Step 6: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-05-20
**Review Level:** 2
**Review Counter:** 13
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

### Step 1: Model interval-source heuristics
**Status:** ✅ Complete

- [x] Inspect current interval payload fields to find any explicit source markers.
- [x] Record source-marker search scope and candidate key inventory in STATUS.md.
- [x] Define deterministic interval-source classifier contract with thresholds, precedence, unit assumptions, and helper/API placement.
- [x] Add Step 1 acceptance examples for structured, 1 km/1 mi auto-lap, unknown, and structured-repeat negative cases.
- [x] Define near-uniform distance/duration heuristic with unit-aware tolerances.
- [x] Document false-positive tradeoffs in STATUS.md.

---

### Step 2: Add additive meta to interval reads
**Status:** ✅ Complete

- [x] Implement the shared interval-source classifier helper with typed source constants.
- [x] Emit `_meta.interval_source` as `structured_workout`, `device_laps`, or `unknown`.
- [x] Emit `_meta.auto_lap_suspected: true` for near-uniform auto-lap patterns.
- [x] Do not remove or rename existing interval fields.
- [x] Fix generic-row gate so non-empty non-generic interval names/types/labels cannot become device laps by empty sibling fields.
- [x] Recognize explicit boolean `auto_lap` and `auto` lap-type/source markers as device laps.

---

### Step 3: Propagate to analyzers
**Status:** ✅ Complete

- [x] Document Step 3 analyzer propagation contract and no-public-tool scope in STATUS.md.
- [x] Add shared helper so analyzer source_tools/meta can propagate auto-lap suspicion.
- [x] Ensure interval-consuming analyzers decline per-interval execution quality claims when suspicion is true.
- [x] Add placeholder propagation tests if analyzer tools are not implemented yet.
- [x] Record TP-091/TP-093 downstream handoff notes for analyzer implementers.

---

### Step 4: Tests and docs
**Status:** ✅ Complete

- [x] Add fixtures for structured intervals, 1 km/1 mi auto-laps, and unknown source.
- [x] Update docs/reference and CHANGELOG.md.
- [x] Run full quality gate.

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
| R001 | Plan | 1 | REVISE | .reviews/R001-plan-step1.md |
| R002 | Plan | 1 | APPROVE | inline |
| R003 | Code | 1 | APPROVE | inline |
| R004 | Plan | 2 | APPROVE | .reviews/R004-plan-step2.md |
| R005 | Code | 2 | REVISE | .reviews/R005-code-step2.md |
| R006 | Code | 2 | APPROVE | inline |
| R007 | Plan | 3 | REVISE | .reviews/R007-plan-step3.md |
| R008 | Plan | 3 | APPROVE | .reviews/R008-plan-step3.md |
| R009 | Code | 3 | APPROVE | inline |
| R010 | Plan | 4 | APPROVE | inline |
| R011 | Code | 4 | APPROVE | inline |
| R012 | Plan | 5 | APPROVE | inline |
| R013 | Code | 5 | APPROVE | inline |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Analyzer public tools are still intentionally absent; TP-090 added shared metadata/policy helpers only for TP-091/TP-093 to consume later. | Documented handoff in Notes; no catalog/tool registration in this task. | Step 3 |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-20 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-20 17:20 | Task started | Runtime V2 lane-runner execution |
| 2026-05-20 17:20 | Step 0 started | Preflight |
| 2026-05-20 18:03 | Worker iter 1 | done in 2585s, tools: 208 |
| 2026-05-20 18:03 | Task complete | .DONE created |

---

## Blockers

*None*

---

## Notes

### Step 1 heuristic contract

- Field inventory inspected in `internal/intervals/activity_details.go` and current fixtures/tests: `IntervalsDTO` types `id`, `analyzed`, `icu_intervals`, `icu_groups` and preserves top-level `Raw`; `ActivityInterval` types `id`, `name`, `type`, `unit`, `start_index`, `end_index`, `start_time`, `end_time`, `start_distance`, `end_distance`, `distance`, `duration`, `average_power`, `average_hr`, `pace` and preserves interval `Raw`; `IntervalGroup` types `id`, `name`, `type`, `start_index`, `end_index` and preserves group `Raw`. Existing fixtures expose `label` as a raw interval key for extended metrics, but no typed explicit interval-source field.
- Source-marker search scope: classifier will inspect typed interval/group fields plus `IntervalsDTO.Raw`, `ActivityInterval.Raw`, and `IntervalGroup.Raw` for explicit source/workout-step markers before applying heuristics. Candidate keys checked in current code/fixtures: top-level activity `source` exists only for activity rows/Strava fallbacks, not normal interval DTOs; interval raw `label` may identify named work/rest segments; typed/raw `name`, `type`, and `icu_groups` are structured signals. No existing fixture exposes an explicit interval-origin key such as `interval_source`, `lap_source`, `source`, `auto_lap`, `manual_lap`, `workout_step`, or `workout_doc` inside normal interval rows/groups.
- Classifier contract: implement a shared `internal/analysis` helper with typed constants `IntervalSourceStructuredWorkout`, `IntervalSourceDeviceLaps`, and `IntervalSourceUnknown`, plus an `InferIntervalSource` API over small interval/group sample structs (names/types/labels/raw markers, indexes, start/end distance, distance, duration). Precedence is deterministic: (1) explicit structured/workout-step markers or strong structured signals classify `structured_workout`; (2) explicit device/lap source markers classify `device_laps` only when no structured signal is present; (3) near-uniform generic rows classify `device_laps`; (4) everything else is `unknown`. Strong structured signals include non-empty `icu_groups`, raw workout-step keys, or non-generic interval/group labels/types/names such as warmup/cooldown/work/rest/recovery/tempo/threshold/repeat/interval; generic labels such as `Lap`, numeric IDs, or empty names are not structured by themselves.
- Near-uniform contract: require at least 4 usable core rows after dropping at most 2 edge rows (first and/or last warmup/cooldown/partial rows only). Usable rows need positive `distance` or a positive `end_distance - start_distance`; duration, when used, must be positive. Rows with missing/zero/negative dimensions are ignored for the corresponding dimension and cannot be the sole basis for device-lap classification if fewer than 4 usable rows remain. Distance auto-lap targets are 1000 m and 1609.344 m; tolerance is max(25 m, 2.5%) for 1 km and max(40 m, 2.5%) for 1 mi. Duration auto-lap targets are 60, 300, 600, 900, 1800, and 3600 seconds with tolerance max(5 s, 2%). At least 80% of core usable rows must match the same target, and the core sequence must be monotonic/non-overlapping by distance ranges when distances exist, otherwise by indexes when indexes exist. Distance evidence wins over duration evidence; duration-only classification is allowed only for generic lap rows with no structured signal.
- Unit assumptions and metadata semantics: `distance`, `start_distance`, and `end_distance` are treated as meters because the response already exposes them as `*_m`; the `unit` enum is not used as distance-unit evidence because it may describe pace/targets. If meter assumptions are contradicted or cannot yield positive usable rows, classify `unknown`. Successful interval responses will always carry `_meta.interval_source` and a boolean `_meta.auto_lap_suspected`; unavailable/Strava-blocked responses keep their existing unavailable shape and do not get classifier metadata unless interval rows exist to evaluate. Diagnostic reasons stay internal/test-only.
- Acceptance examples for Step 4 fixtures/tests: (1) structured workout with `icu_groups` or named `Warmup`, `Work`, `Rest`, `Cooldown` intervals -> `_meta.interval_source="structured_workout"`, `_meta.auto_lap_suspected=false`; (2) six generic intervals named `Lap 1`...`Lap 6` with contiguous 1000 m distances within tolerance -> `device_laps`, `auto_lap_suspected=true`; (3) six generic intervals with contiguous ~1609.344 m distances within tolerance -> `device_laps`, `auto_lap_suspected=true`; (4) one to three rows, missing/zero distances/durations, or mixed non-target distances -> `unknown`, `auto_lap_suspected=false`; (5) structured-repeat negative case such as `6x1 km` with work/recovery names or groups, even though near-uniform over the work reps -> `structured_workout` or `unknown`, never `device_laps`, and analyzers must not make per-interval execution-quality claims from auto-lap-suspected rows.
- False-positive tradeoff: the heuristic is intentionally conservative. It may leave some real auto-laps as `unknown` when labels are non-generic or distance units are ambiguous, but it should not relabel structured workouts as device laps. Structured evidence takes precedence over uniformity, and only generic, contiguous, sufficiently numerous rows can become `device_laps`.

### Step 3 analyzer propagation contract

- Scope: TP-090 Step 3 will not implement or register `analyze_efforts_delta`, `compute_compliance_rate`, or any analyzer-family public tool/catalog/schema/docs entries. Those remain TP-091/TP-093 work. This step adds shared `internal/analysis` support and placeholder tests only.
- Metadata contract: future interval-consuming analyzers call a shared helper with `IntervalSourceResult` evidence. The helper appends/deduplicates `get_activity_intervals` in `_meta.source_tools`, emits optional `_meta.interval_source` only when source evidence is known, and emits optional `_meta.auto_lap_suspected` only when interval-source evidence was evaluated (pointer/omitempty semantics so non-interval analyzers do not gain misleading `false`).
- Execution-claim policy: future analyzers must call a shared policy helper before making per-interval structured-workout execution-quality claims. When `AutoLapSuspected` is true, the policy declines those claims with the reason `auto_lap_suspected`; when false for `structured_workout` or `unknown`, it does not decline by this rule. If exposed in analyzer text/meta later, TP-091/TP-093 should use that reason string verbatim.
- Test plan: add `internal/analysis` helper tests for source-tool dedupe/meta fields and policy decisions; add shaped analyzer placeholder tests proving supplied interval evidence propagates and existing non-interval analyzer goldens do not grow interval fields.
- Downstream handoff for TP-091/TP-093: call `analysis.ApplyIntervalSourceEvidence` when an analyzer consumes `get_activity_intervals`; call `analysis.IntervalExecutionClaimPolicy` before saying an athlete hit/missed a structured interval target. If the decision declines, avoid per-interval execution-quality claims and use reason `analysis.IntervalExecutionDeclineAutoLapSuspected` / `auto_lap_suspected`.
- Documentation review: README setup/catalog content did not need a change for response-only metadata; `web/content/reference/tools.md` was affected and updated; PRD behavior remains aligned with v0.6 analyzer scope and did not need a product-scope change.

| 2026-05-20 17:24 | Review R001 | plan Step 1: REVISE |
| 2026-05-20 17:29 | Review R002 | plan Step 1: APPROVE |
| 2026-05-20 17:30 | Review R003 | code Step 1: APPROVE |
| 2026-05-20 17:32 | Review R004 | plan Step 2: APPROVE |
| 2026-05-20 17:38 | Review R005 | code Step 2: UNKNOWN |
| 2026-05-20 17:41 | Review R006 | code Step 2: APPROVE |
| 2026-05-20 17:43 | Review R007 | plan Step 3: REVISE |
| 2026-05-20 17:45 | Review R008 | plan Step 3: APPROVE |
| 2026-05-20 17:49 | Review R009 | code Step 3: APPROVE |
| 2026-05-20 17:51 | Review R010 | plan Step 4: APPROVE |
| 2026-05-20 17:57 | Review R011 | code Step 4: APPROVE |
| 2026-05-20 17:58 | Review R012 | plan Step 5: APPROVE |
| 2026-05-20 18:01 | Review R013 | code Step 5: APPROVE |
