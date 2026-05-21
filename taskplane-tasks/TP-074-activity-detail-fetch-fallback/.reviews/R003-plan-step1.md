# Plan Review — Step 1: Add failing tests for each error class

Decision: **approve plan**.

The Step 1 checklist now addresses the blocking concerns from the previous reviews. It calls for compile-safe table-driven tests, includes the full accepted reason set (`strava_blocked`, `not_found`, `unauthorized`, `rate_limited`, `upstream_unavailable` for both 500 and 400), updates only the in-scope activity-detail read Strava expectations, and names each tool's actual terminal error source.

## Notes for implementation

- Keep the tests red for behavior, not compilation. For `get_activity_streams` and `get_activity_splits`, use the existing constructor call sites in Step 1 and fakes that can later support detail lookup, rather than writing tests against a future constructor signature.
- Preserve the distinction between `strava_blocked` and `unauthorized`: a 403/`ErrUnauthorized` row should only become `strava_blocked` when the detail lookup proves the activity is Strava-blocked; otherwise it should assert `reason: "unauthorized"` with no `workaround` key.
- For unavailable responses, assert the handler returns `err == nil`, includes no fabricated collections/metrics, and omits `workaround` for all non-Strava reasons.
- Leave unrelated existing contracts (`get_activities`, `get_activity_messages`) on `strava_tos` unless a separate product decision broadens the rename.

With those constraints, this plan should produce useful failing tests aligned with TP-074's acceptance criteria.
