# Plan Review — TP-087 Step 1: Design the enum and aliases

## Verdict: REVISE

The Step 1 record is now close to implementable and addresses the major concerns from R001–R003: it defines canonical snake_case values, conservative aliases, expression rejection, deterministic hint categories, and a reusable `internal/analysis` boundary with multi-source metadata.

I would still make a small revision before Step 2 because a few inventory/source details remain inconsistent with the stated “numeric/scalar read-tool fields” contract. These are easy to fix in the design record now, but would be harder to unwind after the enum is encoded in tests and future analyzer tools depend on it.

## Required revisions before implementation

1. **Remove or stage boolean wellness flags from the metric enum.**

   The wellness inventory includes `temp_weight` and `temp_resting_hr`, but the current typed wellness read fields are booleans (`tempWeight`, `tempRestingHR`), not numeric analysis series. They do not fit trend/distribution/correlation metric semantics in the same way as `weight`, `rhr`, `sleep_secs`, `feel`, etc.

   Please either:

   - exclude them from first-pass `analysis_metric`, or
   - explicitly classify them as boolean indicator metrics with analyzer support rules.

   Given the current plan says the enum is limited to numeric/scalar fields, exclusion/staging is the safer choice.

2. **Complete the multi-source descriptors for duplicated canonical metrics.**

   The metadata API now correctly allows multiple `MetricSource` entries, but the inventory still omits or blurs a few actual duplicate sources:

   - `ctl` / `atl` are emitted by both `get_fitness` and `get_wellness_data`; decide whether the wellness copies are source descriptors or intentionally ignored.
   - `feel` is emitted by wellness rows and by `get_extended_metrics` activity metrics; decide whether it is one canonical metric with multiple sources or wellness-only for v0.
   - `session_rpe`, `training_load`, time/distance/calorie totals, and HR fields already have duplicate-source handling noted; keep that same explicitness for the additional duplicates above.

   This is not asking for analyzer routing now, only for the source table target to be accurate enough that Step 2 does not freeze an incomplete contract.

3. **Separate extended activity metrics from extended interval metrics in the inventory.**

   The record currently labels the `get_extended_metrics` list as “extended per-activity scalar metrics” while including interval-only fields such as `dfa_alpha1`, `w_prime_balance_start_kj`, and `w_prime_balance_end_kj`. The final API section mentions `extended_activity` and `extended_interval`, which is good, but the inventory should mirror that split so implementers create the right `MetricSource.Grain` / `SourceFamily` entries.

   Suggested shape:

   - extended activity metrics: `stride_length_m`, `cardiac_decoupling_percent`, `pw_hr`, `aerobic_decoupling_percent`, zone arrays excluded, `joules_above_ftp_kj`, `if`, `vi`, `polarization_index`, `trimp`, `strain_score`, `hr_load`, `pace_load`, `power_load`, `training_load`, `left_right_balance_percent`, `rpe`, `feel`, `session_rpe`, `compliance_pct`, etc.
   - extended interval metrics: `dfa_alpha1`, `w_prime_balance_start_kj`, `w_prime_balance_end_kj`, plus the interval-overlap fields exposed on `extendedIntervalMetrics`.

## What looks good

- Unknown metric hints are concise, deterministic, and point toward the planned analyzer family instead of accepting arbitrary expressions.
- Free-form arithmetic examples are explicit enough to drive Step 3 table tests.
- `MetricSources(metric) []MetricSource` is the right direction for duplicate source families and avoids forcing each analyzer to duplicate the enum table.
- Deferring unit-contextual interval `pace` is the right call until analyzer unit normalization owns it.
- Staging PRD examples (`pace_at_lt2`, `power_at_lt2`, `np`) as v0.6 follow-ups is now documented rather than silently narrowing scope.

## Recommendation

Revise the Step 1 design record for the three source/inventory details above, then proceed to Step 2. The remaining implementation plan is sound: an SDK-free `internal/analysis` package with canonical values, private aliases, schema helpers, short invalid-argument errors, and metadata suitable for future analyzer routing.
