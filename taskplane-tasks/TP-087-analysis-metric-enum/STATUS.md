# TP-087: `analysis_metric` closed enum and unknown-metric hints — Status

**Current Step:** Step 6: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-05-20
**Review Level:** 2
**Review Counter:** 19
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

### Step 1: Design the enum and aliases
**Status:** ✅ Complete

- [x] List supported metric identifiers from existing read-tool fields.
- [x] Decide canonical names and any safe aliases; reject derived/free-form expressions.
- [x] Document hint strategy for unknown metrics.
- [x] Record concrete first-pass enum inventory with source read surfaces and excluded effort/derived metrics.
- [x] Define snake_case canonical naming, unit suffix, abbreviation, and collision rules.
- [x] List conservative safe aliases and intentionally rejected near-misses.
- [x] Specify free-form arithmetic/expression rejection examples.
- [x] Decide reusable package/API boundary and schema/docs implications.
- [x] Resolve scalar-vs-structured metrics by excluding zone arrays from scalar enum or documenting kind metadata.
- [x] Align intentional canonical-name divergences (`if`, `vi`, `compliance_pct`, `ramp`) with source-field aliases.
- [x] Correct source attribution for activity row versus interval row metrics.
- [x] Record PRD-example exclusions (`pace_at_lt2`, `power_at_lt2`, `np`) as v0.6 staging follow-ups.
- [x] Clarify source-family metadata helper in `internal/analysis` API.
- [x] Resolve duplicate canonical metrics by allowing multiple source descriptors and analyzer-specific selection.
- [x] Reclassify `weekly_tss` and `weekly_hours` as analyzer-level aggregate metrics with source/aggregation metadata.
- [x] Specify metadata fields for unit label, row grain, metric kind, and optional scale reference.
- [x] Clarify interval `pace` naming as unit-contextual or defer it.
- [x] Remove or stage boolean wellness flags (`temp_weight`, `temp_resting_hr`) from scalar enum.
- [x] Complete duplicated-source targets for `ctl`, `atl`, and `feel`.
- [x] Split `get_extended_metrics` inventory into activity and interval source families.

---

### Step 2: Implement validation helpers
**Status:** ✅ Complete

- [x] Add shared enum parsing/JSON schema helpers under `internal/analysis` or a small tools helper.
- [x] Return short actionable invalid-argument errors with hints.
- [x] Keep validation reusable by all analyzer tools.
- [x] Define `Metric`, metadata/source types, canonical values, and aliases in `internal/analysis`.
- [x] Implement `ParseMetric`, concise invalid errors, expression detection, and hint categorization.
- [x] Implement schema helper(s) that enumerate canonical values and expose metric metadata without MCP coupling.
- [x] Fix `ParseMetric` so canonical values containing `_per_` parse before expression rejection.
- [x] Remove the remaining pre-alias expression guard so `_per_` canonical metrics round-trip.

---

### Step 3: Tests
**Status:** ✅ Complete

- [x] Add table-driven parsing/schema tests for valid metrics, aliases, unknown names, and arithmetic expressions.
- [x] Add round-trip coverage proving every `MetricValues()` schema enum entry parses successfully.
- [x] Assert error text is concise and does not expose internals.
- [x] Run targeted tests.
- [x] Add explicit unknown-metric hint cases for efforts, zones, segment stats, compliance, and generic unsupported names.
- [x] Add schema contract tests for canonical enum-only values, concise description, aliases, and expression rejection.
- [x] Add metadata helper tests for catalog validation, source descriptors, multi-source, derived, scale, and defensive copies.
- [x] Add error type/predicate contract tests.
- [x] Use `go test ./internal/analysis` as the Step 3 targeted command.

---

### Step 4: Docs and verification
**Status:** ✅ Complete

