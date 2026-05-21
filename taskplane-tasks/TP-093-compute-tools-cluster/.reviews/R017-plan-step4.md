# R017 plan review — Step 4: Tests and verification

Verdict: REVISE

The Step 4 plan points at the right broad areas, but it is too thin for the size of the public contract being locked in. The tests added in this step need to exercise the deterministic edge cases that were introduced or fixed in prior reviews, not just happy-path goldens.

## Blocking findings

### 1. The plan does not cover truncation/partial-status semantics

- Location: `STATUS.md` Step 4 checklist
- Severity: High

The contract requires any activity- or event-backed path that hits the 500-candidate cap to return `status:"partial"`, set the corresponding truncation flag, add `_meta.assumptions.*_truncated:true`, and include a boundary. This was also a blocking issue in R015.

The current Step 4 plan has no explicit tests for:

- `compute_zone_time` / `compute_load_balance` activity-backed truncation, including the no-usable-zone-at-cap case.
- `compute_baseline` activity-backed truncation taking precedence over insufficient baseline/current samples.
- `compute_compliance_rate` activity and event truncation metadata/status.

Please add truncation goldens or focused assertions before Step 4 is considered complete.

### 2. The plan misses regression coverage for source priority and no-stream guarantees

- Location: `STATUS.md` Step 4 checklist; `CONTRACT.md` source-priority sections
- Severity: High

The plan says “precomputed aggregation” and “no-precomputed fallback behavior,” but it does not pin the source-priority cases that make these tools deterministic:

- Unfiltered `power` zones prefer `get_training_summary` and do not enumerate activities when summary zones are usable.
- Sport-filtered or non-power zone queries avoid whole-day summary zones unless deterministic, and use activity/extended precomputed arrays.
- Missing precomputed arrays return unavailable/partial with boundaries instead of raw-stream math.
- `compute_load_balance.training_load_total` uses the documented load priority (`power_load`, `hr_load`, `pace_load`, `icu_training_load`, typed/list-row load), independent of requested zone metric.

These are contract-critical and include a prior R015 regression. The Step 4 plan should explicitly include tests with fake clients that record calls and fail if a raw-stream fallback or wrong source path is used.

### 3. Baseline coverage is narrower than the contract

- Location: `STATUS.md` Step 4 checklist; `CONTRACT.md` `compute_baseline`
- Severity: Medium

Baseline tests are currently planned only for insufficient samples and current-window missing data. That leaves important public statuses and calculations untested:

- Zero-variance baseline => `status:"insufficient_variance"`, null z-score, `insufficient_reason:"zero_baseline_variance"`.
- Cross-window ordering validation (`baseline_end_date < current_start_date`).
- At least one successful z-score golden with `_meta.formula_ref`, `n`, missing-day counts, terse/full shaping, and interpretation direction for a wellness metric.
- Weekly/activity source-grain behavior if those paths were implemented as part of Step 2.

Please broaden the baseline plan or explicitly justify which non-covered contract cases are deferred.

### 4. Compliance tests need to include deterministic pairing edge cases, not only one pairing happy path

- Location: `STATUS.md` Step 4 checklist; `CONTRACT.md` `compute_compliance_rate`
- Severity: Medium

“Scheduled/completed pairing and auto-lap/interval caution” is not specific enough to protect the compliance contract. Add explicit cases for:

- Linked pairs reserve activities before auto-pairing, and linked conflicts do not reuse activities.
- Targetless events are excluded and do not reserve linked activities.
- When both `sport` and `event_type` are supplied, scheduled filtering uses `event_type` while auto-pairing compares activity sport to request `sport`.
- Interval evidence unavailable adds a series caution and `_meta.boundaries`, while auto-lap suspected sets `auto_lap_caution` and `_meta.auto_lap_suspected`.
- Mean delta denominators use only completed paired activities and are exposed in overall and breakdown rows.

These are not optional polish; they are core deterministic behavior from the contract and previous review fixes.

### 5. Verification commands are underspecified and weaker than the task checklist

- Location: `PROMPT.md` Step 4; `STATUS.md` Step 4 checklist
- Severity: Medium

The prompt's Step 4 says to run the full quality gate, while the status plan only says “targeted compute-tool quality gate.” Step 5 later repeats the full-suite/build/lint gate, so it is fine to defer the final full gate there, but the Step 4 plan should still name the targeted commands that cover all affected surfaces.

At minimum, Step 4 should run and record:

```sh
go test ./internal/analysis ./internal/tools ./internal/toolcatalog ./internal/toolchecks ./internal/safety ./cmd/gendocs
```

If docs or generated catalog files changed, also include the relevant generation/check command used by this repo. If the full gate is intentionally deferred to Step 5, say that explicitly in `STATUS.md` so Step 4 is not claiming the original checklist item prematurely.

## Suggested plan adjustment

Replace the current four bullets with a more explicit checklist grouped by tool and verification surface: zone/load source-priority + truncation + meta goldens, baseline status/calculation goldens, compliance pairing/interval/truncation goldens, catalog/docs freshness checks, and named targeted test commands. That will make Step 4 reviewable and prevent the prior Step 1/2 regressions from reappearing.
