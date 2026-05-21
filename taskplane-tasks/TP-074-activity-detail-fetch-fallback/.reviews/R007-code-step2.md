# Code Review — Step 2: Broaden the fallback predicate

Decision: **approve**.

## Findings

No blocking findings for Step 2.

The revision keeps `get_activity_messages` on the legacy `not_found`/`unauthorized` fallback predicate while broadening the activity-interval read fallback to cover `ErrRateLimited` and `ErrUpstream`. The interval handler now preserves context cancellation, positively emits `strava_blocked` only after a successful fallback activity lookup, and otherwise returns a structured `unavailable.reason` derived from the original read error.

## Verification

- `go test ./internal/tools -run 'Test(GetActivityIntervalsUnavailableReasons|GetActivityIntervalsUnavailableForHiddenSuccessPayload|GetActivityIntervalsFallbacksToDetailsForBlockedError)'` — **passes**.
- `go test ./internal/tools` — **fails** for the Step 1 red tests on `get_activity_streams`, `get_activity_splits`, and `get_extended_metrics`; this matches the pending Step 3 wiring rather than a Step 2 regression.
