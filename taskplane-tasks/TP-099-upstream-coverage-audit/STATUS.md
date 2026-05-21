# TP-099: Upstream coverage audit for zone-time/load-balance analyzers — Status

**Current Step:** Step 6: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-05-20
**Review Level:** 1
**Review Counter:** 10
**Iteration:** 1
**Size:** S

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no protected docs are changed without explicit approval

---

### Step 1: Define measurement method
**Status:** ✅ Complete

- [x] Identify exact eligible fixture corpus, exclusions, and fields that count as precomputed zone times.
- [x] Define output metrics with denominators, family grouping, and fallback/missing-precomputed semantics.
- [x] Choose threshold or mark threshold as operator decision if not already agreed.
- [x] Clarify exclusion precedence so excluded paths cannot re-enter through broad shape detection.
- [x] Define deterministic metric-family applicability rules for precomputed/fallback/unknown counts.

---

### Step 2: Implement/run audit
**Status:** ✅ Complete

- [x] Add a small script or Go test/helper to scan fixtures and report coverage.
- [x] Run it against the v0.2 fixture set.
- [x] Record results in STATUS.md and a durable doc if user-facing.

---

### Step 3: Document gap or close it
**Status:** ✅ Complete

- [x] If fallback exceeds threshold/looks risky, create `docs/upstream-gaps/zone-time-coverage.md` with evidence and feature-request text.
- [x] If coverage is sufficient, document the result in `docs/kr5-benchmark.md` or STATUS.md.
- [x] Do not change analyzer behavior except to fix discovered coverage bugs.

---

### Step 4: Verify
**Status:** ✅ Complete

- [x] Re-run `go run scripts/audit_zone_time_coverage.go` and compare summary with STATUS/docs.
- [x] Run quality gate commands (`make test`, `make build`, `make lint`) or document unrelated/pre-existing failures.
- [x] Update CHANGELOG.md for the new user-facing upstream-gap documentation.

---


### Step 5: Testing & Verification
**Status:** ✅ Complete

- [x] Targeted verification passing: `go run scripts/audit_zone_time_coverage.go` matches STATUS/docs summary.
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
| R001 | plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |
| R002 | plan | 1 | REVISE | `.reviews/R002-plan-step1.md` |
| R003 | plan | 1 | APPROVE | inline |
| R004 | plan | 2 | APPROVE | inline |
| R005 | plan | 3 | UNAVAILABLE | inline |
| R006 | plan | 4 | REVISE | `.reviews/R006-plan-step4.md` |
| R007 | plan | 4 | APPROVE | inline |
| R008 | plan | 5 | REVISE | `.reviews/R008-plan-step5.md` |
| R009 | plan | 5 | REVISE | `.reviews/R009-plan-step5.md` |
| R010 | plan | 5 | APPROVE | inline |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| v0.2 fixture roots contain 6 eligible activity-like objects but no valid precomputed zone-time arrays for power, heart-rate, or pace. | Documented as upstream/fixture coverage risk with reproducible audit script. | `scripts/audit_zone_time_coverage.go`, `docs/upstream-gaps/zone-time-coverage.md` |
| Current `compute_zone_time` / `compute_load_balance` behavior is precomputed-only; "fallback" in this audit means missing-precomputed/stream-math would be required, not an implemented stream fallback. | Preserved analyzer behavior and documented semantics. | `docs/upstream-gaps/zone-time-coverage.md` |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-20 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-20 22:23 | Task started | Runtime V2 lane-runner execution |
| 2026-05-20 22:23 | Step 0 started | Preflight |
| 2026-05-20 22:50 | Worker iter 1 | done in 1604s, tools: 158 |
| 2026-05-20 22:50 | Task complete | .DONE created |

---

## Blockers

*None*

---

## Notes

R002 suggestions: keep positive-total validation; phrase threshold policy as operator decision plus non-zero missing-precomputed opportunities as risky evidence, not automatic pass/fail.

### Step 1 measurement method

**Fixture corpus and exclusions:** scan only `internal/intervals/testdata/**/*.json` and `internal/tools/testdata/**/*.json` as the v0.2 fixture roots named by the task. Apply path/type exclusions before object-shape rules: exclude wellness, events, gear, workout library, custom items, activity messages, activity interval fixtures, analyzer golden responses, schema snapshots, and non-JSON files; excluded paths never re-enter through broad `id`/date matching. Treat an eligible fixture unit as each JSON object (root object or element of a root array) from a non-excluded path that matches one of these type-detection rules: activity/detail-like (`id` plus `start_date` or `start_date_local`, excluding event-only markers such as `category`, `workout_doc`, or `show_as_note`), training-summary-like (`date` plus `timeInZones` or `timeInZonesTot`), or extended-metrics-like (`id` plus any precomputed zone-time key below).

**Precomputed zone-time fields:** count a field only when it is a non-empty numeric array with positive total seconds. Training summary power coverage is `timeInZones` with `timeInZonesTot > 0`. Activity/extended fields by metric family are power: `icu_zone_times`, `power_zone_distribution_seconds`, `power_zone_times`; pace: `gap_zone_times`, `pace_zone_times`, `pace_zone_time_seconds`; heart rate: `hr_zone_times`, `heartrate_zone_times`, `heart_rate_zone_times`, `hr_time_in_zones`.

