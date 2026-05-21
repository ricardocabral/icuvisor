# Code Review — Step 2: Broaden the fallback predicate

Decision: **revise**.

## Findings

1. **Shared predicate broadening changes `get_activity_messages` out of scope.**  
   `isActivityReadFallbackCandidate` is used not only by the TP-074 activity-detail tools, but also by `get_activity_messages` (`internal/tools/get_activity_messages.go:86`). Broadening it at `internal/tools/get_activity_details.go:239-240` means a messages endpoint `ErrRateLimited`/`ErrUpstream` now triggers a detail lookup and can return a successful `strava_tos` unavailable response when the activity detail is Strava-blocked, instead of preserving the original messages failure. `get_activity_messages` is not in TP-074's four affected tools and its terminal path still returns the old generic error when the fallback misses, so this is an accidental contract change and can also mask 429/5xx message failures as Strava ToS. Please either split the predicate so `get_activity_messages` keeps its previous `ErrNotFound`/`ErrUnauthorized` behavior, or explicitly include/update/test that tool as part of this product change.

## Verification

- `go test ./internal/tools -run 'Test(GetActivityIntervalsUnavailableReasons|GetActivityIntervalsUnavailableForHiddenSuccessPayload|GetActivityIntervalsFallbacksToDetailsForBlockedError)'` — **passes**.
- `go test ./internal/tools -run 'Test(GetActivityIntervalsUnavailableReasons|GetActivityStreamsUnavailableReasons|GetActivitySplitsUnavailableReasons|ExtendedMetricsUnavailableReasons|GetActivityIntervalsUnavailableForHiddenSuccessPayload|GetActivityIntervalsFallbacksToDetailsForBlockedError|ExtendedMetricsStravaUnavailableIncludesFullWhenRequested)'` — **fails** for streams/splits/extended metrics, which appears to be the pending Step 3 wiring rather than a Step 2 regression.
