# R004 code review — Step 1: Design deterministic contracts

Verdict: REVISE

## Blocking findings

### 1. `compute_baseline` still lacks the PRD-required wellness interpretation

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/CONTRACT.md:78-80`
- Severity: High

The baseline response still only defines the z-score and sample statistics. PRD §7.2.C requires `compute_baseline` to return a z-score plus a `suppressed` / `elevated` flag for wellness metrics. Without a contracted field and deterministic thresholds/directionality, Step 2 can implement a public shape that is missing a catalog promise, or different metrics can be interpreted inconsistently.

Please add a response field such as `interpretation`/`state` and pin the exact behavior: which metrics are wellness/interpretable, which direction is adverse or beneficial (`hrv` low = suppressed, `rhr` high = elevated, subjective scales, sleep, etc.), the z-score thresholds, and what is emitted for non-directional metrics. Record these assumptions in `_meta.assumptions`.

### 2. `compute_compliance_rate` still does not satisfy the “mean delta, per sport / event type” contract

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/CONTRACT.md:87`, `:91`, `:99-101`; PRD reference `docs/prd/PRD-icuvisor.md:283`
- Severity: High

The request schema exposes `sport` and `category`, but no event-type filter or grouping even though the source-priority text mentions “category/type filters”. The terse result also lacks aggregate delta fields; percent difference is only described for full `series` rows. The PRD explicitly promises “mean delta to target, per sport / event type”, so the default response must include aggregate delta(s), not require `include_full:true` and manual reduction.

Please pin the public shape for this before implementation: e.g. optional `event_type`, grouping keys (`by_sport`, `by_event_type` or a combined breakdown), and terse aggregate fields such as `mean_delta_percent` plus metric-specific absolute deltas (`mean_delta_seconds`, `mean_delta_meters`, or `mean_delta_load`).

### 3. Compliance matching still leaves link and time-target behavior ambiguous

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/CONTRACT.md:92-93`, `:99`
- Severity: Medium

The revised contract adds one-to-one matching, but it still allows implementation variance in two places. First, link detection still includes “or equivalent ID fields”, which makes the accepted raw keys open-ended. Second, for `target_metric:"time"`, the calculation says “actual moving/elapsed seconds” without defining whether `time_target` maps to `moving_time`, `elapsed_time_target` maps to `elapsed_time`, and what happens when both targets are present.

Please enumerate the exact raw keys accepted for links (or explicitly say only typed `paired_event_id` / named raw keys are supported) and define target precedence plus actual-field selection. Otherwise two implementations can compute different completions and deltas for the same event/activity payloads.

### 4. Activity enumeration and truncation are under-specified for tools that depend on `get_activities`

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/CONTRACT.md:27`, `:70-71`, `:93`
- Severity: Medium

Several contracts depend on enumerating activities (`compute_zone_time`, `compute_baseline` for activity/extended metrics, and compliance auto-pairing), but the contract does not define the listing limit, whether the implementation must page until exhaustion, what fields are requested, or what status/meta is emitted if the activity list is truncated. Since `get_activities` is paginated and activity order affects auto-pairing candidates, this can change aggregate results and `_meta.n` across implementations.

Please add a deterministic acquisition rule: the date range used, limit/page behavior or maximum cap, `source_tools` inclusion, and a `partial`/`truncated` status or boundary if the tool cannot exhaust the candidate set.

## Non-blocking notes

- `CONTRACT.md:26` uses `_meta.assumptions.zone_metric_source` as the condition for trusting summary zones for non-power metrics. Because `_meta` is output metadata, this should instead be tied to a concrete upstream field/source marker; otherwise an implementation can self-declare metric-agnostic summary zones and return the same array for HR/pace.
- The status notes append review-log table rows under `## Notes` rather than the `## Reviews` table. This is not a product blocker, but it makes the task status harder to scan.

I did not run tests/build/lint because this step only changes task contract/status documentation.