- [x] Update analyzer docs/reference stubs if present.
- [x] Run full quality gate.
- [x] Update CHANGELOG.md.
- [x] Inspect README.md, web/content/reference/tools.md, web/data/tools.json, docs/prd/PRD-icuvisor.md, and ROADMAP.md for analyzer-doc impact and record no generated docs churn if applicable.
- [x] Add `[Unreleased]` CHANGELOG entry for closed `analysis_metric` helpers and unknown-metric hints.
- [x] Run and record `go test ./internal/analysis`, `make test`, `make build`, and `make lint`, fixing or documenting failures.
- [x] Record Step 4 verification evidence and Step 5 handoff policy in STATUS.md: Step 5 may reuse Step 4 results only if no files changed after those commands; otherwise re-run affected gates, with full re-run preferred for final confirmation.

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
| R002 | Plan | 1 | REVISE | .reviews/R002-plan-step1.md |
| R003 | Plan | 1 | REVISE | .reviews/R003-plan-step1.md |
| R004 | Plan | 1 | REVISE | .reviews/R004-plan-step1.md |
| R005 | Plan | 1 | APPROVE | inline reviewer |
| R006 | Code | 1 | APPROVE | inline reviewer |
| R007 | Plan | 2 | APPROVE | .reviews/R007-plan-step2.md |
| R008 | Code | 2 | REVISE | .reviews/R008-code-step2.md |
| R009 | Code | 2 | REVISE | .reviews/R009-code-step2.md |
| R010 | Code | 2 | APPROVE | inline reviewer |
| R011 | Plan | 3 | REVISE | .reviews/R011-plan-step3.md |
| R012 | Plan | 3 | APPROVE | .reviews/R012-plan-step3.md |
| R013 | Code | 3 | APPROVE | inline reviewer |
| R014 | Plan | 4 | REVISE | .reviews/R014-plan-step4.md |
| R015 | Plan | 4 | REVISE | .reviews/R015-plan-step4.md |
| R016 | Plan | 4 | APPROVE | .reviews/R016-plan-step4.md |
| R017 | Code | 4 | APPROVE | inline reviewer |
| R018 | Plan | 5 | APPROVE | .reviews/R018-plan-step5.md |
| R019 | Code | 5 | APPROVE | inline reviewer |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| No out-of-scope discoveries; analyzer docs/reference surfaces need no generated update until analyzer tools are registered. | Recorded for delivery. | Step 4/6 notes |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-20 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-20 13:34 | Task started | Runtime V2 lane-runner execution |
| 2026-05-20 13:34 | Step 0 started | Preflight |
| 2026-05-20 14:54 | Worker iter 1 | done in 4820s, tools: 260 |
| 2026-05-20 14:54 | Task complete | .DONE created |

---

## Blockers

*None*

---

## Notes

- Code review R008 found canonical `pace_seconds_per_km` / `pace_seconds_per_mile` were emitted by schema but rejected as expressions because `_per_` was checked before alias lookup; fix parser order and cover round-trip in Step 3.
- Code review R009 found the original pre-alias expression guard remained in `ParseMetric`; remove it before re-review.
- Plan review R011 requires Step 3 to cover deterministic hint categories, expression examples including `_per_` regression, schema contract, metadata helpers, invalid error predicate, and targeted command `go test ./internal/analysis`.
- Plan review R014 requires concrete docs surfaces, CHANGELOG plan, exact verification commands, STATUS evidence, and Step 4/Step 5 handoff policy.
- Plan review R015 requires the Step 4/Step 5 handoff policy to state whether Step 5 re-runs or can reuse Step 4 gate results.
- Step 4 docs inspection: README.md, web/content/reference/tools.md, web/data/tools.json, docs/prd/PRD-icuvisor.md, and ROADMAP.md exist. README/tool reference/generated tools data contain no registered analyzer or `analysis_metric` stubs to update; PRD/ROADMAP already document v0.6 analyzer scope, so no generated docs churn is required for this helper-only task.
- Step 4 verification: `go test ./internal/analysis` passed; first `make lint` flagged `unparam` in new helpers, fixed by simplifying helper signatures; final `make test && make build && make lint` passed with 0 lint issues. Step 5 may reuse these results only if no files change afterward, but final confirmation should re-run gates when practical.
- Step 5 verification reran `go test ./internal/analysis`, `make test`, `make build`, and `make lint`; all passed, with no pre-existing unrelated failures to document.
- Documentation delivery: Must Update docs were modified (`CHANGELOG.md` and `STATUS.md`). Check If Affected docs reviewed: README.md, web/content/reference/tools.md, web/data/tools.json, docs/prd/PRD-icuvisor.md, and ROADMAP.md; no changes needed beyond the changelog because analyzer tools are not registered yet.
- Plan review R001 requires STATUS.md to record the concrete metric inventory, canonical/alias policy, expression rejection examples, hint categories, reusable API boundary, and schema/docs implications before Step 2.
- Plan review R002 requires resolving scalar/structured metric contradictions, canonical-name source alias divergences, activity-vs-interval source attribution, PRD-example staging exclusions, and source-family metadata API shape before Step 2.
- Plan review R003 requires making duplicate-source metrics unambiguous, classifying weekly aggregates as analyzer-level derived metrics, adding unit/grain/kind/scale metadata targets, and clarifying interval pace semantics before Step 2.
- Plan review R004 requires staging boolean wellness temp flags, completing duplicate source descriptors for `ctl`/`atl`/`feel`, and splitting extended activity versus interval metric inventory.

