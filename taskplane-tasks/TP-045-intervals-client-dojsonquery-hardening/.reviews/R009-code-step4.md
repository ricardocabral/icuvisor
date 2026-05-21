# Code Review — TP-045 Step 4 Tests

**Verdict:** Request changes

The new tests compile and `go test ./internal/intervals` passes, but a few Step 4 acceptance checks are not actually locked down yet.

## Findings

1. **Context-cancel body-close test does not exercise the retry-sleep cancellation path.**  
   In `TestDoJSONClosesResponseBodyAcrossPaths`, the `context canceled with response` case cancels the context before calling `doJSON` and then expects `ErrUpstream` (`internal/intervals/client_test.go:236-247`). Because the fake `RoundTripper` ignores the canceled request context, this only proves that a response returned under an already-canceled context is closed before returning a status error. It does not prove the required path from the Step 4 plan: first attempt returns a retryable response, that response body is closed, the context is canceled before/during `sleepBeforeRetry`, and the returned error wraps `context.Canceled`. A regression that sleeps/returns on cancellation before closing the retryable response would not be caught. Please make this case cancel from the response body's `Close` (or an equivalent deterministic hook), assert one close, assert no second attempt, and assert `errors.Is(err, context.Canceled)`.

2. **Retry success tests do not assert the successful response was decoded.**  
   `TestDoJSONRetriesRateLimitThenSucceeds` and `TestDoJSONRetriesServerErrorThenSucceeds` only assert no error and the attempt count (`internal/intervals/client_test.go:323-329`, `348-354`). The prompt asks these paths to assert the final value is decoded. As written, a regression that returns nil after a retry without decoding into `out` would still pass. Add an assertion such as `got.ID == "i12345"` in both tests.

3. **Status-error tests should assert the structured error, not just sentinels/strings.**  
   `TestDoJSONDoesNotRetryBadRequest` only checks `ErrUpstream`, and `TestDoJSONRetryBudgetExhaustionReturnsLastStatus` checks the 503 status via `strings.Contains(err.Error(), "HTTP 503")` (`internal/intervals/client_test.go:370-393`). The approved Step 4 plan called for `errors.As` into `*intervals.Error` and checking `StatusCode`, which better locks down the client contract and avoids brittle string matching. Please assert `errors.As(err, &apiErr)` with `apiErr.StatusCode == http.StatusBadRequest` / `http.StatusServiceUnavailable` while keeping the sentinel checks.

## Tests run

- `go test ./internal/intervals`
