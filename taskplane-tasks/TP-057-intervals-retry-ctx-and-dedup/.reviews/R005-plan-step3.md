# Plan Review — TP-057 Step 3

Verdict: **approve**

I reviewed `PROMPT.md`, `STATUS.md`, and the current Step 1/2 retry code in `internal/intervals/client.go`, `events.go`, `workout_library.go`, and `retry_test.go`.

The updated Step 3 plan addresses the blocking gaps from R004:

- It now includes the `workout_library.go` / `doNoJSON` DELETE path, so deleting the old helpers should not leave compile-time references behind.
- It explicitly routes call sites through the `(retry, wait)` returned by `decideRetry`, which is necessary to avoid leaving retry-delay / `Retry-After` decision logic duplicated at each caller.
- It includes retargeting the Step 1 truth-table adapter away from the old `shouldRetry*` helpers before deleting them.
- It calls out preserving no-body DELETE behavior, close/drain ordering, context-cancellation behavior, changelog update, and the full build/test/lint matrix.

Implementation cautions to keep in mind, but not blockers for the plan:

1. When retrying status responses, compute `retry, wait := c.decideRetry(...)` while `resp` is still available, then drain/close the body before sleeping and continuing.
2. Preserve the existing non-retry close-error behavior: if the response body close fails on a non-retry status path, return the close error before the API status error as the current code does.
3. Preserve the intentional canceled-context asymmetry locked by the tests: GET status retries stop immediately when `ctx` is canceled, while write/no-body DELETE status decisions may still return retry and let the subsequent wait surface the context error.
4. After retargeting tests, `grep -n "shouldRetry" internal/intervals/` should not find old production or test helper references.

With those details followed, Step 3 is safe to implement.
