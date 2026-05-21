# R006 plan review — Step 2: Implement aggregation logic

Verdict: REVISE

Step 2 should not proceed from the current plan/status as-is. The Step 2 checklist is directionally right, but it depends on a deterministic Step 1 contract, and the latest review artifacts still contain unresolved Step 1 blockers plus inconsistent status metadata.

## Blocking findings

### 1. Step 1 is recorded as approved, but the latest review file says REVISE

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/STATUS.md` Reviews table / Notes vs `.reviews/R005-code-step1.md`
- Severity: High

`STATUS.md` marks `R005 | Code | 1 | APPROVE`, but `.reviews/R005-code-step1.md` has `Verdict: REVISE` and lists three blocking findings. `R002` has the same kind of inconsistency (`STATUS.md` says APPROVE; `.reviews/R002-plan-step1.md` says “Changes requested”). This makes the current Step 2 plan untrustworthy because it appears to advance past a prerequisite that was not actually approved.

Before implementing aggregation logic, update the status/review log honestly and either resolve the R005 blockers in `CONTRACT.md` or add a subsequent approving review artifact.

### 2. Baseline aggregation semantics are still not deterministic enough to implement

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/CONTRACT.md`, `compute_baseline` calculation section
- Severity: High

The contract still says current value is the mean “unless the metric is a load total such as `weekly_tss`, in which case it is the sum.” That leaves many accepted `analysis_metric` values ambiguous: `weekly_hours`, training load, time/distance/calorie/elevation totals, steps, nutrition totals, and activity-row metrics can reasonably be treated as sums, means, per-day means, or per-activity means.

The Step 2 plan must first pin aggregation semantics by metric/source grain (and fallback behavior when the first source has insufficient baseline/current samples). Otherwise the planned baseline helper can produce a public API shape that is deterministic in code but not deterministic by contract.

### 3. Compliance event acquisition/truncation remains unspecified

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/CONTRACT.md`, `compute_compliance_rate` source priority
- Severity: High

Activity acquisition has a `max_activity_candidates = 500` partial-status rule, but scheduled event acquisition has no equivalent limit/cap/truncation contract. `compute_compliance_rate` denominators and per-sport/event-type breakdowns depend on the full scheduled-event set, so Step 2 needs a deterministic event listing rule before implementation: limit, cap behavior, whether paging is possible, and result/meta fields such as `truncated_event_candidates` with `status:"partial"`.

Without this, two implementations can compute different compliance rates over the same window depending on upstream defaults or chosen `ListEvents` limit.

### 4. Compliance mean-delta denominator is still undefined

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/CONTRACT.md`, `compute_compliance_rate` calculation/response
- Severity: Medium

The contract exposes `mean_delta_percent` and a metric-specific absolute mean delta, but it does not specify whether unpaired scheduled events are included in the delta denominator (for example as `-100%`) or excluded from deltas and counted only as non-compliant. This materially changes the headline “mean delta to target” result.

Pin the denominator for overall and breakdown deltas before coding, and expose a supporting count such as `delta_sample_count` if deltas are computed only over paired/completed activities.

### 5. Pace-zone source precedence conflicts with existing read output behavior

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/CONTRACT.md`, zone source priority; `internal/tools/get_extended_metrics.go`
- Severity: Medium

The contract says pace zones use `gap_zone_times` then `pace_zone_times`, while the existing `get_extended_metrics` shaping uses `gap_zone_times` only when raw `use_gap_zone_times` is true; otherwise it exposes `pace_zone_times`. Since this task is supposed to aggregate existing read outputs first, Step 2 should either mirror the existing selector or explicitly document the intentional divergence before implementing the shared zone helper.

## Required plan changes

1. Correct `STATUS.md` review outcomes/execution log so prior REVISE reviews are not shown as approvals.
2. Update `CONTRACT.md` for the unresolved R005 blockers: baseline aggregation by metric/source, event acquisition/truncation, and compliance delta denominators.
3. Add a Step 2 implementation note/checklist item to keep analyzer helpers aligned with the corrected contract, especially for pace-zone source selection and partial-status metadata.
4. Only then implement handlers/helpers; registration can still remain deferred to Step 3 as planned.

I did not run tests/build/lint because this is a plan review.
