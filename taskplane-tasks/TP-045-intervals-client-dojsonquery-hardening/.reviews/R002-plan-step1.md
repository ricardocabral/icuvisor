# Plan Review — TP-045 Step 1

**Verdict:** Approved for Step 2, with two implementation cautions.

The revised `STATUS.md` now records the concrete boundaries requested by R001: `doJSONQuery` remains the public/internal entry point and owns the retry loop; `do(ctx, query, pathParts...)` performs one GET attempt and returns a body owned by the caller; `readBody` is bounded at 32 MiB via `io.LimitReader`; and `shouldRetry` is classification-only while the outer loop owns attempt budget, context checks, sleeps, `Retry-After`, draining, and closes.

The plan also correctly preserves the current retry policy from `internal/intervals/client.go` and `go.mod`: hand-rolled retries, max attempts `3`, base delay `200ms`, max delay `2s`, retryable statuses `429`/`>=500`, transport retries only while the context is live and attempts remain, `Retry-After` capped by max delay, and the existing zero-config-vs-partial-config jitter semantics without the struct zero-value comparison.

## Implementation cautions

1. **Be explicit about success close-error semantics in tests or comments.**
   The plan says success-body close errors will be returned before bounded JSON decode. That is an intentional behavior change from the current deferred close, whose error is ignored on success. It is acceptable for hardening if intentional, but Step 4 should include or adjust a close-error test so this precedence is locked down rather than accidental.

2. **Ensure every early return after `readBody` closes the body.**
   Since the refactor avoids `defer` in the loop, the implementation needs a tight pattern: read bounded bytes, close the body, then return/decode. This matters especially for oversize and malformed bodies, not just the happy path.

No further Step 1 planning changes are required before implementation.
