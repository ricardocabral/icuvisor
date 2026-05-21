# Plan Review — TP-045 Step 4

**Verdict:** Approved, with required test-design refinements.

The Step 4 checklist covers the required acceptance areas: bounded-body overflow, retryable 429/5xx, non-retryable 4xx, retry-budget exhaustion, response-body closure on every branch, and `RetryConfig.WithDefaults` semantics. That is enough to proceed, but the tests need a few concrete guardrails so they actually prove the hardening rather than only exercising the happy retry paths.

## Required refinements before/while implementing

1. **Do not rely only on `httptest.Server` handler counters to prove `resp.Body.Close`.**
   A server-side handler counter can prove attempts and writes, but it does not deterministically prove the client closed every response body. Keep or expand the existing fake `RoundTripper` / `closeTrackingBody` pattern for body-lifecycle tests. A table should cover success, retry-then-success, retry-budget exhaustion, oversize success body, 4xx no-retry, and a context-cancelled retry path, asserting one close per non-nil response returned.

2. **Make the context-cancel body-close case unambiguous.**
   Cancelling the context before `Do` returns produces no response body, so it cannot satisfy the close-accounting acceptance criterion. Prefer a test where the first attempt returns a retryable response, that response body is closed, and then the context is cancelled before/during `sleepBeforeRetry` (for example, cancel from the tracking body's `Close`). Assert the returned error wraps `context.Canceled` and that the retryable response body was closed exactly once.

3. **The oversize test must send/read more than the cap, not just advertise it.**
   Current `readBody` ignores `Content-Length` and trips only after reading `maxResponseBodyBytes+1` bytes through `io.LimitReader`. A test that only sets `Content-Length: huge` is insufficient. Use a streaming reader or chunked `httptest` response that actually emits `maxResponseBodyBytes+1` bytes, and assert `errors.Is(err, ErrResponseTooLarge)` plus body closure. Avoid a large `strings.Repeat` allocation if a small custom repeating reader is easy.

4. **Keep retry tests fast and deterministic.**
   For 429/503 tests, set `RetryConfig{MaxAttempts: ..., BaseDelay: time.Nanosecond, MaxDelay: time.Nanosecond}` so partial config semantics keep `Jitter == 0`. Avoid `Retry-After: 1` unless the test specifically targets `Retry-After`, otherwise the suite will sleep unnecessarily.

5. **Assert structured errors, not only strings.**
   For 4xx and exhausted 5xx cases, assert attempts and use `errors.As` to inspect `*intervals.Error.StatusCode`; for exhausted 503 also assert `errors.Is(err, ErrUpstream)`. For oversize, assert `errors.Is(err, ErrResponseTooLarge)` so the wrapping contract is locked down.

6. **Make `WithDefaults` table-driven enough to catch the subtle semantics.**
   Include at least: zero value gets all documented defaults including default jitter; partial config such as `RetryConfig{MaxAttempts: 5}` fills missing durations but preserves explicit zero jitter; negative jitter clamps to zero. This protects the specific audit concern from regressing.

With those refinements, the Step 4 plan is aligned with the prompt and should provide meaningful regression coverage for the refactor.
