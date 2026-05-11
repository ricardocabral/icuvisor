# Code Review: TP-004 Step 1 — tool contract

## Verdict

**Needs changes before Step 2.** The contract is much improved and covers the key safety boundaries, but two parts are still too ambiguous for implementation/tests: pace units and the `include_full` response shape.

## Findings

### 1. Pace fields are unit-ambiguous and internally inconsistent

- `STATUS.md:63-65` documents `threshold_pace: 4.2`, `pace_units: "min/km"`, and decimal `pace_zones` values.
- `STATUS.md:73` says these fields use “the intervals.icu pace value plus pace_units.” Existing typed client fixtures use upstream-shaped values such as `threshold_pace: 255.5` with `pace_units: "MINS_KM"`, which appears to be a different representation than decimal minutes.

Because Step 1 is meant to pin the response contract, Step 2 currently has no clear answer for whether to return upstream seconds, decimal minutes, formatted pace, or converted per-athlete units. This also weakens the PRD requirement that units be disambiguated in keys or metadata.

**Required adjustment:** choose one representation and document it explicitly. For example:

- normalized fields: `threshold_pace_seconds_per_km` / `pace_zones_seconds_per_km`, with `pace_units_source: "MINS_KM"`; or
- display fields: `threshold_pace_min_per_km` / `pace_zones_min_per_km`, with conversion from upstream seconds documented.

Also document how miles vs km is selected when the athlete uses imperial units.

### 2. `include_full: true` additions are not a concrete contract

- `STATUS.md:32` says full mode includes “additional non-secret profile and sport-setting fields.”
- `STATUS.md:79` says additions include fields “such as” sport-setting `id` and normalized `athlete_id`.

Since Step 1 opted into `include_full`, the response shape needs to name the exact additional fields that Step 2 will implement and Step 4 will test. Leaving this as “such as” lets the implementation drift and makes default/full behavior hard to assert.

**Required adjustment:** add a concrete `include_full: true` response delta, e.g. exact field names like `sport_setting_id` and `sport_setting_athlete_id`, plus any top-level profile fields if intended. Keep the current explicit ban on raw upstream JSON, secrets, headers, and fetched timestamps.

## Minor notes

- `STATUS.md:11-16` marks every Step 1 checkbox complete while the step status remains “In Progress.” If the above contract clarifications are applied, mark Step 1 consistently as complete/ready for Step 2.
- `STATUS.md:128-131` has a blank line between the Markdown table header and rows, so the discoveries may not render as one table on GitHub.
