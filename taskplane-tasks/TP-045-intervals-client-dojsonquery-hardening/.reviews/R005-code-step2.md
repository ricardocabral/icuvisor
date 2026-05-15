# Code Review — TP-045 Step 2

**Verdict:** APPROVE

## Findings

No blocking code findings for Step 2.

The implementation splits `doJSONQuery` as requested: request execution is isolated in `do`, retry classification is side-effect-free in `shouldRetry`, successful responses are read through `readBody` with `io.LimitReader(maxResponseBodyBytes+1)`, and JSON decoding now happens from a bounded buffer after the response body is closed. The retry loop closes every non-nil response body before returning or continuing, and there is no `defer resp.Body.Close()` inside `internal/intervals/client.go`'s retry loop.

`ErrResponseTooLarge` is a proper sentinel, and the caller wraps it in a way that preserves `errors.Is(err, ErrResponseTooLarge)`. Close errors do not mask bounded-read errors on the success path.

## Verification run

- `go test ./internal/intervals` — pass
- `go test ./...` — pass
- `git diff --check 8e56243e30084f09972e8d386c7695d6718e60db..HEAD` — fails only on trailing whitespace in `taskplane-tasks/TP-045-intervals-client-dojsonquery-hardening/.reviews/R004-plan-step2.md` lines 11, 14, 17, and 20. This is review-document hygiene rather than a Step 2 code issue, but it is worth stripping before merge if whitespace checks are enforced.

## Notes for later steps

- Step 3 still needs to replace `normalizeRetryConfig` with `RetryConfig.WithDefaults`; the zero-value struct comparison remains for now as expected by the current step split.
- Step 4 should add the planned tests for oversize responses, body-close accounting, retryable statuses, non-retryable 4xx, and budget exhaustion.
