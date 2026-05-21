# TP-093 deterministic compute-tool contracts

This note pins the public contract before implementation. All date inputs are athlete-local `YYYY-MM-DD` inclusive windows. All tools reject inverted windows and never impute missing days or zero-fill absent upstream arrays.

## Shared analyzer response contract

Every tool returns `{ "result": ..., "_meta": ... }` by default and adds an audit `series` collection only when `include_full:true`. The `_meta` object uses `analysis.AnalyzerMeta` fields: `method`, sorted `source_tools`, `n`, `missing_days`, `missing_action:"skip"`, `insufficient_sample`, optional `formula_ref`, `assumptions`, and `boundaries`. Missing days mean calendar dates in the requested inclusive window for which no usable source row existed. Rows with a source record but no required metric are counted separately in result-level `missing_sources` and in `_meta.assumptions` so missing data is explicit without pretending the date did not exist.

Shared schema fields:

- `include_full` (optional boolean, default `false`): include per-day/per-event audit rows and raw-source summaries sufficient to explain aggregates. Terse mode returns aggregate results, missing/insufficient flags, and metadata only.
- `sport` (optional string): intervals.icu activity type/category filter. For summary-backed scalar metrics, filter `SummaryWithCats.byCategory` only when that category row contains the requested scalar. For zone distributions, `SummaryWithCats.TimeInZones` is a whole-day array and `CategorySummary` does not contain zone arrays, so summary zones are valid only for unfiltered windows or dates proven homogeneous by one category/count. Sport-filtered zone/load-balance queries must use per-activity precomputed arrays when activity rows are available; otherwise those rows are partial/unavailable rather than returning whole-day zones for the requested sport. For activity/event-backed tools, filter activity or event type/category exact case-insensitive.
- No compute tool accepts raw stream values. `get_activity_streams` is not a fallback source for these tools; if precomputed zone/load fields are unavailable the tool returns `status:"unavailable"` or `insufficient_sample:true` with an explicit reason.

### Activity acquisition for activity-backed paths

When a tool needs activity candidates (`compute_zone_time` sport/metric-specific paths, `compute_load_balance`, activity/extended `compute_baseline` sources, or `compute_compliance_rate` auto-pairing), it requests `get_activities`/`ActivitiesClient.ListActivities` over the relevant inclusive local date window with stable fields needed for ID, sport/type, start date, moving/elapsed time, distance, training load, and raw link IDs. For `compute_baseline`, the relevant activity window is the union from `baseline_start_date` through `current_end_date`; the 500-candidate cap applies to that union and a single `truncated_activity_candidates` flag means either baseline or current activity samples may be incomplete. The implementation must exhaust the available candidate list up to `max_activity_candidates = 500`; if the upstream/client result reaches that cap or otherwise signals truncation before exhaustion, the tool returns `status:"partial"`, sets `_meta.assumptions.activity_candidates_truncated:true`, includes `truncated_activity_candidates:true` in `result`, and records a boundary that aggregates may exclude activities past the cap. Candidate ordering is normalized by local start date/time then activity ID before any matching or per-activity metric fetches. `source_tools` must include `get_activities` whenever this acquisition path is used.

## `compute_zone_time`

### Request schema

Required: `start_date`, `end_date`, `zone_metric`.
Optional: `sport`, `include_full`.

`zone_metric` enum: `power`, `heart_rate`, `pace`.

### Source priority

