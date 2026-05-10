# Plan Review — TP-002 Step 4

## Verdict

APPROVE

## Findings

No blocking findings.

## Notes

- The Step 4 plan is appropriately scoped to the intervals client read path only: add a `GetAthleteProfile(ctx)`-style method, use the already-normalized configured athlete ID, and avoid MCP tool wiring in this task.
- Implement the call against the Step 1 planned primary endpoint: `GET /api/v1/athlete/{id}` via the existing `doJSON(ctx, http.MethodGet, ..., "athlete", c.athleteID)` helper, so Basic Auth, `User-Agent`, retries, context cancellation, and structured errors remain centralized.
- Return the existing typed `AthleteWithSportSettings` shape (or a pointer to it) and keep unknown/raw payload preservation out of the default API unless a fixture proves a required v0.1 field is otherwise inaccessible.
- Ensure the returned profile exposes the v0.1 fields needed later by `get_athlete_profile`: normalized athlete ID, display/name fields, timezone/locale/unit preference fields, and sport settings/thresholds/zones including FTP.
- Do not introduce write endpoints, activities/wellness/events APIs, network tests, logging of raw profile payloads, or dependency changes in this step.
- Add focused unit coverage either in Step 4 or Step 5 for the public profile method: path construction with the configured normalized athlete ID, successful decoding of representative sport settings, and propagation/classification of `ErrUnauthorized`, `ErrNotFound`, and retryable upstream errors through the existing helper.