### Step 1 design record

Supported first-pass `analysis_metric` inventory is limited to numeric/scalar fields exposed by current read tools, not arbitrary upstream raw keys or `include_full` sidecars:

- Fitness daily series from `get_fitness`: `ctl`, `atl`, `tsb`.
- Wellness daily series from `get_wellness_data`: `ramp`, `ctl_load`, `atl_load`, `rhr`, `hrv`, `hrv_sdnn`, `weight`, `kcal_consumed`, `sleep_secs`, `sleep_score`, `sleep_quality`, `avg_sleeping_hr`, `feel`, `soreness`, `fatigue`, `stress`, `mood`, `motivation`, `sp_o2`, `systolic`, `diastolic`, `hydration`, `hydration_volume`, `readiness`, `baevsky_si`, `blood_glucose`, `lactate`, `body_fat`, `abdomen`, `vo2max`, `steps`, `respiration`, `carbohydrates`, `protein`, `fat_total`. Boolean flags `tempWeight` and `tempRestingHR` are staged out of first-pass `analysis_metric` because they are indicators, not numeric analysis series.
- Activity/detail row metrics from `get_activities` / `get_activity_details`: `moving_time_seconds`, `elapsed_time_seconds`, `distance_km`, `distance_mi`, `pace_seconds_per_km`, `pace_seconds_per_mile`, `average_speed_kmh`, `average_speed_mph`, `max_speed_kmh`, `max_speed_mph`, `elevation_gain_m`, `elevation_loss_m`, `training_load`, `average_heart_rate_bpm`, `max_heart_rate_bpm`, `average_cadence_rpm`, `calories_burned`.
- Activity interval row metrics from `get_activity_intervals`: `duration_seconds`, `distance_m`, `average_power_watts`, `average_heart_rate_bpm`; interval row `pace` is deferred from the first-pass enum because its unit semantics depend on the interval row `unit` field and future analyzer unit-normalization rules. Interval-only metrics carry source-family metadata so analyzer tools can decide whether a window is activity-level or interval-level.
- Window aggregate rows from `get_training_summary`: `time_seconds`, `moving_time_seconds`, `elapsed_time_seconds`, `calories_burned`, `elevation_gain_m`, `distance_km`, `distance_mi`, `training_load`, `session_rpe`, `time_in_zones_total_seconds`.
- Analyzer-level derived weekly metrics from `get_training_summary`: `weekly_tss` means weekly-bucketed sum of source `training_load` (TSS-equivalent label from current read surface); `weekly_hours` means weekly-bucketed `time_seconds / 3600`. These stay in the enum because the PRD explicitly lists them, but their metadata must mark `Kind=derived`, `SourceFamily=derived_weekly`, source tools/fields, unit labels, and method/formula text for future `_meta.method`.
- Extended activity scalar metrics from `get_extended_metrics`: `stride_length_m`, `cardiac_decoupling_percent`, `pw_hr`, `aerobic_decoupling_percent`, `joules_above_ftp_kj`, `if`, `vi`, `polarization_index`, `trimp`, `strain_score`, `hr_load`, `pace_load`, `power_load`, `training_load`, `left_right_balance_percent`, `rpe`, `feel`, `session_rpe`, `compliance_pct`.
- Extended interval scalar metrics from `get_extended_metrics`: `dfa_alpha1`, `w_prime_balance_start_kj`, `w_prime_balance_end_kj`, `joules_above_ftp_kj`, `aerobic_decoupling_percent`, `left_right_balance_percent`, `stride_length_m`, `strain_score`, `training_load`.
- Structured/non-scalar source fields such as `power_zone_distribution_seconds`, `pace_zone_time_seconds`, interval group buckets, and curve buckets are excluded from scalar `analysis_metric`; their unknown names hint toward zone/load-balance or efforts analyzers instead of parsing as scalar trend/distribution metrics.
- Duplicate source descriptors are explicit: `ctl` and `atl` include both `get_fitness` daily sources and wellness-row copies from `get_wellness_data`; `feel` includes both wellness daily subjective scale and extended activity `feel`; duplicate `training_load`, `session_rpe`, distance/time/calorie totals, HR fields, and extended activity/interval overlaps carry all current source descriptors with distinct grain/source-family metadata.
- Curve/best-effort surfaces from `get_best_efforts` and `get_power_curves` are intentionally not scalar `analysis_metric` values because their meaning depends on duration/distance buckets; future analyzer tools should model those as structured arguments and unknown metric hints should point to `analyze_efforts_delta`.

