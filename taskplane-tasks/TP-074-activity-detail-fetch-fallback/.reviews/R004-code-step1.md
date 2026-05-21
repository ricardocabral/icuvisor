# Code Review — Step 1: Add failing tests for each error class

Decision: **request changes**.

I ran:

- `git diff 598d3f0..HEAD --name-only`
- `git diff 598d3f0..HEAD`
- `go test ./internal/tools -run 'Test(GetActivityIntervalsUnavailableReasons|GetActivityStreamsUnavailableReasons|GetActivitySplitsUnavailableReasons|ExtendedMetricsUnavailableReasons|GetActivityIntervalsUnavailableForHiddenSuccessPayload|GetActivityIntervalsFallbacksToDetailsForBlockedError|ExtendedMetricsStravaUnavailableIncludesFullWhenRequested)'`

The targeted tests compile and fail red as intended, but there are two coverage/consistency problems that should be fixed before moving on.

## Findings

### 1. Conflicting extended-metrics expectations remain

`TestExtendedMetricsUnavailableReasons` now expects primary `GetActivity` sentinel failures (`ErrNotFound`, `ErrUnauthorized`, `ErrRateLimited`, `ErrUpstream`) to return `err == nil` with a structured `unavailable.reason` (`internal/tools/get_extended_metrics_test.go:35-65`). However, `TestExtendedMetricsSourceErrorsReturnUserError` still includes an `activityErr: intervals.ErrUnauthorized` case and asserts the handler returns the old generic `fetchExtendedMetricsMessage` error (`internal/tools/get_extended_metrics_test.go:162-188`).

After the implementation is changed to satisfy the new table, the existing `activity error` row will fail. Please update that old test so it only covers truly non-categorized source errors (for example a plain `errors.New(...)` activity error if that should still be generic), or split out the sentinel activity-fetch cases into the new structured-unavailable assertions.

### 2. No test proves `ErrUpstream` still routes through Strava-block detection

The task's root-cause fix is specifically to broaden the fallback predicate so upstream 400/5xx (`ErrUpstream`) routes through the activity detail lookup before returning a terminal category. The shared table currently covers:

- `strava_blocked` only with `upstreamErr: intervals.ErrUnauthorized` (`internal/tools/get_activity_details_test.go:158`)
- `upstream_unavailable_500/400` only where the fallback detail lookup also fails (`internal/tools/get_activity_details_test.go:162-163`)

That leaves a gap: an implementation could classify every `ErrUpstream` directly as `upstream_unavailable` without attempting the Strava-block fallback and still pass these tests. Please add rows for 500/400 `ErrUpstream` with a successful Strava-marked fallback activity and expected `reason: "strava_blocked"` for the tools that use the fallback detail lookup. This locks the core TP-074 acceptance criterion that 400/5xx route through Strava detection, not just terminal categorization.

## Notes

The existing Strava expectation updates from `strava_tos` to `strava_blocked` are scoped to the activity detail read tools, which matches the approved plan. The helper assertions for no fabricated collections/metrics and workaround-only-for-Strava are useful.
