# Plan Review — TP-085 Step 3

**Verdict:** APPROVE

The revised Step 3 plan is now specific enough for the task. It keeps the original fixture goal, adds helper-level coverage for provider inference, and, most importantly, requires exact response-level `workaround` plus stable `reason` assertions for every Strava-unavailable constructor path identified during Step 1:

- list rows / `activityRow` (`strava_tos`)
- messages / `stravaUnavailableMessagesResponse` (`strava_tos`)
- intervals / `stravaUnavailableIntervalsResponse` (`strava_blocked`)
- streams/splits via `detectActivityUnavailable` shared fallback (`strava_blocked`)
- extended metrics / `stravaUnavailableExtendedMetricsResponse` (`strava_blocked`)

The plan also preserves the required known-provider and unknown-provider coverage. During implementation, make sure at least one known-provider case and one unknown-provider case are asserted at the response level, not only in helper tests, and replace any existing stale substring expectations such as the old `connect device directly` wording with exact comparisons to the new Connections-page remedy.

The targeted Strava/unavailable test run called out in the plan is appropriate for Step 3; full `make test`, `make build`, and `make lint` can remain for the later verification steps already listed in STATUS.md.