Canonical naming and alias policy:

- Canonical enum values are lower `snake_case`; existing camelCase wellness keys become snake_case (`sleepSecs` -> `sleep_secs`, `restingHR` -> `rhr`, `rampRate` -> `ramp`). Intentional analyzer canonical divergences from current read JSON are source-field aliases: `rampRate` -> `ramp`, `intensity_factor` -> `if`, `variability_index` -> `vi`, and `compliance_percent` -> `compliance_pct`. Abbreviations are canonical only where already product-facing, PRD-listed, or physiology-standard: `ctl`, `atl`, `tsb`, `hrv`, `rhr`, `if`, `vi`, `np` if added later, `pw_hr`, `rpe`, `trimp`.
- Unit suffixes are part of the canonical name when the current read surface is unit-specific (`distance_km`, `distance_mi`, `pace_seconds_per_km`, `average_heart_rate_bpm`, `joules_above_ftp_kj`). Unit-contextual fields without a stable unit-explicit key, such as interval `pace`, are deferred until analyzer unit normalization owns them.
- Collisions keep the source field name: `fatigue` means wellness subjective fatigue, while `atl` remains the fitness/fatigue load metric; `training_load` is activity or summary load, while `hr_load`, `pace_load`, and `power_load` stay source-specific.
- Conservative aliases are allowed only for exact casing/style or established expansion: `resting_hr`, `restingHR`, `resting_heart_rate`, `restingHeartRate` -> `rhr`; `sleepSecs`, `sleep_seconds` -> `sleep_secs`; `sleepScore` -> `sleep_score`; `sleepQuality` -> `sleep_quality`; `hrvSDNN`, `hrv_sdnn_ms` -> `hrv_sdnn`; `intensity_factor` -> `if`; `variability_index` -> `vi`; `compliance_percent`, `compliance_percentage` -> `compliance_pct`; `rampRate`, `ramp_rate` -> `ramp`; exact canonical names parse unchanged.
- Near-misses and ambiguous names are rejected rather than guessed: `fatigue_load`, `fitness_fatigue`, `distance`, `ftp`, `tss`, `sleep`, `score`, `load`, and bucketed efforts like `5min_power` until a structured analyzer argument owns them. PRD examples `pace_at_lt2`, `power_at_lt2`, and `np` remain intentional v0.6 staging follow-ups because current read tools do not expose stable scalar fields for them yet; they should be added to this enum only when a source tool or formula registry entry owns their definitions.
- Any free-form arithmetic/expression syntax is invalid even when component fields are known: `ctl/atl`, `ctl - atl`, `weekly_tss/weekly_hours`, `tss_per_hour`, `power:weight`, `np/ftp`, `(ctl+atl)/2`, `hrv*sleep_quality`, and comma/pipe joined formulas are rejected.