1. For unfiltered windows only, `get_training_summary` / `FitnessClient.ListAthleteSummary` daily `SummaryWithCats.TimeInZones` and `TimeInZonesTot` may be used for `zone_metric:"power"` as upstream ICU training-load/power-style zones. For `heart_rate` or `pace`, summary zones may be used only if a concrete upstream raw marker identifies the zone family, such as `zone_family`, `zone_metric`, `zone_type`, or a namespaced key containing `hr`/`heart_rate` or `pace` alongside the zone array; otherwise use activity precomputed arrays or return unavailable/partial. These whole-day arrays are never used for sport-filtered mixed days because `CategorySummary` has no zone arrays. If a day has exactly one category matching `sport` and no other category/count, the summary row is deterministic for that sport and may contribute with `_meta.assumptions.summary_zone_homogeneous:true`.
2. For sport-filtered queries, non-power metrics, or mixed days, use `get_activities` to enumerate candidate activity IDs and `get_extended_metrics` / activity raw precomputed fields: `icu_zone_times` for `power`, `gap_zone_times` then `pace_zone_times` for `pace`, and HR arrays only if present under raw keys matching `hr_zone_times`, `heartrate_zone_times`, `heart_rate_zone_times`, or `hr_time_in_zones`.
3. No stream fallback. Absence of deterministic precomputed arrays yields `status:"unavailable"` or `status:"partial"`, `missing_sources:[...]`, and `_meta.boundaries` explaining that raw streams were intentionally not reduced manually. The implementation must not return the same summary zone array as `power`, `heart_rate`, and `pace` unless the upstream source identifies it as valid for that metric.

### Calculation and response

Aggregate seconds element-wise by upstream zone order and compute the same three-bucket polarization summary used by `compute_load_balance`: low = Z1+Z2, moderate = Z3, high = Z4+. `result` fields: `status` (`ok`, `partial`, `unavailable`), `zone_metric`, `sport`, `start_date`, `end_date`, `zones` (`zone`, `seconds`, `share`), `total_seconds`, `polarization_index`, `polarization_state`, `classification`, `missing_sources`, `insufficient_reason`. `_meta.method:"precomputed_zone_time_sum"`, `source_tools` as used, `n` = count of source rows contributing at least one zone second, `formula_ref:"icuvisor://analysis-formulas#polarization_index"` when total bucketed time is positive, and boundaries include "precomputed zones only; raw streams are not reduced". Full `series` contains date/activity rows with selected source key, zone seconds, low/moderate/high bucket seconds, and missing reason.

## `compute_load_balance`

### Request schema

Required: `start_date`, `end_date`.
Optional: `zone_metric` (`power`, `heart_rate`, `pace`; default `power`), `sport`, `include_full`.

### Source priority

1. Reuse `compute_zone_time` precomputed zone aggregation and the same deterministic restrictions for the requested metric/sport: summary `TimeInZones` only for unfiltered or proven homogeneous rows, and per-activity precomputed arrays for sport-filtered or metric-specific rows.
2. Use precomputed load fields only for contextual totals, not classification. `training_load_total` is the sum across the same source rows that contributed zone seconds. For each contributing activity row, add the first available load in priority `power_load`, `hr_load`, `pace_load`, `icu_training_load`, then typed/list-row training load; for summary-backed rows add `SummaryWithCats.TrainingLoad`. Do not synthesize a load for zone rows without any load field; full series records the selected load source when present.
3. No stream fallback.

### Calculation and response

Map zones to three buckets: low = Z1+Z2, moderate = Z3, high = Z4+. Compute shares from bucket seconds. If total bucketed seconds is zero, return unavailable. If moderate or high share is zero, return `polarization_index:null` with a `polarization_state` of `undefined_moderate_zero` or `undefined_high_zero`; classification falls back to shares. Otherwise compute `log10((low_share / moderate_share) * (high_share / moderate_share) * 100)` using `icuvisor://analysis-formulas#polarization_index`.

`result` fields: `status`, `zone_metric`, `sport`, `buckets`, `polarization_index`, `polarization_state`, `classification` (`polarized`, `pyramidal`, `threshold`, `unclassified`), `training_load_total`, `missing_sources`, `insufficient_reason`. `_meta.method:"precomputed_zone_load_balance"`, `formula_ref:"icuvisor://analysis-formulas#polarization_index"`, `n` = contributing zone rows. Full `series` mirrors zone-time audit rows plus per-row low/moderate/high seconds.

