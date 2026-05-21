# Plan Review — TP-057 Step 3

Verdict: **request changes**

I reviewed `PROMPT.md`, `STATUS.md`, and the current Step 1/2 code in `internal/intervals/client.go`, `events.go`, `workout_library.go`, and `retry_test.go`. The Step 3 checklist is close, but it is still too narrow for safely deleting the old helpers.

## Blocking issues

1. **Include `workout_library.go` call sites in Step 3.**

   The plan says to replace call sites in `client.go` and `events.go`, but `internal/intervals/workout_library.go` still calls `c.shouldRetryTransport` and `c.shouldRetryStatus` in `doNoJSON`. If Step 3 deletes the helpers without routing this path through `decideRetry`, the package will not compile. If it leaves those helpers behind, it misses the task acceptance criterion of replacing the five `shouldRetry*` helpers with one retry-decision method.

   Required plan update: route `doNoJSON` through `decideRetry(ctx, http.MethodDelete, resp, err, attempt)` as part of Step 3, and keep the no-body DELETE truth-table rows passing.

2. **Specify how call sites use the returned `wait`.**

   `decideRetry` now returns `(retry bool, wait time.Duration)`, so Step 3 should not continue to parse `Retry-After` only to compute sleep separately at each call site. Otherwise the retry delay logic remains duplicated and jitter/`Retry-After` handling can drift.

   Required plan update: each call site should follow the same shape:

   - call `retry, wait := c.decideRetry(ctx, method, resp, err, attempt)`;
   - if `retry`, sleep for exactly `wait` and continue;
   - preserve the existing context-cancellation error shape from `sleepBeforeRetry`/waiting.

   This likely means either changing `sleepBeforeRetry` to accept a precomputed wait duration (or adding a small `sleep(ctx, wait)` helper) rather than passing a precomputed delay into a parameter named `retryAfter`.

3. **Update the truth-table tests before deleting helpers.**

   `retry_test.go` still has `TestCurrentRetryDecisionTruthTable` and `currentRetryDecision` calling `shouldRetry`, `shouldRetryTransport`, `shouldRetryStatus`, `shouldRetryWrite`, and `shouldRetryWriteStatus`. Deleting the helpers will break the tests unless Step 3 explicitly retargets this adapter.

   Required plan update: keep the shared case table, but remove old-helper references from tests after call-site routing. The remaining test should validate `decideRetry` against the locked truth table, including the read/write/no-body path distinctions. This also makes the Step 4 `grep -n "shouldRetry" internal/intervals/` verification meaningful across test files as well as production files.

## Required implementation details to carry in the plan

- Preserve status-error handling order: compute the retry decision while `resp` is available, drain and close the response body before sleeping/continuing, and keep existing close-error behavior on non-retry paths.
- Preserve the Step 1 canceled-context asymmetry: read status decisions stop when `ctx` is canceled, while write/no-body DELETE status decisions still decide retry and then the wait returns the context error.
- Preserve method policy exactly: `POST` does not retry; existing non-POST write/delete paths do retry for transport errors and `429`/`5xx` while attempts remain.
- Update `CHANGELOG.md` under `[Unreleased]` / `Changed` in the same step, as already listed.
- Run the focused retry tests first, then `make build`, `make test`, `make test-race`, and `make lint`.

Once the plan is expanded to include the `workout_library.go` path and the test/wait-handling details above, Step 3 should be safe to implement.
