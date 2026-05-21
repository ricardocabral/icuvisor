# Plan Review — TP-087 Step 1: Design the enum and aliases

## Verdict: REVISE

The updated `STATUS.md` now contains the concrete Step 1 design record requested in R001, and it is much closer to implementable. However, I would not proceed to Step 2 yet because the proposed enum contract has a few inconsistencies that would be hard to unwind once analyzer tools start depending on it.

## Required revisions before implementation

1. **Remove or explicitly reclassify non-scalar/structured metrics from `analysis_metric`.**
   The design says the enum is limited to numeric/scalar fields, but it includes array/structured zone fields such as `power_zone_distribution_seconds` and `pace_zone_time_seconds`. Those should not be canonical scalar `analysis_metric` values if unknown zone-distribution names are supposed to hint toward `compute_zone_time` / `compute_load_balance`. Either:
   - exclude those distribution arrays from this enum and list them under hint-only/structured analyzer inputs, or
   - explicitly document a separate metric kind that analyzer implementations can distinguish from scalar values.

   As written, the plan contradicts itself and risks making `analyze_trend(metric="power_zone_distribution_seconds")` look valid even though it cannot produce a single scalar series without extra dimensions.

2. **Align canonical names with current public read-tool JSON fields or document every intentional divergence.**
   A few canonical names do not match the current read surfaces:
   - `if` / `vi` are proposed canonical names, while `get_extended_metrics` currently emits `intensity_factor` / `variability_index`.
   - `compliance_pct` is proposed canonical, while `get_extended_metrics` emits `compliance_percent`.
   - `ramp` is proposed canonical, while wellness emits `rampRate`.

   These may be good product-facing canonical names, and PRD examples include `if`, `vi`, and `compliance_pct`, but Step 1 should explicitly mark them as intentional analyzer canonical names with source-field aliases. Otherwise Step 2 may accidentally encode an undocumented rename away from the existing read contract.

3. **Fix source-surface attribution for activity/interval fields.**
   The inventory labels `duration_seconds`, `average_power_watts`, and `pace` under `get_activities` / `get_activity_details`, but those are exposed by `get_activity_intervals`, not the activity/detail row. Conversely, `get_activities` / `get_activity_details` expose moving/elapsed time, distance, pace-per-unit, speed, elevation, HR, cadence, calories, and training load. Please separate:
   - activity row metrics from `get_activities` / `get_activity_details`, and
   - interval row metrics from `get_activity_intervals`.

   This matters because future analyzers need to know whether a metric is windowed by activity rows, interval rows, wellness days, or fitness days.

4. **Resolve PRD-example exclusions deliberately.**
   The PRD’s analyzer design examples include `pace_at_lt2`, `power_at_lt2`, and `np`, but the plan lists them as rejected near-misses because no current read surface exposes them as scalar fields. That may be the right short-term decision for this task, but because the PRD is authoritative, record this as an intentional v0.6 staging decision/follow-up rather than a generic near-miss. Otherwise a future reviewer may read the Step 1 record as silently narrowing the PRD contract.

5. **Clarify whether the helper package exposes source metadata, not just values.**
   The proposed `internal/analysis` API has `Metric`, `ParseMetric`, and `MetricValues()`, but the design inventory is source-dependent. Future analyzer tools will need at least a stable way to know the metric family/source (`fitness_daily`, `wellness_daily`, `activity_row`, `activity_interval`, `extended_activity`, etc.) or Step 2 will create a parser that every analyzer must wrap with its own duplicate routing table. If full routing is out of scope, say so and record the follow-up; otherwise include a small metadata helper in the Step 2 boundary.

## What looks good

- The plan now records concrete supported metrics, alias rules, rejected expressions, hint categories, and a reusable `internal/analysis` package boundary.
- The alias policy is appropriately conservative and rejects ambiguous names instead of guessing.
- The free-form arithmetic rejection examples cover the important failure modes (`ctl/atl`, `weekly_tss/weekly_hours`, `np/ftp`, formulas, joined expressions).
- The schema direction is right: enumerate canonical values in JSON Schema and mention aliases only in prose.

## Recommendation

Revise the Step 1 design record to resolve the scalar-vs-structured contradiction, correct source attribution, and document intentional canonical-name divergences from current read fields/PRD examples. After that, Step 2 can implement the parser and schema helpers with much lower risk of freezing the wrong analyzer contract.
