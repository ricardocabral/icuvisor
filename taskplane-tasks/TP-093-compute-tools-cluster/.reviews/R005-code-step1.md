# R005 code review — Step 1: Design deterministic contracts

Verdict: REVISE

## Blocking findings

### 1. `compute_baseline` still leaves current-window aggregation ambiguous

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/CONTRACT.md:70-82`
- Severity: High

The contract now accepts the full `analysis_metric` enum and lists supported source families, but the calculation still says the current value is the mean “unless the metric is a load total such as `weekly_tss`, in which case it is the sum.” That is not deterministic enough for Step 2: accepted metrics include other totals/window metrics (`weekly_hours`, training load, time, distance, calories, elevation, steps, nutrition totals, activity-row totals) where implementations could reasonably choose sum, mean-per-day, mean-per-activity, or a weekly bucket value.

Please pin aggregation semantics for every accepted metric/source family (or at least a deterministic rule keyed by `MetricSource.Grain`/metric class): how baseline samples are formed, whether the current window is mean/sum/weighted mean, and exactly which metrics are “load totals.” Also specify whether the implementation falls through to the next `analysis.MetricSources` entry when the first source exists but has insufficient baseline/current samples.

### 2. `compute_compliance_rate` does not define event enumeration/truncation behavior

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/CONTRACT.md:99`, `:109`
- Severity: High

Activity acquisition has a deterministic `max_activity_candidates = 500` and partial/truncation contract, but scheduled event acquisition has no equivalent. The existing `get_events` tool has capped list behavior (default 100, max 500, truncated meta), while `EventsClient.ListEvents` can be called directly with an unspecified `Limit`. As written, two implementations can compute different denominators depending on upstream defaults or chosen limits for a busy calendar window.

Please add a deterministic event acquisition rule: limit/cap, whether the tool pages or rejects large windows, what happens when the event candidate set reaches the cap, and result/meta fields such as `truncated_event_candidates` plus `status:"partial"`/boundary text. This is especially important because `scheduled_count`, `compliance_rate`, and the per-sport/event-type breakdowns all depend on the complete scheduled-event set.

### 3. Compliance mean-delta denominator is undefined for unpaired events

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/CONTRACT.md:107-109`
- Severity: Medium

The contract says scheduled target events with no completion count as non-compliant, and it exposes default aggregate `mean_delta_percent` / absolute mean delta fields. It does not say whether unpaired events contribute to those mean deltas (for example as `-100%`, as `null`, or are excluded from the delta denominator and counted only in compliance). That choice materially changes the headline “mean delta to target” promised by the PRD.

Please pin the delta denominator for both overall and `by_sport`/`by_event_type` rows, and expose any supporting counts needed to interpret it (for example `delta_sample_count` if deltas are calculated only over paired completed activities).

## Non-blocking findings

### 4. Pace zone source precedence conflicts with the existing `get_extended_metrics` shape

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/CONTRACT.md:31`; current behavior `internal/tools/get_extended_metrics.go:220-224`
- Severity: Medium

The contract says pace zones use `gap_zone_times` then `pace_zone_times`. The current `get_extended_metrics` output uses `gap_zone_times` only when raw `use_gap_zone_times` is true; otherwise it exposes `pace_zone_times`. Since this task is supposed to aggregate existing read outputs first, the compute contract should either honor the same `use_gap_zone_times` selector or explicitly document that the analyzer intentionally diverges from `get_extended_metrics` for pace zones.

### 5. Status review log is inconsistent with the review files

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/STATUS.md:92-99`, `:126-134`; `.reviews/R002-plan-step1.md:3`
- Severity: Low

`STATUS.md` records R002 as `APPROVE`, but the R002 file itself says “Verdict: Changes requested.” The timestamp rows are also appended under `## Notes` rather than the `## Execution Log` table. Please keep the review/status metadata consistent so later reviewers can tell which revisions were actually blocking.

## Notes

I did not run tests/build/lint because this step only changes task contract/status documentation.
