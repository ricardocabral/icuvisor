# R006 Code Review — Step 1: Define histogram contract

**Verdict:** REQUEST CHANGES

The R005 boundary/name-pair ambiguity is fixed. I found one remaining blocking contract gap around pace-zone unit conversion. This needs to be locked before Step 2 because the repository currently has two plausible pace-unit interpretations, and choosing the wrong one will shift every pace-zone boundary.

## Blocking finding

1. **Pace zone conversion does not define exact `SportSettings.PaceUnits` semantics.**
   - Location: `taskplane-tasks/TP-092-activity-histogram/STATUS.md:135`
   - The contract says configured pace zones are converted from `SportSettings.PaceUnits` into the emitted unit, but it does not define the conversion table or whether upstream `MINS_KM`/`MINS_MILE` values are stored as minutes-per-distance or seconds-per-distance.
   - This is ambiguous in this repo today: `response.ToPreferredWithRaw` treats `MINS_KM`/`MINS_MILE` as minute pace units (`internal/response/units.go:191-202`), while the sport-settings write path converts `seconds_per_km` input into upstream `MINS_KM` by writing seconds per km (`internal/tools/update_sport_settings.go:292-300`). Either helper could look reasonable to a Step 2 implementation, but only one can be correct for histogram bucket edges.
   - Please update the Step 1 contract with an explicit pace-zone conversion table for at least `MINS_KM`, `MINS_MILE`, `SECS_100M`, and `SECS_500M`, including the exact factors to output `seconds_per_km` and `seconds_per_mile`. Also state the fallback behavior for unknown/empty `PaceUnits` when pace zones are present: ignore zones and use fixed-width, or treat values as already in the emitted unit.

## Non-blocking notes

- Fixed-width bucket `label` text remains undefined even though `label` is a required bucket field. Ranges are enough for math, but adding a deterministic label rule now would reduce golden-test churn.
- The execution-log rows are still appended under `## Notes` rather than in the `## Execution Log` table (`STATUS.md:146-150`). This is STATUS hygiene only.

Tests were not run; this step only changes task/status documentation.
