# Plan Review — TP-085 Step 3

**Verdict:** REVISE

The Step 3 plan is close, but it is not quite specific enough to guarantee the fixture/assertion coverage promised by the task and by the Step 1 discoveries. In particular, the plan says to cover “one direct constructor outside the shared fallback,” while Step 1 identified multiple direct Strava-unavailable constructors and explicitly expanded the file scope to include messages and extended metrics. Covering only one of those paths could still leave stale wording or weak `reason`-only assertions in another user-facing response.

## Required plan changes

1. **Cover every identified Strava-unavailable construction path at least once with an exact workaround assertion.**
   The plan should explicitly name the paths/tests it will cover:
   - `activityRow` via `get_activities` / list-row output (`strava_tos`).
   - `detectActivityUnavailable` shared fallback via `get_activity_streams` and/or `get_activity_splits` (`strava_blocked`). If both output shapes are easy to cover, keep both as the current plan says.
   - `stravaUnavailableMessagesResponse` via `get_activity_messages` (`strava_tos`).
   - `stravaUnavailableIntervalsResponse` via the direct hidden-success payload path in `get_activity_intervals` (`strava_blocked`).
   - `stravaUnavailableExtendedMetricsResponse` via `get_extended_metrics` (`strava_blocked`).

2. **Assert both exact strings at response level, not only helper level.**
   Helper-level assertions for Wahoo provider inference and unknown-provider fallback are useful, but at least one response-level test should prove the provider-aware Wahoo text is emitted from raw activity payloads, and at least one response-level test should prove the provider-neutral fallback is emitted when provider evidence is absent or unallowlisted.

3. **Keep reason-code assertions alongside exact workaround assertions.**
   The task requires `unavailable.reason` to remain stable. When updating tests from “reason only” or “non-empty workaround” to exact string checks, retain the `strava_tos` / `strava_blocked` assertions in the same cases.

4. **Replace stale substring expectations.**
   Existing tests that look for the old generic wording such as `connect device directly` should be updated to compare against the new exact Connections-page remedy. Otherwise the suite may either fail for the wrong reason or continue to allow wording drift.

## Suggested targeted verification

After revising the plan and implementing the assertions, run a targeted package test for the affected tool tests, for example:

```sh
go test ./internal/tools -run 'Test(IsStravaBlocked|StravaBlockedWorkaround|GetActivities.*Strava|GetActivity(Streams|Splits|Intervals|Messages).*Unavailable|GetExtendedMetrics.*Strava)'
```

The exact regex can vary, but it should exercise all updated Strava/unavailable assertion sites.

Once the plan names the direct constructor coverage explicitly and includes both provider-aware and provider-neutral exact response assertions, Step 3 should be safe to implement.
