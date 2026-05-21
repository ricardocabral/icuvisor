# Review R001 — Plan review for Step 1

Verdict: **changes requested**

The Step 1 checklist is directionally right, but as written it is unlikely to reproduce the actual leak reliably.

## Blocking concerns

1. **A final `Close()` count assertion will not catch the bug.**
   The reported bug is `defer resp.Body.Close()` inside a retry loop. Deferred closes still run when the function returns, so a test that calls the method and only then asserts `closed == attempts` can pass even on the buggy implementation. To reproduce the leak, the test must observe that each retry response body is closed **before the next attempt starts**.

2. **`httptest.Server` alone cannot wrap the client-side `resp.Body`.**
   The handler writes bytes; it does not control the `io.ReadCloser` returned by `http.Client.Do`. If the plan keeps `httptest.Server`, add a wrapping `RoundTripper` around `server.Client().Transport` that replaces each returned `resp.Body` with a counting/tracking `io.ReadCloser`. Alternatively, use a custom `roundTripFunc` as the existing `client_test.go` retry/body-close tests do.

3. **Use a retried event-write path.**
   `doJSONBody` does not retry `POST` writes (`shouldRetryWriteStatus` returns false for POST). A 503-three-times-then-200 test must exercise a retriable method, e.g. `AddOrUpdateEvent` with `EventID` set so it uses `PUT`, or call `client.doJSONBody(ctx, http.MethodPut, ...)` directly from the package test.

4. **Set retry config explicitly.**
   Three 503s then a 200 requires `RetryConfig{MaxAttempts: 4, BaseDelay: time.Nanosecond, MaxDelay: time.Nanosecond}` (and no jitter) so the test is fast and deterministic. The default max attempts is 3, so the described sequence would never reach the 200.

5. **Current checkout does not appear to match the prompt's precondition.**
   `internal/intervals/events.go` already drains and closes bodies in the non-2xx retry branch; the remaining `defer resp.Body.Close()` is on the terminal success branch. A per-iteration leak regression test may already pass in this worktree. Before recording “fails on main”, verify against the actual target baseline or reconcile the task status/prompt with the current code.

## Recommended Step 1 test shape

- Add a test in `internal/intervals/events_test.go` for `doJSONBody`/event update over `PUT`.
- Instrument bodies client-side with a custom transport.
- On each `RoundTrip` after the first, assert the previous response body has already been closed before returning the next response. This is the key assertion that fails with a deferred close inside the loop.
- Return 503 for attempts 1-3 and 200 with valid JSON for attempt 4.
- Also assert total attempts and final closed count after the call, but treat those as secondary assertions.
- Run the targeted test and record the failing output in `STATUS.md` if it fails against the intended baseline.

With these adjustments, Step 1 should produce a meaningful regression test rather than one that can pass while the retry-loop leak still exists.