**Metrics and denominators:** report by source path/type and by zone metric family (`power`, `heart_rate`, `pace`) plus totals. `fixture_count` is the number of eligible fixture units. A `(fixture_unit, metric_family)` opportunity is `precomputed_count` when a valid precomputed zone array is present for that family, `fallback_count` when the unit is eligible for that family but lacks valid precomputed zones and would require stream math/missing-precomputed handling, and `unknown_count` when the unit shape is eligible but lacks enough signal to decide applicability for that family. A metric family is applicable when it has a valid precomputed array, or when family-specific signals exist without a valid array: power signals are `icu_training_load`, `power_load`, `average_watts`, `weighted_average_watts`, `max_watts`, or stream types containing `watts`/`power`; heart-rate signals are `hr_load`, `average_heartrate`, `max_heartrate`, or stream types containing `heartrate`/`heart_rate`/`hr`; pace signals are `pace_load`, `average_speed`, `max_speed`, `distance`/`icu_distance` with moving or elapsed time, or stream types containing `velocity_smooth`, `pace`, or `speed`. Training-summary rows with valid `timeInZones` apply only to power because current `collectZoneAggregate` uses `get_training_summary` only for unsport-filtered power. If an activity-like unit has no family-specific signal, count one `unknown_count` per family rather than fallback. Because current `compute_zone_time` and `compute_load_balance` are precomputed-only, `fallback_count` means “would require stream math or returns missing/partial,” not an actual current stream fallback.

**Threshold:** no agreed fallback threshold was found in `ROADMAP.md` or `docs/prd/PRD-icuvisor.md`; threshold is an operator decision. Step 3 will document an upstream gap if measured fallback coverage is non-zero and therefore risky for the current fixture corpus, without silently changing analyzer behavior.

### Step 2 audit results

Command: `go run scripts/audit_zone_time_coverage.go`.

Result against the v0.2 fixture roots (`internal/intervals/testdata/**/*.json`, `internal/tools/testdata/**/*.json`): 6 eligible activity-like fixture objects, 36 skipped excluded/non-eligible objects or files. No valid precomputed zone-time arrays were present. No deterministic family-specific metric signals were present either, so fallback/missing-precomputed opportunities were not asserted from sparse fixtures; all 18 `(fixture, metric_family)` opportunities are `unknown`.

| metric_family | precomputed_count | fallback_count | unknown_count | coverage |
|---|---:|---:|---:|---:|
| power | 0 | 0 | 6 | 0.0% |
| heart_rate | 0 | 0 | 6 | 0.0% |
| pace | 0 | 0 | 6 | 0.0% |

Durable user-facing documentation is deferred to Step 3 so the result can be classified as an upstream gap or sufficient coverage in the required destination.

### Step 3 documentation decision

Coverage was not sufficient to close the audit in `docs/kr5-benchmark.md`: the fixture set contains 0 precomputed zone arrays and 18 unknown metric-family opportunities. The risky/inconclusive result is documented in `docs/upstream-gaps/zone-time-coverage.md`; `docs/kr5-benchmark.md` is intentionally unchanged.

### Step 4 verification results

- `go run scripts/audit_zone_time_coverage.go` passed and matched STATUS/docs: fixture_count 6, skipped 36, and each metric family has precomputed 0 / fallback 0 / unknown 6.
- `make test` passed.
- `make build` passed.
- `make lint` passed with 0 issues.

### Step 5 final verification plan

Step 5 will rerun the final gate commands rather than carrying Step 4 results forward: `go run scripts/audit_zone_time_coverage.go`, `make test`, `make build`, and `make lint`.

### Step 5 final verification results

- `go run scripts/audit_zone_time_coverage.go` passed and matched STATUS/docs: fixture_count 6, skipped 36, and each metric family has precomputed 0 / fallback 0 / unknown 6.
- `make test` passed.
- `make build` passed.
- `make lint` passed with 0 issues.

### Step 6 delivery notes

- Must-update docs: `CHANGELOG.md` and `STATUS.md` updated; user-facing gap doc added at `docs/upstream-gaps/zone-time-coverage.md`.
- Check-if-affected docs: `README.md` not affected because setup/tool behavior did not change; `web/content/reference/tools.md` not affected because the tool catalog did not change; `docs/prd/PRD-icuvisor.md` not affected because analyzer behavior did not intentionally diverge from product scope.

| 2026-05-20 22:26 | Review R001 | plan Step 1: REVISE |
| 2026-05-20 22:30 | Review R002 | plan Step 1: REVISE |
| 2026-05-20 22:32 | Review R003 | plan Step 1: APPROVE |
| 2026-05-20 22:34 | Review R004 | plan Step 2: APPROVE |
| 2026-05-20 22:40 | Review R006 | plan Step 4: REVISE |
| 2026-05-20 22:41 | Review R007 | plan Step 4: APPROVE |
| 2026-05-20 22:44 | Review R008 | plan Step 5: REVISE |
| 2026-05-20 22:45 | Review R009 | plan Step 5: REVISE |
| 2026-05-20 22:46 | Review R010 | plan Step 5: APPROVE |
