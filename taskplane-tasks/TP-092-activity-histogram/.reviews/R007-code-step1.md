# R007 Code Review — Step 1: Define histogram contract

**Verdict:** REQUEST CHANGES

The R006 pace-zone conversion table and fixed-width label text are added. I found two remaining contract gaps that can still lead to incompatible Step 2 implementations and golden tests.

## Blocking findings

1. **Pace stream unit conversion does not define the stream distance/time base units.**
   - Location: `taskplane-tasks/TP-092-activity-histogram/STATUS.md:143`
   - The contract says pace interval `i` uses `(time[i+1]-time[i]) / (distance[i+1]-distance[i])` and is converted to `seconds_per_km` or `seconds_per_mile`, but it does not state what units the canonical `distance` and `time` stream samples are in.
   - Existing stream-derived split code treats `distance` as meters and `time` as seconds (`internal/tools/get_activity_streams.go:287-326`). If Step 2 assumes kilometers, miles, or timestamps for those arrays, every pace histogram bucket will be wrong while still appearing to satisfy the written contract.
   - Please lock the formula explicitly, e.g. `dt_seconds := time[i+1]-time[i]`, `dd_meters := distance[i+1]-distance[i]`, `seconds_per_km = dt_seconds / (dd_meters / 1000)`, and `seconds_per_mile = dt_seconds / (dd_meters / 1609.344)`.

2. **Sport-setting selection is still ambiguous when multiple settings match.**
   - Location: `taskplane-tasks/TP-092-activity-histogram/STATUS.md:135`
   - The contract says activity sport/type matches `SportSettings.Type` or any `SportSettings.Types`, but it does not define normalization or precedence if multiple sport settings match (for example, one setting has `Type:"Run"` and another has `Types:["Run"]`, or duplicate matching settings appear in profile order).
   - Different implementations could pick different zone arrays and still be consistent with “matches”, producing different bucket edges and labels for the same activity.
   - Please define deterministic selection rules: trim/case-folding, whether `Activity.Type` or `SubType` is considered, exact `Type` vs `Types` precedence, and tie-breaking (for example, first profile order after precedence). Also state whether no deterministic match falls back to fixed-width.

## Non-blocking notes

- `STATUS.md:137` says profile/details missing falls back to fixed-width, but pace output units depend on athlete preferred units from the profile. Consider saying that profile-fetch failure defaults pace output to `seconds_per_km`, or that profile-fetch failure is a user-facing fetch error rather than a fallback.
- Fixed-width labels are deterministic in shape, but not in numeric formatting. Golden tests will be less brittle if the contract names the formatting/rounding used for label values, even though `lower`/`upper` carry the authoritative numeric bounds.
- The execution-log rows remain appended under `## Notes` instead of the `## Execution Log` table. This is STATUS hygiene only.

Tests were not run; this step only changes task/status documentation.