Unknown-metric hint strategy is deterministic and one-line: arithmetic/expression inputs return `invalid analysis_metric: expressions are not supported; choose a supported metric`; effort duration/distance names (`5min_power`, `20_min_power`, `5k_pace`) return `try analyze_efforts_delta for best-effort durations/distances`; zone distribution or load-balance names return `try compute_zone_time or compute_load_balance for zone distributions`; segment/stream-stat names (`mean_power_segment`, `hr_drift_stream`) return `try compute_activity_segment_stats for within-activity stream stats`; compliance/adherence names return `try compute_compliance_rate for scheduled-vs-completed analysis`; otherwise return `invalid analysis_metric: use one of: ctl, atl, tsb, weekly_tss, hrv, sleep_secs, if, vi` with a concise sample, not the full internal list.

Implementation boundary for Step 2: create SDK-free `internal/analysis` package with typed `Metric` string enum, `ParseMetric(string) (Metric, error)`, `MustParseMetric` avoided, `MetricValues() []string` for schemas/tests, `MetricSources(metric) []MetricSource` returning one or more source descriptors rather than a single source, `MetricSchemaDescription()` or `MetricSchemaProperty()` helper, alias metadata kept private, and an exported invalid-argument error type or predicate that exposes only short user text plus optional hint. `MetricSource` metadata carries at least `SourceFamily`, `SourceTool`, `SourceField`, `Grain` (`daily`, `activity`, `interval`, `summary_window`, `derived_weekly`), `Kind` (`scalar`, `subjective_scale`, `derived`, `structured_excluded` only for hint metadata if needed), `UnitLabel` (for example `seconds`, `hours`, `km`, `mi`, `bpm`, `%`, `kJ`, `TSS-equivalent`, `unitless`), optional `ScaleRef`/`ScaleLabel` for subjective fields (`feel`, `fatigue`, `sleep_quality`, `mood`, etc.), and optional `Method` for derived metrics. Source families include `fitness_daily`, `wellness_daily`, `activity_row`, `activity_interval`, `training_summary`, `extended_activity`, `extended_interval`, and `derived_weekly`. Duplicate canonical metrics such as `training_load`, `average_heart_rate_bpm`, `session_rpe`, distance/time/calorie totals, and extended interval/activity overlaps remain single user-facing enum values but carry multiple `MetricSource` entries; analyzer tools choose the source family that matches their window/grain and return an invalid-argument hint if no source matches. Full analyzer routing remains future-tool scope, but this metadata prevents each analyzer from duplicating the enum/source table. Future analyzer tools should depend on this package rather than duplicating enum lists. Schema implications: analyzer input schemas enumerate canonical values only and describe aliases in prose if needed. Generated tool docs are not affected until analyzer tools are registered; `CHANGELOG.md` is updated when behavior lands.

| 2026-05-20 13:37 | Review R001 | plan Step 1: UNKNOWN |
| 2026-05-20 13:43 | Review R002 | plan Step 1: REVISE |
| 2026-05-20 13:47 | Review R003 | plan Step 1: REVISE |
| 2026-05-20 13:52 | Review R004 | plan Step 1: REVISE |
| 2026-05-20 13:56 | Review R005 | plan Step 1: APPROVE |
| 2026-05-20 13:58 | Review R006 | code Step 1: APPROVE |
| 2026-05-20 14:01 | Review R007 | plan Step 2: APPROVE |
| 2026-05-20 14:06 | Review R008 | code Step 2: REVISE |
| 2026-05-20 14:09 | Review R009 | code Step 2: REVISE |
| 2026-05-20 14:12 | Review R010 | code Step 2: APPROVE |
| 2026-05-20 14:14 | Review R011 | plan Step 3: REVISE |
| 2026-05-20 14:15 | Review R012 | plan Step 3: APPROVE |
| 2026-05-20 14:21 | Review R013 | code Step 3: APPROVE |
| 2026-05-20 14:22 | Review R014 | plan Step 4: REVISE |
| 2026-05-20 14:23 | Review R015 | plan Step 4: REVISE |
| 2026-05-20 14:25 | Review R016 | plan Step 4: APPROVE |
| 2026-05-20 14:34 | Review R017 | code Step 4: APPROVE |
| 2026-05-20 14:38 | Review R018 | plan Step 5: APPROVE |
| 2026-05-20 14:52 | Review R019 | code Step 5: APPROVE |