Classification rule: `polarized` when low share >= 0.70 and high share >= moderate share; `pyramidal` when low > moderate > high; `threshold` when moderate >= low or moderate >= high and moderate share >= 0.30; otherwise `unclassified`. These are heuristic labels and `_meta.assumptions.classification_rule` records the thresholds.

## `compute_baseline`

### Request schema

Required: `metric`, `baseline_start_date`, `baseline_end_date`, `current_start_date`, `current_end_date`. The baseline window must end before the current window starts (`baseline_end_date < current_start_date`); overlapping or reversed cross-window requests are invalid so baseline/current sample membership and activity cap scope stay deterministic.
Optional: `sport`, `min_samples` (integer, minimum 2, default `analysis.MinBaselineSamples` = 7), `include_full`.

`metric` uses the closed `analysis_metric` enum and aliases accepted by `analysis.ParseMetric`; examples: `ctl`, `atl`, `tsb`, `weekly_tss`, `hrv`, `sleep_secs`, `if`, `vi`.

### Source priority

`compute_baseline` accepts the closed `analysis_metric` schema, then resolves the canonical metric through `analysis.ParseMetric` and `analysis.MetricSources(metric)` in catalog order. It supports source families that can produce deterministic window samples without raw stream math:

- `fitness_daily` and `training_summary`: `get_fitness`/`get_training_summary` daily summary fields such as CTL, ATL, TSB, training load, time, distance, calories, elevation, and weekly derived load/hours. Sport filters are honored only when the underlying summary/category scalar can be isolated; otherwise the row is marked missing for that sport.
- `wellness_daily`: `get_wellness_data` daily wellness and subjective-scale fields, including RHR, HRV, sleep, nutrition, body metrics, feel/soreness/fatigue/stress/mood/motivation, and provider readiness/scale values.
- `activity_row`: `get_activities` activity-level fields such as moving/elapsed time, distance, pace/speed, elevation, load, HR, cadence, and calories; sport filters apply at activity row level.
- `extended_activity`: `get_extended_metrics` precomputed activity-level metrics such as IF, VI, decoupling, load variants, TRIMP, strain, balance, RPE, and compliance percent.
- `derived_weekly`: analyzer-created weekly samples from supported daily/window fields, e.g. `weekly_tss` and `weekly_hours`.

Metrics whose only available sources are interval-grain families (`activity_interval`, `extended_interval`) return `status:"unsupported_metric_source"` with `insufficient_reason:"interval_grain_not_supported_for_baseline"` unless the same metric also has a supported daily/activity/window source in `MetricSources`. No raw stream fallback is allowed.

### Calculation and response

Collect baseline samples after applying sport filters where the source supports them. Source fallback is deterministic: try `analysis.MetricSources(metric)` in catalog order; if a source is unavailable because its client is absent or unsupported interval grain, try the next source, but if a supported source returns rows with insufficient baseline/current samples, return that insufficiency rather than silently falling through to a lower-priority source. Baseline/current samples are formed by source grain: `daily` uses one value per local date and compares the current-window mean against the baseline daily-sample mean/stddev; `activity` uses one value per activity and compares current activity mean against baseline activity mean/stddev; `summary_window` uses the lower-level daily `SummaryWithCats` rows returned by the summary client for the requested inclusive range, emits one sample per local date with a usable value, and records `_meta.assumptions.summary_window_sample_grain:"daily_summary_rows"`; `derived_weekly` buckets both baseline and current windows into ISO-week sums using the same field, then compares the current weekly-bucket mean against baseline weekly-bucket mean/stddev. Totals such as training load, time, distance, calories, steps, and nutrition remain daily/activity sample values unless their metric source is explicitly `derived_weekly`; they are not summed across the whole current window before z-scoring. `n` is usable baseline sample count. If `n < min_samples`, return `status:"insufficient_sample"`, `z_score:null`, `_meta.insufficient_sample:true`, and no imputation. If the current window has zero usable samples, return `status:"insufficient_current_sample"`, `current_value:null`, `z_score:null`, and `insufficient_reason:"no_current_samples"`. If standard deviation is zero, return `status:"insufficient_variance"`, `z_score:null`, `insufficient_reason:"zero_baseline_variance"`.

