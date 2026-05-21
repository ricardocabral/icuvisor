# R013 code review — Step 1: Design deterministic contracts

Verdict: REVISE

I reviewed the Step 1 contract/status changes against the current tree and previous Step 1 review blockers. The prior R009 items are mostly addressed, but two contract edges are still not deterministic enough to hand to implementation/registration.

## Blocking findings

### 1. `compute_baseline` does not define cross-window ordering for activity acquisition

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/CONTRACT.md:17`, `:63`, `:82`
- Severity: High

The baseline contract now says activity-backed sources use “the union from `baseline_start_date` through `current_end_date`”, but the request schema only rejects inverted individual windows. It does not require the baseline window to precede the current window, and it does not define what happens when the two windows overlap or the current window is earlier than the baseline window.

For a valid-by-schema request such as baseline `2026-05-01..2026-05-31` and current `2026-04-01..2026-04-07`, “baseline_start through current_end” is itself inverted. Implementations can reasonably reject the request, query `min(start)..max(end)`, or query the two windows separately, which changes activity cap/truncation behavior and sample membership.

Please pin one rule before treating the contract as deterministic: either require `baseline_end_date < current_start_date` (or `<=` if touching is allowed), or explicitly define the union as `min(baseline_start_date,current_start_date)..max(baseline_end_date,current_end_date)` with baseline/current membership determined by the two original windows and a clear overlap policy.

### 2. Compliance filtering/pairing is still ambiguous when `sport` and `event_type` are both supplied

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/CONTRACT.md:95`, `:99`, `:101`
- Severity: Medium

Line 95 states that, when both filters are present, `event_type` alone controls scheduled-event type matching and `sport` controls completed-activity candidates. But the acquisition and pairing rules still say scheduled rows match “sport/type filters” and auto-pairing is by same local date and “sport/type”. That leaves the auto-pair key undefined for requests where the scheduled `Event.Type` selected by `event_type` differs from the requested completed-activity `sport`.

Two implementations can therefore diverge: one can pair any `sport`-filtered activity to an `event_type`-filtered event on the same date, while another can require the activity sport to equal the event type and mark the event unpaired. Please state the exact auto-pair comparison when both filters are supplied (for example, compare activity sport to `sport` when provided, otherwise to `Event.Type`; or require both and document that result).

### 3. Auto-lap interval-evidence trigger is not pinned

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/CONTRACT.md:102`, `:109`
- Severity: Medium

The contract says to call `get_activity_intervals` “when interval-target compliance requires interval execution evidence,” but the request schema and calculation section only define aggregate target metrics (`time`, `distance`, `load`). It never defines which scheduled events require interval evidence, whether that is based on `WorkoutDoc` presence, interval targets inside the workout document, event category/type, or user input. It also does not define whether a missing/unavailable interval response is ignored, marks a boundary, or changes status.

Because this changes `source_tools`, `_meta.auto_lap_suspected`, `auto_lap_caution`, and potentially returned errors, please add a deterministic trigger and failure policy for interval evidence before Step 2/3 depend on this contract.

## Non-blocking notes

- `STATUS.md:143-145` still says `Blockers: None` while the same file records R010/R012 as `REVISE`, and the current tree still has a generated-doc test failure. That is not a Step 1 contract blocker by itself, but it should be corrected before advancing later steps.

## Verification performed

- `go test ./internal/tools ./internal/toolcatalog ./internal/toolchecks ./cmd/gendocs` — fails in `cmd/gendocs` with `generated catalog differs from golden` because the committed gendocs golden does not include the four registered compute tools. This is a Step 3/doc freshness issue, not one of the Step 1 contract findings above.
