# Plan Review — TP-087 Step 1: Design the enum and aliases

## Verdict: REVISE

The Step 1 record has addressed the previous review's main items: structured zone arrays are now excluded from the scalar enum, canonical-name divergences are documented, activity vs interval row sources are separated, PRD-example exclusions are staged, and the proposed `internal/analysis` boundary is much clearer.

I would still revise the plan before Step 2 because the current enum/source model is internally ambiguous in ways that will make the shared analyzer contract hard to implement safely.

## Required revisions before implementation

1. **Resolve duplicate canonical metrics across source families.**

   The plan proposes a single `MetricInfo(metric) (MetricMetadata, bool)` with one `SourceFamily` and one source field, but several canonical metric names appear in multiple source surfaces with different row grains/semantics:

   - `training_load`: activity row (`get_activities` / `get_activity_details`), summary aggregate (`get_training_summary`), extended activity, and extended interval.
   - `moving_time_seconds`, `elapsed_time_seconds`, `distance_km`, `distance_mi`, `calories_burned`, `elevation_gain_m`: activity rows and training-summary aggregates.
   - `average_heart_rate_bpm`: activity rows and interval rows.
   - `session_rpe`: training summary and extended activity metrics.
   - `stride_length_m`, `aerobic_decoupling_percent`, `joules_above_ftp_kj`, `left_right_balance_percent`, `strain_score`, and `training_load`: appear in extended activity and/or extended interval shapes.

   A closed enum can still use these names, but the metadata API must not pretend each metric has exactly one source. Either split ambiguous names into distinct canonical metrics, exclude duplicates until the owning analyzer needs them, or change the API plan to return multiple source descriptors plus analyzer-specific selection rules/defaults. Without this, future analyzers will duplicate routing tables or accidentally analyze an interval metric as an activity/window aggregate.

2. **Reclassify `weekly_tss` and `weekly_hours` as derived/aggregated metrics, not existing read-tool fields.**

   `get_training_summary` currently exposes `training_load`, `time_seconds`, and related totals; it does not expose JSON fields named `weekly_tss` or `weekly_hours`. The PRD does list `weekly_tss` and `weekly_hours`, so keeping them may be correct, but the Step 1 record should explicitly mark them as analyzer-level aggregate metrics with source fields and aggregation semantics, for example `weekly_tss = weekly bucketed training_load` and `weekly_hours = weekly bucketed time_seconds / 3600`.

   If this task is intended to include only fields directly mirrored from current read surfaces, stage these two like `np` / `pace_at_lt2` / `power_at_lt2`. If they remain in v0, add metadata for source field, grain/window, unit, and future formula reference so they do not look like raw supported read fields.

3. **Specify enough metadata for units, scale, grain, and metric kind.**

   The planned `SourceFamily` + source field metadata is a good start, but the enum inventory includes materially different metric classes:

   - daily scalar series (`ctl`, `atl`, `hrv`, `sleep_secs`),
   - per-activity rows (`distance_km`, `training_load`),
   - interval rows (`duration_seconds`, interval `pace`),
   - summary/window aggregates (`time_seconds`, `time_in_zones_total_seconds`),
   - subjective scales (`feel`, `fatigue`, `sleep_quality`, `mood`, etc.), and
   - derived/aggregate PRD metrics if `weekly_tss` / `weekly_hours` stay in the enum.

   Step 2 needs a design target for at least `Unit`/unit label, row grain, metric kind, and optional scale metadata or scale reference. This is also needed for schema descriptions and future `_meta.method` / `_meta.source_tools` requirements. It does not need to implement analyzer routing now, but the shared contract should preserve this information instead of flattening everything to one source field string.

4. **Clarify interval `pace` canonical naming.**

   The plan includes interval `pace` without a unit suffix while the general naming rule says unit suffixes are part of canonical names when the current read surface is unit-specific. `get_activity_intervals` exposes `pace` alongside a row `unit`, so the plan should state whether `pace` is allowed as a unit-contextual interval metric, whether it should be renamed to a unit-explicit canonical value, or whether it is deferred until analyzer unit normalization owns it. This prevents a schema that says `pace` is a scalar metric but gives the LLM no unit semantics.

## What looks good

- The alias policy is conservative and rejects ambiguous near-misses instead of guessing.
- Free-form arithmetic rejection examples are explicit and cover the important cases.
- The unknown-metric hint categories are deterministic and aligned with the planned analyzer family.
- Excluding zone arrays from scalar `analysis_metric` is the right direction; those should route toward zone/load-balance analyzers.

## Recommendation

Revise `STATUS.md` to make the enum/source metadata unambiguous before implementation. The key decision is whether `analysis_metric` is only a closed string validator or a reusable analyzer contract with source/aggregation metadata. Given this task's mission, it should be the latter; otherwise Step 2 will freeze a parser that future analyzer tools immediately need to wrap or work around.