`result` fields: `status`, `metric`, `metric_source` (selected `analysis.MetricSource` family/tool/field/grain), `baseline_window`, `current_window`, `current_value`, `baseline_mean`, `baseline_stddev`, `z_score`, `interpretation`, `n_baseline`, `n_current`, `min_samples`, `missing_baseline_days`, `missing_current_days`, `insufficient_reason`.

`interpretation` is always present. It is `not_interpreted` for non-wellness/non-directional metrics and whenever z-score is null. Wellness directionality is deterministic and recorded in `_meta.assumptions.interpretation_direction`: beneficial-high metrics (`hrv`, `hrv_sdnn`, `sleep_secs`, `sleep_score`, `sleep_quality`, `readiness`, `feel`, `mood`, `motivation`, `steps`) interpret `z_score <= -1.0` as `suppressed`, `z_score >= 1.0` as `elevated_beneficial`, otherwise `typical`; adverse-high metrics (`rhr`, `avg_sleeping_hr`, `soreness`, `fatigue`, `stress`, `baevsky_si`, `blood_glucose`, `lactate`) interpret `z_score >= 1.0` as `elevated`, `z_score <= -1.0` as `suppressed_beneficial`, otherwise `typical`. Neutral body/nutrition measurements (`weight`, `kcal_consumed`, `hydration`, `hydration_volume`, `sp_o2`, `systolic`, `diastolic`, `body_fat`, `abdomen`, `vo2max`, `respiration`, `carbohydrates`, `protein`, `fat_total`) emit `not_interpreted` unless a later formula contract defines directionality. `_meta.method:"baseline_z_score"`, `formula_ref:"icuvisor://analysis-formulas#z_score"`, `n` = baseline samples. Full `series` contains baseline/current sample rows with date, value, source tool, and missing reason.

## `compute_compliance_rate`

### Request schema

Required: `start_date`, `end_date`.
Optional: `sport`, `event_type`, `category` (default `WORKOUT`), `tolerance_percent` (default 20, min 0, max 100), `target_metric` enum (`time`, `distance`, `load`; default `time`), `include_full`.

`event_type` filters scheduled events by exact case-insensitive `Event.Type`/raw `type`. `sport` always filters completed activity sport/type; it filters scheduled event type only when `event_type` is omitted. When both `sport` and `event_type` are supplied, `event_type` alone controls scheduled-event type matching and `sport` controls completed-activity candidates; auto-pairing compares candidate activity sport to `sport` when `sport` is present, otherwise to the scheduled event type.

### Source priority and pairing semantics

