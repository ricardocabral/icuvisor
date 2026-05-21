# R007 code review — Step 2: Implement aggregation logic

Verdict: REVISE

`go test ./...` passes, but the Step 2 implementation has several contract-level correctness issues that should be fixed before registration/documentation.

## Blocking findings

### 1. Compliance auto-pairing can steal activities that have later linked events

- Location: `internal/tools/compute_compliance_rate.go:191-225`, `internal/tools/compute_compliance_rate.go:352-369`
- Severity: High

The contract requires linked pairs to reserve activity IDs before deterministic auto-pairing, and linked conflicts to surface as `pairing_source:"linked_conflict"`. The implementation processes each event in order and only checks `used` inside `linkedActivityForEvent`; if an earlier unlinked event is auto-paired first, it can consume an activity that a later event is explicitly linked to. Conversely, when two events link to the same activity, the later event falls through to auto-pairing or `unpaired` instead of being marked `linked_conflict`.

This changes `completed_count`, `compliant_count`, per-event audit rows, and the headline compliance rate. Build a linked-reservation pass over the scheduled target events before auto-pairing, record conflicts explicitly, and only let auto-pairing consider activities not reserved by authoritative links.

### 2. `weekly_tss` / `weekly_hours` baseline samples are daily rows, not derived weekly buckets

- Location: `internal/tools/compute_baseline.go:155-178`, `internal/tools/compute_baseline.go:286-316`, `internal/analysis/compute.go:95-103`
- Severity: High

`analysis.MetricSources` marks `weekly_tss` and `weekly_hours` as `SourceDerivedWeekly` / `GrainDerivedWeekly`, but `collectBaselineSamples` sends that family directly to `collectSummaryBaseline`, where each summary day is appended as an independent sample. The current window is then summed while the baseline mean/stddev are computed over daily samples, not over analyzer-created weekly sums.

For a two-week baseline of 100 TSS/week and a one-week current value of 100 TSS, this implementation can compare `current_value=100` against a baseline mean around `14.3` (daily average), producing a bogus z-score. Implement the derived weekly bucketing promised in the contract (same bucket rule for baseline and current), or do not advertise these metrics as supported for baseline yet.

### 3. Summary-backed baseline ignores the `sport` filter

- Location: `internal/tools/compute_baseline.go:178-205`
- Severity: High

The contract says summary scalar metrics may honor `sport` only when the requested sport/category scalar can be isolated; otherwise the row should be marked missing for that sport rather than returning whole-day values. `collectSummaryBaseline` never reads `args.Sport` and always uses whole-day `SummaryWithCats` fields. A sport-filtered `compute_baseline` for a summary-backed metric can therefore include other sports in both baseline and current windows.

Either isolate the requested category from `SummaryWithCats.ByCategory` where deterministic, fall back to an activity-backed source when available, or mark those rows missing/unsupported for sport-filtered summary metrics.

### 4. Compliance silently caps scheduled events without partial/truncation metadata

- Location: `internal/tools/compute_compliance_rate.go:168`, `internal/tools/compute_compliance_rate.go:260-275`
- Severity: High

The implementation calls `ListEvents` with `Limit: maxEventsLimit` but never detects or reports that the scheduled-event candidate set may have hit that cap. If intervals.icu returns 500 events for a busy window, the denominator can exclude additional scheduled target events while still reporting `status:"ok"`, with no `truncated_event_candidates` flag or `_meta` boundary.

This was already called out by the preceding plan review as a determinism requirement. Add an event truncation rule/field analogous to activity truncation and mark compliance results partial when the event candidate list reaches the deterministic cap (or page/exhaust events if supported).

## Non-blocking findings

### 5. `compute_load_balance` never populates `training_load_total`

- Location: `internal/tools/compute_zone_time.go:81-91`, `internal/tools/compute_zone_time.go:293-295`
- Severity: Medium

The result type exposes `TrainingLoadTotal`, and the contract says load balance should include contextual load totals from `power_load`, `hr_load`, `pace_load`, `icu_training_load`, or summary `TrainingLoad`. The aggregate collector currently stores only zone seconds, and `loadBalanceResult` never sets `TrainingLoadTotal`, so the public field is always omitted.

If this field is part of the Step 2 contract, collect and aggregate the load source in parallel with the zone source; otherwise remove it from the result shape/contract before the tool is registered.

### 6. Compliance interval source metadata is only attached when auto-lap is suspected

- Location: `internal/tools/compute_compliance_rate.go:241-247`, `internal/tools/compute_compliance_rate.go:275-278`
- Severity: Medium

When `event.WorkoutDoc != nil`, the handler calls `GetActivityIntervals` to evaluate interval-execution evidence, but `_meta.source_tools` only gets `get_activity_intervals` through `ApplyIntervalSourceEvidence` when `autoLap` is true. Successful interval checks with no auto-lap still used that source and should be represented in `source_tools`/interval-source metadata per the contract. Track whether interval evidence was attempted/used separately from whether it produced a caution.

### 7. Status metadata records R006 as approved even though the review says REVISE

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/.reviews/R006-plan-step2.md:1-4`, `taskplane-tasks/TP-093-compute-tools-cluster/STATUS.md:135-140`
- Severity: Medium

The added R006 review file says `Verdict: REVISE`, but `STATUS.md` logs `Review R006 | plan Step 2: APPROVE` and does not add R006 to the reviews table. Please keep the task status/review log consistent so future agents do not proceed from an incorrect approval state.
