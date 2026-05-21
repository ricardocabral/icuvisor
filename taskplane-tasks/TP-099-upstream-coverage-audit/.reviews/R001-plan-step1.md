# Review R001 — Plan Review for Step 1

**Verdict:** Changes requested

Step 1 is still only the generic checklist in `STATUS.md`; it does not yet define a measurement method that Step 2 can implement reproducibly.

## Blocking issues

1. **Fixture corpus is not identified.**
   The plan must state the exact eligible fixture set and denominator. Scanning every JSON file under `internal/intervals/testdata` and `internal/tools/testdata` would mix activity fixtures with wellness/events/workout-library/analyzer golden files and inflate `fixture_count`, `fallback_count`, or `unknown_count`. Define either explicit globs or a type-detection rule for activity/summary/extended-metrics-like fixtures, and say what is excluded.

2. **Precomputed-zone criteria are not defined.**
   Step 1 should list the fields that count as precomputed zone time, aligned with the current implementation:
   - training summary: `timeInZones` with `timeInZonesTot > 0`
   - power: `icu_zone_times`, `power_zone_distribution_seconds`, `power_zone_times`
   - pace: `gap_zone_times`, `pace_zone_times`, `pace_zone_time_seconds`
   - heart rate: `hr_zone_times`, `heartrate_zone_times`, `heart_rate_zone_times`, `hr_time_in_zones`

   Also specify that a field only counts when it is a non-empty numeric array with positive total seconds.

3. **Metric definitions need precise denominators.**
   Define whether counts are per JSON file, per activity object, per summary row, or per `(activity, zone_metric)` opportunity. Without this, `precomputed_count`, `fallback_count`, and `unknown_count` are ambiguous. Given the tools operate by zone family, reporting by metric family (`power`, `heart_rate`, `pace`) plus totals would be more useful than a single blended count.

4. **“Fallback” must be reconciled with current behavior.**
   `compute_zone_time`/`compute_load_balance` currently avoid raw stream reduction and return unavailable/partial when precomputed zones are missing. The plan should define `fallback_count` as “would require stream math / missing precomputed zones” unless the task intentionally asks for a hypothetical fallback measurement. Do not imply the current tools actually call stream endpoints.

5. **Threshold is absent.**
   ROADMAP/PRD do not appear to provide an agreed threshold. Step 1 must either record a concrete threshold for documenting an upstream gap or explicitly mark it as an operator decision before Step 2 runs. Avoid choosing a threshold after seeing results.

## Recommended Step 1 output

Update `STATUS.md` (Notes or Step 1 section) with:

- eligible fixture corpus and exclusions;
- exact precomputed field list and validation rules;
- count definitions for `fixture_count`, `precomputed_count`, `fallback_count`, and `unknown_count`;
- grouping dimensions, at least by source path/type and zone metric;
- threshold decision or “operator decision: no threshold agreed”.

Once those are recorded, the plan should be acceptable for the small audit script/helper in Step 2.
