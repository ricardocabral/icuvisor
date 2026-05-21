# Plan Review — TP-057 Step 1

Verdict: **approved with one clarification to carry into implementation**

I reviewed `PROMPT.md`, the updated `STATUS.md`, and the current retry call sites in `internal/intervals/client.go`, `events.go`, and `workout_library.go`. The Step 1 plan now addresses the material gaps from R001: it includes effective read-path and no-body DELETE decisions, active/canceled context cases, the current 408/425 non-retry behaviour, deterministic delay/`Retry-After` coverage, and a swappable adapter so the truth table can survive helper deletion.

## Required implementation note

- When building the shared table, make the write-path method matrix explicit. Current behaviour is **POST never retries**, while **non-POST write methods** (`PUT`/`PATCH`/`DELETE`, and any other non-POST method passed to the helper) share `shouldRetryTransport`/`shouldRetryStatus` semantics. The Step 1 checklist says “every `(method, status, err, attempt)` combination”, which is broad enough, but the test should include at least `POST`, `PATCH`, and `DELETE` rows so the later consolidation does not accidentally replace the current “POST-only non-retryable” rule with a generic idempotency rule.

## Non-blocking caution

- Be deliberate about canceled-context status cases. Today, read-path status retries are guarded by `ctx.Err() == nil`, but write/no-body status helpers do not take `ctx`; they may decide “retry” and then `sleepBeforeRetry` returns the context error. The Step 1 truth table should either capture this as an effective call-site outcome or document it next to the adapter, so Step 2’s `decideRetry(ctx, ...)` change does not hide an accidental error-shape change.

With those points handled during implementation, the plan is sufficient for Step 1.
