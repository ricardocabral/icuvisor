# Plan Review — Step 1: Add failing tests for each error class

Decision: **request small plan adjustments before coding**.

The revised plan addresses the main R001 concerns: compile-safe tests, route-based/fake clients, updating the detail-read Strava reason to `strava_blocked`, asserting handler success, and checking that unavailable responses do not fabricate data. I would still tighten three points so the red tests fully match the acceptance criteria and do not encode ambiguous behavior.

## Findings

1. **Add explicit `unauthorized` coverage distinct from the Strava-blocked 403 case.**
   The current plan covers “403 Strava-block fallback”, but that only verifies `ErrUnauthorized` can become `strava_blocked` when `GetActivity` proves the activity is Strava-blocked. The acceptance criteria also require `unavailable.reason == "unauthorized"`. Add a row where the original error is `intervals.ErrUnauthorized`/401/403 and the fallback detail lookup either fails with unauthorized or succeeds with a non-Strava activity, then assert `reason: "unauthorized"` and no `workaround`.

2. **Do not leave 400 as `upstream_unavailable/bad_request` in the test plan.**
   The task acceptance criteria list valid reasons as `strava_blocked`, `not_found`, `unauthorized`, `rate_limited`, and `upstream_unavailable`; `bad_request` is not accepted, and the prompt says not to introduce a new sentinel unless needed. Unless the task scope/acceptance is amended, Step 1 should expect 400 to classify as `upstream_unavailable`.

3. **Specify the terminal error source per tool, especially for `get_extended_metrics` and `get_activity_splits`.**
   “All four tools” is correct, but the tests need to target each handler’s actual call graph:
   - `get_activity_streams`: terminal error is `GetActivityStreams`; Strava detection will require a separate detail lookup path.
   - `get_activity_splits`: `GetActivityIntervals` errors are currently ignored, so unavailable classification should be tested from the streams fallback path after no usable manual splits are returned.
   - `get_extended_metrics`: Strava-blocked is currently detected from a successful primary `GetActivity`, while terminal categories for 404/401/429/500 should be tested against the primary activity fetch (or explicitly against optional source errors if that is the intended implementation path). Avoid an impossible test setup that requires the same primary `GetActivity` call to both return 403 and provide the Strava payload unless the planned implementation intentionally retries/falls back to a separate detail client.

## Suggested adjustment

Update the Step 1 checklist to enumerate the expected rows as: `strava_blocked`, `not_found`, `unauthorized`, `rate_limited`, `upstream_unavailable_500`, and `upstream_unavailable_400`, with a short note naming the failing upstream method for each tool. With that clarification, the plan should produce useful red tests aligned with the final contract.
