# Plan Review — Step 3: Mirror the shape across the four tools

Decision: **revise plan**.

The direction is right: centralize the classification/Strava-detection path and make `get_activity_intervals`, `get_activity_streams`, `get_activity_splits`, and `get_extended_metrics` all return successful tool results with a structured `unavailable` object. However, the current Step 3 plan is too underspecified for the shape changes needed across the four different response types.

## Required plan revisions

1. **Define the helper API around actual client capabilities.**
   - Current `get_activity_streams` only receives an `ActivityStreamsClient`; it has no explicit `ActivityDetailsClient` even though the tests' fake happens to implement `GetActivity`.
   - The plan should say whether `newGetActivityStreamsTool`/handler will take an additional `ActivityDetailsClient` and update `registry.go`, or whether a narrower combined interface will be introduced.
   - Drop `athleteID` from the proposed helper unless it is genuinely needed; activity detail clients here take only `activityID`.

2. **Avoid refetching the primary activity for `get_extended_metrics`.**
   - For intervals/streams/splits, a fallback `GetActivity` lookup is appropriate after the endpoint-specific call fails.
   - For `get_extended_metrics`, the primary call is already `GetActivity`; on primary failure the tool should classify the original error directly, not call `GetActivity` again as a fallback/retry.
   - The plan should include a small helper split such as: classify terminal error from original error, plus a separate optional `detectStravaBlocked(ctx, detailsClient, activityID, originalErr)` used only when the failed endpoint is not already `GetActivity`.

3. **Preserve success-path response shapes while omitting fabricated data on unavailable responses.**
   - `get_activity_streams` currently has non-omitempty `streams`; `get_activity_splits` has non-omitempty `split_unit`, `source`, and `splits`.
   - Tests require unavailable responses to omit collections like `streams` and `splits`, but the task also says not to change success-path response shape.
   - The plan should specify how this is achieved without accidentally omitting empty collections on successful responses. A separate unavailable payload struct per tool, or careful pointer/omitempty handling only on unavailable payloads, is safer than adding `omitempty` to existing success fields indiscriminately.

4. **Keep classification based on the original failing endpoint error.**
   - The fallback detail lookup may prove `strava_blocked`; otherwise the returned reason must come from the original sentinel (`not_found`, `unauthorized`, `rate_limited`, `upstream_unavailable`).
   - A fallback `GetActivity` error should only override the response when it is `context.Canceled` or `context.DeadlineExceeded`.

5. **Update all constructors/call sites and tests intentionally.**
   - If `newGetActivityStreamsTool` gains a details client, update `registry.go` and existing tests.
   - For `get_activity_splits`, be explicit about which failing call is classified: the ignored manual-intervals miss should remain optional, while the streams fallback failure should produce the structured unavailable response.

Once the plan includes these details, the proposed helper extraction is appropriate and should keep the four tools aligned without turning this into an out-of-scope shared-base refactor.
