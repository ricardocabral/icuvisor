# Plan Review — TP-057 Step 1

Verdict: **needs changes before implementation**

I reviewed `PROMPT.md`, `STATUS.md`, and the current retry code. There is no separate detailed plan beyond the Step 1 checklist in `STATUS.md`, so this review treats that checklist as the plan.

## Findings

1. **Step 1 must capture effective call-site decisions, not only helper return values.**
   `shouldRetry(resp, err)` has no `ctx` or attempt cap; `doJSONQuery` applies `ctx.Err() == nil && attempt < c.retry.MaxAttempts` around it. If the table only calls the five helpers directly, it will not lock the current GET/read retry behaviour for cancellation and max attempts. The table should include the effective read-path predicate as exercised today, or a small adapter that combines `shouldRetry` with the call-site guards.

2. **The current plan misses an existing retry call site in `workout_library.go`.**
   `doNoJSON` uses `c.shouldRetryTransport(ctx, attempt)` and `c.shouldRetryStatus(resp.StatusCode, attempt)` for DELETE requests. Step 1 should include this DELETE/no-body path in the truth table; otherwise the later consolidation can regress deletes while tests still pass.

3. **`ctx` state needs to be an explicit table dimension.**
   The prompt goal is `ctx`-first and cancellation-aware, but Step 1 only lists `(method, status, err, attempt)`. Add active vs canceled context cases, at least for transport errors and GET/read call-site decisions. Also document the current asymmetry: write/status helpers do not take `ctx` today, while the new consolidated method is expected to respect `ctx.Done()`.

4. **The Step 1 test must be designed to survive helper deletion.**
   The prompt says to delete the old helpers later, so a test that permanently calls `shouldRetry*` by name will fail after Step 3. Prefer a shared table of expected decisions plus an adapter for current helpers in Step 1, then switch the adapter to `decideRetry` during the refactor. That preserves the truth table while allowing helper removal.

5. **Include deterministic wait/`Retry-After` coverage now or explicitly defer it.**
   The target `decideRetry` returns `(retry bool, wait time.Duration)`, but the existing helpers only return booleans. To avoid adding untested behaviour in Step 2, add Step 1 coverage for the current delay rules via `retryDelay`/`parseRetryAfter` with `Jitter: 0`: normal backoff, `Retry-After` override, max-delay clamp, and attempt-boundary behaviour. If this is intentionally deferred, update `STATUS.md` with that rationale.

6. **Resolve the 408/425 mismatch before locking expectations.**
   The prompt asks for coverage of `408/425/429/5xx`, but current code retries only `429` and `>=500`; `408` and `425` are currently false. Step 1 should include `408` and `425` as current-behaviour false cases, and any later change to make them retryable needs the explicit `STATUS.md` justification/reviewer sign-off required by the task's “Do NOT” section.

## Suggested Step 1 shape

- Add `internal/intervals/retry_test.go` in package `intervals` so unexported helpers are accessible.
- Use a `Client` with deterministic retry config, e.g. `MaxAttempts: 3`, fixed base/max delay, `Jitter: 0`.
- Table rows should cover:
  - read GET effective decision: transport error, 400, 408, 425, 429, 500/503; attempts below and at `MaxAttempts`; active and canceled contexts;
  - write/body methods: POST (currently never retries), PUT/PATCH/DELETE/GET-as-non-POST (currently retryable on transport and retryable statuses), with attempt cap;
  - no-body DELETE path from `doNoJSON`;
  - retry delay/`Retry-After` cases separately if not part of the decision table.

With these adjustments, Step 1 will provide a useful safety net for the consolidation instead of only testing the soon-to-be-deleted helper internals.
