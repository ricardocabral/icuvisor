# R003 code review — Step 1: Design deterministic contracts

Verdict: REVISE

## Findings

### 1. Summary-backed zone data cannot satisfy the promised sport-filtered zone contract

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/CONTRACT.md:12`, `:26`, `:43`
- Severity: High

The contract says summary-backed tools should sport-filter through `SummaryWithCats.byCategory`, while `compute_zone_time` / `compute_load_balance` should prefer `SummaryWithCats.TimeInZones` / `TimeInZonesTot`. In the current model, `SummaryWithCats.TimeInZones` is a whole-day array and `CategorySummary` has only volume/load totals, not zone arrays (`internal/intervals/fitness.go:35-37`, `:56-70`). That means a mixed-sport day cannot produce sport-filtered zone seconds from the summary row.

As written, an implementation could incorrectly return the whole-day zone distribution for `sport:"Run"` just because a Run category exists, or mark the date missing even though activity-level precomputed zone arrays are available. Please pin the deterministic rule, e.g. summary zones may be used only for unfiltered windows (or for dates proven homogeneous), and sport-filtered zone/load-balance queries must fall back to per-activity precomputed arrays or explicitly return partial/unavailable for those rows.

This also intersects with `zone_metric`: the summary source is documented as metric-agnostic but is still authoritative for requested `power`, `heart_rate`, and `pace`. Returning the same summary zones for all three metrics is likely misleading unless the contract narrows when summary zones are valid for a metric.

### 2. `compute_zone_time` omits a PRD-required polarization output

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/CONTRACT.md:32`; PRD reference `docs/prd/PRD-icuvisor.md:279`
- Severity: High

The PRD says `compute_zone_time` returns time per power/HR/pace zone over a sport-filtered window "with polarization index." The new contract's `compute_zone_time` response omits `polarization_index`, `polarization_state`, and the `icuvisor://analysis-formulas#polarization_index` formula ref, leaving those only on `compute_load_balance`.

Because the PRD is authoritative for product behavior, implementing this contract would ship a tool that does not meet the documented catalog entry. Either add the polarization fields/meta to `compute_zone_time` (possibly sharing the load-balance calculation) or explicitly update the PRD/roadmap in the same step if the intended scope has changed.

### 3. `compute_baseline` accepts the closed metric enum but only defines behavior for a subset

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/CONTRACT.md:62-69`, `:73-75`
- Severity: High

The request schema accepts the closed `analysis_metric` enum and aliases, but the source-priority section only defines broad handling for a few families/examples (`ctl`, `atl`, `tsb`, `weekly_tss`, wellness sleep/HRV, `if`, `vi`). The actual enum contains many more accepted metrics/families (for example `weekly_hours`, `rhr`, subjective scales such as `feel`, activity row metrics, extended load metrics, `compliance_pct`, etc.). For those accepted inputs, the contract does not say which source wins, whether they are unsupported for baseline, or what user-facing status/reason should be returned.

Please either narrow the schema enum for `compute_baseline` to the metrics this tool will support, or define the deterministic source/status rule for every accepted `analysis.MetricValues()` entry (preferably via `analysis.MetricSources`). Also specify the status when the current window has zero usable samples; the contract defines insufficient baseline samples and zero baseline variance, but leaves `current_value`/`z_score` undefined when `n_current == 0`.

### 4. Compliance auto-pairing needs a one-to-one matching rule

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/CONTRACT.md:88`, `:94-96`
- Severity: Medium

The contract says each scheduled event pairs at most one completed activity, but it does not state whether one activity may be reused across multiple scheduled events. With two same-day/same-sport scheduled workouts and one completed activity, the current wording allows both events to choose the same nearest activity, inflating `completed_count`, `compliant_count`, and `compliance_rate`.

Please pin a deterministic one-to-one policy (for example: process events by date/start/id, linked pairs reserve activities first, auto-pairing removes the chosen activity from the candidate set; tie-break by percent difference then stable activity ID), or explicitly document that reuse is allowed and how it is surfaced. One-to-one matching seems more consistent with scheduled-vs-completed compliance.

## Notes

- I did not run tests/build/lint because this step only adds a contract note and updates task status.
- The status file still marks Step 1 as in progress; that may be intentional while awaiting review, but it should be moved forward only after the contract issues above are resolved.
