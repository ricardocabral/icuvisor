# Plan Review — TP-057 Step 2

Verdict: **request changes**

I reviewed `PROMPT.md`, `STATUS.md`, and the current retry helpers/call sites in `internal/intervals/client.go`, `events.go`, `workout_library.go`, plus the Step 1 truth table in `internal/intervals/retry_test.go`. The Step 2 checklist is directionally right, but it is too sparse for the two behavior traps this refactor is specifically meant to avoid.

## Blocking issues

1. **Clarify canceled-context status behavior before implementing `decideRetry`.**

   The plan says both “preserve current retry policy” and “respect `ctx.Done()`.” Today those are not identical across paths:

   - read JSON status retries are guarded by `ctx.Err() == nil`;
   - write JSON status retries and no-body DELETE status retries do **not** check `ctx` in the decision helper; they decide retry, then `sleepBeforeRetry` returns the context error.

   Step 1 intentionally captured this with canceled status rows such as `write delete status 503 still decides retry when context is canceled` and `no-body delete status 503 still decides retry when context is canceled`. A consolidated method with only `(ctx, method, resp, err, attempt)` can easily change the external error shape by returning `retry=false` on a canceled write/delete status and causing the call site to return the upstream API error instead of the context wait error.

   The Step 2 plan should explicitly state which behavior is intended. Given the task acceptance criteria, the safe plan is: **preserve the Step 1 truth table byte-for-byte unless `STATUS.md` records a deliberate policy change and reviewer sign-off.**

2. **Do not replace the existing POST-only write rule with generic HTTP idempotency.**

   The prompt mentions idempotency, but the existing behavior is not “only GET retries.” Current helpers retry transport/status failures for non-POST write methods, including `PUT`, `PATCH`, and `DELETE`; only `POST` is blocked. R002 already called this out, and Step 1 now locks it in.

   The Step 2 plan should spell out the method matrix for `decideRetry`:

   - `GET` read path: retry transport errors and `429`/`5xx` while attempts remain and context policy matches the truth table;
   - `POST`: never retry;
   - existing non-POST writes/no-body deletes: preserve current retry behavior for `PUT`/`PATCH`/`DELETE` unless a policy change is explicitly approved.

3. **Add a Step 2 test path that calls the new method before routing call sites.**

   Adding an unused method will compile even if it is wrong. Step 2 should use the shared Step 1 table to validate `decideRetry` directly while the old helpers still exist. That can be a second adapter/test or a switched adapter in a focused commit, but Step 2 should not rely on Step 3 call-site routing to discover semantic drift.

## Required implementation details to include in the plan

- `decideRetry` should be a pure decision helper: no response body reads/closes and no sleeping.
- It should return `wait=0` whenever `retry=false`; when retrying, compute wait through the existing `retryDelay(attempt, retryAfter)` path.
- For status responses, parse `Retry-After` from `resp.Header.Get("Retry-After")`; for transport errors, pass zero retry-after.
- Preserve the existing attempt cap (`attempt < c.retry.MaxAttempts`) and status set (`429` and `5xx`; not `408` or `425`).
- Define precedence for unusual inputs, especially `err != nil` with a non-nil `resp` and `resp == nil` with `err == nil`, so the table documents the behavior.

Once the plan addresses these points, the Step 2 implementation should be straightforward and reviewable.
