# Plan Review — TP-045 Step 1

**Verdict:** Needs revision before implementation.

The step checklist in `STATUS.md` is still only the original task outline. Step 1 asks to sketch and settle the boundaries, but no concrete mapping, body-size decision, or retry-policy confirmation has been recorded yet. Because this refactor touches every JSON GET call, the implementation should not start until those decisions are written down in `STATUS.md`.

## Required additions to the Step 1 plan

1. **Record the concrete helper boundaries, including the query/path inputs.**
   The prompt's shorthand `do(ctx) (*http.Response, error)` is not enough for this codebase because `doJSONQuery` currently builds the request from `pathParts` and optional `url.Values`. The plan should specify something equivalent to:
   - `do(ctx, query, pathParts...) (*http.Response, error)` or an outer-loop closure that captures `query`/`pathParts`.
   - `readBody(io.Reader) ([]byte, error)` using a bounded reader.
   - `shouldRetry(resp *http.Response, err error) bool` for classification only.
   - outer loop owns attempt budget, `Retry-After`, sleep, and all body closes.

2. **Make body lifecycle rules explicit.**
   The current code drains and closes non-2xx bodies before retry/status-error return, and defers close on success inside the loop. The plan should state the new invariant: every non-nil `resp.Body` is closed in the same loop iteration before `continue` or `return`; no `defer resp.Body.Close()` inside the retry loop. It should also say whether close errors are returned or ignored on success so the behavior is intentional rather than incidental.

3. **Settle and document the body-size cap.**
   Accept the prompt recommendation unless there is a reason not to: package-level `const maxResponseBodyBytes = 32 << 20` with a short note is fine. The plan should also name the sentinel error location, likely `errors.go` as exported `ErrResponseTooLarge`, so callers/tests can use `errors.Is` and the client doc note can point to it.

4. **Confirm the existing retry policy exactly.**
   From `internal/intervals/client.go` and `go.mod`, retry is currently hand-rolled; there is no retry/backoff dependency to preserve. Record these parameters in `STATUS.md`:
   - default max attempts: `3`
   - default base delay: `200ms`
   - default max delay: `2s`
   - default jitter: `0.2`, but only for an entirely zero `RetryConfig`; partial configs currently get zero jitter unless they explicitly set it
   - retryable statuses: `429` and `>=500`
   - transport errors retry only while `ctx.Err() == nil` and attempts remain
   - `Retry-After` seconds/HTTP-date is honored and capped at `MaxDelay`
   - exponential delay remains `BaseDelay << (attempt-1)` plus existing jitter behavior

5. **Clarify `shouldRetry` vs attempt budget.**
   Since the requested `shouldRetry(*http.Response, error) bool` is classification-only, it should not inspect attempt count or sleep. The outer loop should combine `attempt < c.retry.MaxAttempts`, context state, and `shouldRetry(...)`. This keeps the split aligned with the task and avoids hiding budget logic in the classifier.

6. **Preserve `RetryConfig` semantics when replacing `normalizeRetryConfig`.**
   The plan should explicitly preserve the subtle current behavior without the struct zero-value comparison: zero config gets default jitter; partially specified config preserves zero jitter. A method like `func (cfg RetryConfig) WithDefaults() RetryConfig` can compute an explicit `allFieldsUnset` boolean before filling defaults, without using `cfg == (RetryConfig{})`.

## Suggested Step 1 status note

Add a short decision block to `STATUS.md` before coding, for example:

```md
## Decisions

- Body cap: 32 MiB package-level constant (`maxResponseBodyBytes = 32 << 20`); oversize reads return `ErrResponseTooLarge` wrapped by `doJSONQuery`.
- Retry remains hand-rolled; no new dependency. Defaults and retryable statuses stay unchanged.
- Helper split: `doJSONQuery` outer loop builds attempts via `doJSONAttempt(ctx, query, pathParts...)`, calls `shouldRetry` for classification, drains/closes status bodies before retry/error, reads success bodies with `readBody`, closes before decode, then decodes from the bounded buffer.
- `RetryConfig.WithDefaults` preserves zero-config default jitter and partial-config explicit zero jitter.
```

Once those points are recorded, the plan is sound for Step 2 implementation.
