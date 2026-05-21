# Plan Review — Step 1: Add failing tests for each error class

Decision: **request plan adjustments before coding**.

The step is directionally right (lock the current generic-error behavior down with failing tests first), but the test plan needs more precision to avoid brittle/non-compiling tests and to fully cover the promised response contract.

## Findings

1. **Do not let Step 1 tests require Step 3 signature changes.**
   `get_activity_streams` currently only accepts an `ActivityStreamsClient`; it has no detail/fallback client parameter. If the new tests are written against a future constructor signature, Step 1 will fail to compile rather than fail behaviorally. Either:
   - use the existing constructors and fakes that implement extra methods (the handler just will not call them yet), then update implementation later without changing the test call sites; or
   - make the minimal constructor/interface change in the same commit as implementation, not in the red-test-only step.

2. **Update existing Strava expectation tests as part of Step 1.**
   Current tests assert `unavailable.reason == "strava_tos"` for `get_activity_intervals` and `get_extended_metrics`. The task acceptance criteria require `"strava_blocked"` for these detail-read tools. If Step 1 only adds new tests, the later implementation will conflict with existing assertions. Keep unrelated tools (`get_activities`, `get_activity_messages`) out of scope unless the product decision is to rename the reason globally.

3. **Assert the structured shape, not only the reason string.**
   For non-Strava categories, the expected shape is `unavailable: { reason: "..." }`; `workaround` should be absent, not present as an empty string. The shared `unavailableReason` type currently has `Workaround string \`json:"workaround"\``, so tests should catch this by asserting no `workaround` key for `not_found`, `unauthorized`, `rate_limited`, and `upstream_unavailable`. Also assert the handler returns a successful `Result` (no `NewUserError`) and does not fabricate `streams`, `splits`, `metrics`, intervals, or samples for unavailable responses.

4. **The per-tool scenarios need to account for each tool's call graph.**
   A single “httptest server returning responses in sequence” will be brittle because `get_activity_splits` and `get_extended_metrics` call profile first, 5xx may retry unless disabled, and splits may call intervals before streams. Prefer route-based handlers or fake clients. In particular:
   - `get_activity_intervals`: terminal error is from `GetActivityIntervals`; Strava fallback comes from `GetActivity`.
   - `get_activity_streams`: terminal error is from `GetActivityStreams`; implementation will need a detail lookup path for Strava detection.
   - `get_activity_splits`: intervals errors are currently ignored; terminal error usually comes from the streams fallback path after no manual splits are available. Tests must set both intervals and streams behavior intentionally.
   - `get_extended_metrics`: `GetActivity` is the primary gate. Optional intervals/power-vs-HR errors have existing partial-source behavior for 404/403 but terminal behavior for upstream errors; tests should target the exact source being classified.

5. **Add or explicitly defer `rate_limited` coverage.**
   The acceptance criteria include `rate_limited`, and Step 2 mentions categorizing `ErrRateLimited`, but Step 1 only lists 403/404/500/400. Add a 429/`intervals.ErrRateLimited` row (at least at the shared-helper level, preferably per tool if the contract is per tool) or document why it is deferred. Otherwise the predicate can still miss rate-limit errors while tests pass.

6. **If using `httptest.Server`, disable retries and avoid sequence-only assertions.**
   Build the `intervals.Client` with `RetryConfig{MaxAttempts: 1}`; otherwise 500 cases may be retried and consume multiple queued responses. Route by `r.URL.Path` instead of global sequence, because profile/detail/fallback calls differ across tools.

## Suggested Step 1 shape

Use table-driven tests per tool (or a shared assertion helper) with compile-safe fake clients. Each row should assert:

- `err == nil` for categorized unavailable responses.
- `payload["unavailable"].reason` equals the expected category.
- `workaround` exists only for `strava_blocked`.
- `strava_imported == true` only for `strava_blocked`.
- No fabricated data collections/metrics are present on unavailable responses.
- Existing success-path tests still pass unchanged.

With those adjustments, Step 1 will create useful red tests that guide the implementation without over-coupling to transport details.
