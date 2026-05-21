# Plan Review — Step 2: Broaden the fallback predicate

Decision: **approve plan**.

The Step 2 plan is aligned with TP-074's core behavior change: expand eligibility beyond the current `ErrNotFound`/`ErrUnauthorized` path, keep context cancellation as a real error, prove `strava_blocked` only through a successful activity-detail lookup, and otherwise return a structured `unavailable.reason` derived from the original sentinel error.

## Notes for implementation

- Classify from the **original** failing read error, not from the fallback `GetActivity` error. The fallback error should only override behavior when it is `context.Canceled` or `context.DeadlineExceeded`.
- Preserve the distinction between `unauthorized` and `strava_blocked`: an `ErrUnauthorized`/403 should become `strava_blocked` only when the fallback detail activity is positively detected by `isStravaBlocked(activity)`.
- Include `ErrRateLimited` in the structured terminal categories. If you broaden the predicate to perform the detail lookup on a 429, be careful not to hide the original `rate_limited` result when that lookup fails; alternatively, it is acceptable to classify 429 directly without an extra lookup if that fits the helper design.
- Change the Strava reason to `strava_blocked` only for the in-scope activity detail read tools. Leave unrelated existing contracts such as `get_activities` / `get_activity_messages` on `strava_tos` unless a separate product change asks for the rename.
- For this step, avoid duplicating categorization logic in multiple handlers if Step 3 is going to extract the shared helper. A small private classifier introduced now and reused in Step 3 is fine.
- Ensure unavailable responses remain successful tool results (`err == nil`) and do not include fabricated detail collections/metrics for non-Strava terminal categories.

With those constraints, the plan should unblock the red Step 1 tests without changing success-path response shapes.
