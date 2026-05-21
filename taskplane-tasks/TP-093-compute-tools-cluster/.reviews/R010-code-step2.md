# R010 code review — Step 2: Implement aggregation logic

Verdict: REVISE

`go test ./...` passes, but the Step 2 aggregation implementation still has contract-level correctness gaps that should be fixed before registration.

## Blocking findings

### 1. Derived weekly baselines still compare current totals against baseline weekly means

- Location: `internal/tools/compute_baseline.go:113`, `internal/tools/compute_baseline.go:199-220`
- Severity: High

`collectSummaryBaseline` now buckets `weekly_tss` / `weekly_hours` into ISO-week sums for both baseline and current windows, but the handler still calls `analysis.ComputeBaselineStats(..., sumCurrent=true)` for those metrics. That makes `current_value` the sum of all current weekly buckets, while `baseline_mean` and `baseline_stddev` are per-week values.

The updated contract says derived weekly metrics compare the **current weekly-bucket mean** against the baseline weekly-bucket mean/stddev. A current window spanning two normal weeks will therefore look roughly 2x high and produce a bogus z-score. Pass the current weekly samples through the same mean path as the baseline, or otherwise make the contract and implementation agree.

### 2. Activity-backed baseline metrics do not mirror `get_activities` derived fields

- Location: `internal/tools/compute_baseline.go:397-436`; source catalog in `internal/analysis/metrics.go`
- Severity: High

Several supported `analysis_metric` values are advertised as `SourceActivityRow` fields from `get_activities` (`distance_km`, `distance_mi`, `pace_seconds_per_km`, `pace_seconds_per_mile`, `average_speed_kmh`, `average_speed_mph`, `max_speed_kmh`, `max_speed_mph`). The baseline collector chooses the activity source first, but `activityMetricValue` only handles a subset of typed upstream fields and then looks for the canonical output field name in `activity.Raw`.

Those canonical fields are computed by the read-tool shaper from meters / m/s / moving time; they are not guaranteed to exist in upstream raw payloads. As a result, `compute_baseline` can return `missing_metric` / `insufficient_sample` for metrics that `get_activities` would successfully expose from the same rows. The same issue exists for summary-backed fallbacks such as `distance_km`, `distance_mi`, and `session_rpe` in `summaryMetricValue` / `summaryMetricValueForSport`.

Please either reuse the same conversion helpers/semantics as the source tools or implement equivalent deterministic conversions for every catalogued baseline metric before exposing the tool.

### 3. Activity truncation does not make `compute_baseline` partial or add the required boundary

- Location: `internal/tools/compute_baseline.go:119-121`, `internal/tools/compute_baseline.go:256-262`
- Severity: High

The shared activity-acquisition contract says any activity-backed compute path that reaches `max_activity_candidates` must return `status:"partial"`, set `truncated_activity_candidates:true`, set `_meta.assumptions.activity_candidates_truncated:true`, and record a boundary that aggregates may exclude activities past the cap.

`collectActivityBaseline` sets `Truncated` when `len(activities) >= 500`, but the handler leaves `result.status` as whatever the z-score calculation returned, so a truncated activity baseline can still report `status:"ok"`. The meta assumptions include the flag, but the boundary is also missing. This silently publishes a complete-looking baseline over a capped candidate set.

### 4. Compliance truncation metadata is incomplete and status precedence violates the contract

- Location: `internal/tools/compute_compliance_rate.go:282-297`
- Severity: High

The compliance contract says reaching the event cap returns `status:"partial"` and includes a boundary noting that scheduled-event denominators may exclude rows past the cap. The implementation only sets partial in the `else if activityTruncated || eventTruncated` branch after `acc.scheduled == 0`, so a capped event response with no target events in the first page returns `status:"insufficient_sample"` even though later capped-off events could contain valid targets. The `_meta.boundaries` list is also unchanged for event or activity truncation.

This keeps denominators ambiguous in exactly the capped case the contract was meant to make explicit. Mark capped responses as partial (while still surfacing an insufficient reason if useful) and add the required truncation boundary text.

### 5. Compliance breakdown rows omit `delta_sample_count`

- Location: `internal/tools/compute_compliance_rate.go:502-506`
- Severity: Medium

The contract now states that mean deltas exclude unpaired events and that the denominator is exposed as `delta_sample_count` overall **and in breakdown rows**. The overall result sets `DeltaSampleCount: acc.completed`, but `breakdowns` never fills `DeltaSampleCount`, so `by_sport` and `by_event_type` rows omit the denominator for their mean deltas.

This makes the default terse breakdowns insufficient for the promised “mean delta to target, per sport / event type” use case. Set each breakdown's `DeltaSampleCount` from its accumulator's completed count.

## Non-blocking findings

### 6. `compute_load_balance.training_load_total` does not use the documented load-source priority

- Location: `internal/tools/compute_zone_time.go:240-256`, `internal/tools/compute_zone_time.go:300-305`
- Severity: Medium

The contract says contextual load totals use precomputed load fields in priority `power_load`, `hr_load`, `pace_load`, `icu_training_load` from extended activity raw, or summary `TrainingLoad`. The activity-backed path only sums `activity.TrainingLoad` or raw `icu_training_load` from the list row, and it does not inspect the extended-detail raw payload even when that payload supplied the zone array.

So activities whose zones and load are both available only in extended raw can contribute zone seconds but no `training_load_total`, and metric-specific `power_load` / `hr_load` / `pace_load` are ignored.

### 7. Task status metadata still contradicts review artifacts

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/STATUS.md:105-113`, `taskplane-tasks/TP-093-compute-tools-cluster/STATUS.md:149-155`, `.reviews/R009-code-step1.md:1-4`
- Severity: Medium

The reviews table now records R009 as `APPROVE`, but `.reviews/R009-code-step1.md` says `Verdict: REVISE`. The execution log also still says R005/R006/R007 were approved even though the corresponding review files are revisions. Please keep the status log consistent so the next step does not proceed from false approvals.
