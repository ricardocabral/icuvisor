# R009 code review — Step 1: Design deterministic contracts

Verdict: REVISE

`go test ./...` passes. I reviewed Step 1's contract/status changes and treated the Step 2 implementation files as out of scope except where they exposed contract ambiguity.

## Blocking findings

### 1. `summary_window` baseline sampling is internally inconsistent and still not deterministic

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/CONTRACT.md:70-82`
- Severity: High

The baseline contract now pins most grains, but the `training_summary` / `summary_window` path conflicts with itself. The source list describes `get_training_summary` as “daily summary fields,” while the calculation says `summary_window` uses “one value per requested window” and still requires `min_samples` window rows “only when the source can produce daily/window rows.” For metrics whose source catalog points at `GrainSummaryWindow` (for example time/distance/elevation summary-backed metrics), two implementations can reasonably do different things:

- call `get_training_summary` once for the whole baseline/current window and return `n=1`/insufficient;
- call the lower-level daily summary client and treat each day as a daily sample;
- split the baseline into daily/weekly sub-windows and synthesize multiple window rows.

Please pin the acquisition and sample rule for `summary_window` explicitly: exact client/tool call shape, sub-window size if any, how `n_baseline`/`n_current` are counted, and how `min_samples` applies. If these are actually daily `SummaryWithCats` rows, mark the selected grain/method as daily rather than leaving it as `summary_window`.

### 2. Activity-backed baseline acquisition does not define the two-window cap/truncation contract

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/CONTRACT.md:17`, `:82-84`
- Severity: High

The shared activity acquisition rule says to list activities over “the same inclusive local date window” with `max_activity_candidates = 500`, but `compute_baseline` has two independent windows: baseline and current. The contract does not say whether activity/extended baseline sources should query the union, query each window separately, apply the 500 cap per window or across the union, or how to report truncation separately for baseline vs current. A large baseline window can therefore change `n_baseline`, `n_current`, `current_value`, and `z_score` depending on the implementation strategy.

Please define deterministic activity acquisition for `compute_baseline`: baseline and current listing windows, ordering, cap scope, and result/meta fields for `truncated_baseline_activity_candidates` vs `truncated_current_activity_candidates` (or an intentionally combined flag with clear semantics).

### 3. `compute_load_balance.training_load_total` lacks aggregation semantics

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/CONTRACT.md:47-55`
- Severity: Medium

The result includes `training_load_total`, and source priority lists several load fields, but the contract never states how that value is formed. It is unclear whether the load is summed across the same rows that contributed zones, whether source priority is selected per row or globally for the window, whether summary `TrainingLoad` may be mixed with activity raw loads, and what happens when zone rows exist but load fields are absent.

Please pin `training_load_total` as a deterministic aggregate, for example “sum the first available load field per contributing zone row using this priority; do not mix summary and activity sources in the same response unless recorded in `series`; omit/null when no load source exists.” Without this, registered clients can see the same zone classification with different contextual load totals.

### 4. Compliance sport/event-type filtering remains ambiguous when both filters are supplied

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/CONTRACT.md:93-101`
- Severity: Medium

The contract says `event_type` filters `Event.Type`, while `sport` also filters scheduled event type/category “unless `event_type` is supplied more specifically.” It later says scheduled rows match both `event_type` and sport/type filters. That leaves the behavior undefined when both are present: does `event_type` override the scheduled-event side of `sport`, must both match, or does `sport` only filter completed activities? The answer changes `scheduled_count`, breakdowns, and auto-pair candidates.

Please state a single precedence rule for scheduled event filtering and activity filtering when `sport` and `event_type` are both supplied.

## Non-blocking notes

- `STATUS.md:46-54` marks Step 2 complete while the task still says the current step is Step 1 and the review table records R006/R007 as `REVISE`. Since this review is for Step 1, either move the Step 2 status back behind the unresolved Step 2 review or add a superseding approving review before treating it as complete.
- `STATUS.md:139-145` still has review-log rows under `## Notes` that say R002/R005/R006/R007 were approved, contradicting the Reviews table and the review files. Please keep the status metadata consistent for the next reviewer.
