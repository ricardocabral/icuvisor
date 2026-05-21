# R015 code review — Step 2: Implement aggregation logic

Verdict: REVISE

I reviewed the Step 2 aggregation implementation in the current tree. The targeted compute packages compile, but the full test suite currently fails, and several aggregation/status semantics still diverge from the Step 1 contract.

## Blocking findings

### 1. Full test suite fails after registering the compute tools

- Location: `internal/tools/catalog.go:106-109`, `internal/safety/adversarial_test.go:23-66`, `internal/safety/adversarial_test.go:111-112`
- Severity: High

`go test ./...` fails in `internal/safety` because the static adversarial catalog matrix was not updated for the four newly registered read tools. The registry now includes `compute_baseline`, `compute_compliance_rate`, `compute_load_balance`, and `compute_zone_time`, but `v03ToolCatalog` still enumerates the previous catalog, so the expected registered counts are off by four in every mode.

Reproduction:

```sh
go test ./...
```

Result:

```text
registered tool count in mode safe = 41, want 37
registered tool count in mode none = 31, want 27
registered tool count in mode full = 48, want 44
```

Please update the safety adversarial catalog expectations if these tools are intended to be registered in this change, or defer the registration until the Step 3/catalog surfaces are complete.

### 2. Activity truncation still does not reliably force `status:"partial"`

- Location: `internal/tools/compute_baseline.go:118-124`, `internal/tools/compute_zone_time.go:223-224`, `internal/tools/compute_zone_time.go:333-339`
- Severity: High

The shared contract says any activity-backed compute path that reaches `max_activity_candidates` must return `status:"partial"`, set the truncation flag/assumption, and include a boundary because aggregates may exclude rows past the cap.

Two paths still let other statuses win over truncation:

- `compute_baseline` only changes the status to `partial` when the baseline stats were otherwise `ok`. A capped activity list with too few visible baseline/current samples can still return `insufficient_sample` or `insufficient_current_sample`, even though rows beyond the cap may contain the missing samples.
- `compute_zone_time` records `agg.Truncated`, but `aggregateStatus` returns `unavailable` before checking truncation when the first 500 candidates have no usable precomputed zones. That hides the fact that the result is capped and incomplete.

This is the exact ambiguity the truncation contract is meant to avoid. Truncation should have status precedence (or at least produce a partial status with the insufficiency reason preserved) for all activity-backed compute tools.

### 3. `compute_load_balance.training_load_total` still ignores the documented load-source priority

- Location: `internal/tools/compute_zone_time.go:255-260`, `internal/tools/compute_zone_time.go:275-289`
- Severity: Medium

The contract says each contributing activity row should add the first available load in priority `power_load`, `hr_load`, `pace_load`, `icu_training_load`, then the typed/list-row training load. The implementation instead changes the priority by requested zone metric:

- `heart_rate` checks only `hr_load`, then `icu_training_load`;
- `pace` checks only `pace_load`, then `icu_training_load`;
- `power` checks only `power_load`, then `icu_training_load`.

So, for example, a heart-rate zone row with only `power_load` available contributes zone seconds but contributes no contextual load, even though `power_load` is the first documented fallback. This makes `training_load_total` inconsistent with the deterministic contract and with the status note that the R010 load-priority issue was fixed.

### 4. Compliance interval-evidence failures are swallowed without the required caution/boundary

- Location: `internal/tools/compute_compliance_rate.go:259-269`, `internal/tools/compute_compliance_rate.go:455-466`
- Severity: Medium

The contract says that when a paired scheduled event has a non-empty `workout_doc`, `get_activity_intervals` is used; unavailable non-cancellation interval responses should not fail aggregate compliance, but they must add a series caution and a `_meta.boundaries` entry saying interval execution could not be verified.

`complianceIntervalEvidence` currently converts any non-context interval error into a nil error and an empty `analysis.IntervalSourceResult`. The caller then marks interval evidence as used, but it does not set any row caution or add any boundary for the unverifiable interval evidence unless auto-lap is detected. If `intervalsClient` is nil, the same required verification boundary is also absent.

Please preserve the non-fatal unavailable state separately from successful `unknown` evidence, add the contracted boundary, and include a row caution reason so users can distinguish “checked and unknown” from “could not verify intervals”.

### 5. Targetless compliance events can reserve linked activities before being excluded

- Location: `internal/tools/compute_compliance_rate.go:214-220`, `internal/tools/compute_compliance_rate.go:390-404`
- Severity: Medium

`buildComplianceReservations` reserves linked activities for every filtered scheduled event before the main loop checks whether the event has the selected target metric. The main loop then excludes targetless events from the denominator, but their reserved activity remains unavailable for later target events and auto-pairing.

That violates the contract’s “scheduled target event” denominator/pairing semantics: events without a selected target are excluded and should not consume a completed activity. A targetless linked event can currently cause a valid target event on the same day to become unpaired, lowering `completed_count`, `compliant_count`, and `compliance_rate`.

Filter reservations to events that have the selected target metric, or release reservations for excluded events before auto-pairing.

## Verification performed

- `go test ./internal/analysis ./internal/tools ./internal/toolcatalog ./cmd/gendocs` — passes.
- `go test ./...` — fails in `internal/safety` due to stale adversarial catalog expectations, as described above.