1. `get_events` / `EventsClient.ListEvents` with `resolve:true`, `limit = max_event_candidates = 500`, and the requested date window. If the returned event count reaches the cap or upstream reports truncation, return `status:"partial"`, `truncated_event_candidates:true`, `_meta.assumptions.event_candidates_truncated:true`, and a boundary noting that scheduled-event denominators may exclude rows past the cap. Scheduled rows are events matching `category`, `event_type`, and sport/type filters and having the selected target: for `time`, `time_target` or `elapsed_time_target`; for `distance`, `distance_target`; for `load`, `load_target`.
2. Existing upstream/manual links are reused only from enumerated keys. Event raw accepted activity-ID keys: `activity_id`, `icu_activity_id`, `paired_activity_id`, `completed_activity_id`. Activity raw accepted event-ID keys: `paired_event_id`, `event_id`, `calendar_event_id`, `icu_event_id`. Typed fields take precedence when they exist in future models; otherwise these raw keys are parsed as strings/numbers. If a link exists, `pairing_source:"linked"`; if the link was created by `link_activity_to_event`, it is still surfaced as linked because the upstream `paired_event_id` is authoritative.
3. Matching is one-to-one. Process scheduled events by start date/time, then event ID. Linked pairs reserve their activity IDs first; conflicting linked events keep the earliest event and mark later conflicts `pairing_source:"linked_conflict"` with no reusable completion. If no linked activity is visible, auto-pair by same local date and the deterministic sport key (request `sport` when present, otherwise scheduled `Event.Type`) using unused `get_activities` rows; choose the nearest target metric match deterministically by smallest absolute percent difference, then stable activity ID, and remove the chosen activity from the candidate set. Auto-paired rows use `pairing_source:"date_metric_match"`. One completed activity can never satisfy multiple scheduled events.
4. When a paired scheduled event has a non-empty `workout_doc`, call `get_activity_intervals` and use `analysis.IntervalExecutionClaimPolicy`; unavailable/non-cancellation interval responses do not fail aggregate compliance, but add a series caution and `_meta.boundaries` entry that interval execution could not be verified; if auto-lap is suspected, do not claim interval compliance and emit `auto_lap_caution:true` plus `_meta.auto_lap_suspected:true`. Aggregate target-metric compliance can still be reported with the caution.
5. No raw stream fallback.

### Calculation and response

For each scheduled target event, pair at most one completed activity and compare actual vs target for the selected metric. For `time`, `time_target` maps to activity `moving_time`; `elapsed_time_target` maps to activity `elapsed_time`; when both targets are present, `time_target`/moving time is primary and elapsed-time delta is emitted only as an audit field in full series. For `distance`, compare `distance_target` meters to activity distance meters. For `load`, compare `load_target` to activity training load. A pair is compliant when absolute percent difference <= `tolerance_percent`. Events without a target for the selected metric are excluded from denominator and reported in `excluded_events`; scheduled events with targets and no completion count as non-compliant. Deltas are signed as `actual - target`, so negative means under target.

`result` fields: `status`, `start_date`, `end_date`, `sport`, `event_type`, `target_metric`, `tolerance_percent`, `scheduled_count`, `completed_count`, `compliant_count`, `compliance_rate`, `delta_sample_count`, `mean_delta_percent`, one metric-specific absolute mean delta (`mean_delta_seconds`, `mean_delta_meters`, or `mean_delta_load`), `by_sport`, `by_event_type`, `excluded_events`, `unpaired_events`, `auto_lap_caution`, `truncated_activity_candidates`, `truncated_event_candidates`, `insufficient_reason`. Mean deltas use only paired completed activities as the denominator; unpaired scheduled events count as non-compliant in `compliance_rate` but do not contribute a synthetic -100% delta. The denominator is exposed as `delta_sample_count` overall and in breakdown rows. `_meta.method:"scheduled_completed_event_compliance"`, no `formula_ref`, `source_tools` include `get_events`, `get_activities`, and conditionally `get_activity_intervals`, `n` = scheduled target events in denominator. Default `by_sport` and `by_event_type` breakdown rows repeat scheduled/completed/compliant counts, compliance rate, and mean deltas so users do not need `include_full:true` to calculate "mean delta to target, per sport / event type". Full `series` contains event-level rows with event ID/name/date/type/sport, target, paired activity ID, actual, absolute and percent difference, compliant bool, pairing source, elapsed-time secondary delta when applicable, and caution reason.

## Implied golden coverage

- Zone time uses precomputed summary/extended fields and never `get_activity_streams` when arrays exist.
- Missing dates and rows with missing precomputed arrays are explicit in result and `_meta` without forward-fill.
- Baseline insufficient samples and zero variance suppress z-score with explicit reasons.
- Load balance produces polarization index/formula ref and handles moderate/high zero states.
- Compliance counts scheduled/completed pairings deterministically, reuses linked pairs before auto-pairing, and propagates auto-lap caution for interval-derived claims.
- No-precomputed fallback returns unavailable/insufficient instead of stream math.
