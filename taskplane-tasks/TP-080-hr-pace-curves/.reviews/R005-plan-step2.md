# Plan Review — TP-080 Step 2

Verdict: **APPROVE**

The revised Step 2 plan in `STATUS.md` addresses the blocking gaps from R004 and is ready for implementation.

## What looks good

- The two new tools now have distinct metric contracts:
  - `get_hr_curves` is duration-axis based and uses `duration_seconds` / `heart_rate_bpm` from `ListAthleteHRCurves`.
  - `get_pace_curves` is distance-axis based and uses `distance_meters` plus upstream elapsed seconds from `ListAthletePaceCurves`.
- The pace response contract is now explicit enough to implement safely: preserve raw `elapsed_seconds`, emit exactly one preferred pace field (`pace_seconds_per_km` or `pace_seconds_per_mile`), include `activity_id`, and expose unit metadata in `_meta.units`.
- The default pace bucket contract is specified as the run-style best-effort set `400,1000,1609,5000,10000` meters, with positive/sorted/dedup normalization and `missing_buckets` metadata.
- The plan now covers the shared `internal/toolcatalog` additions, which are required for registry validation, coach ACLs, and athlete-scoped schema injection.
- Registration scope is correct: both tools belong in the full toolset under the fitness catalog group, with catalog/tier/coach-ACL test updates.
- Profile behavior for pace units is defined: profile fetch failure returns a short user-facing error, and unknown unit preferences fall back to metric via existing profile-unit behavior.

## Non-blocking implementation notes

- Keep the HR/pace sport argument contract explicit in the schemas and tests. The intervals client permits omitted `type` for HR/pace, but when a caller supplies `sport`, it should be forwarded and reflected consistently in response metadata.
- Preserve context cancellation behavior when adding the profile lookup for pace, matching the existing tool conventions where possible.
- Include tests that assert the unit-disambiguated pace field choice, not just the numeric conversion, so metric and imperial outputs cannot both appear or silently swap.

