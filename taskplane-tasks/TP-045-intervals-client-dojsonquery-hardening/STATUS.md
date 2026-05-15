# TP-045-intervals-client-dojsonquery-hardening — Status

**Current Step:** Step 5: Verify
**Status:** ✅ Complete
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 9
**Iteration:** 3
**Size:** M

---

### Step 1: Sketch the new function boundaries

**Status:** ✅ Complete

- [x] Map current `doJSONQuery` into `do` / `readBody` / `shouldRetry` / outer loop
- [x] Agree on body-size cap default (recommendation: 32 MiB)
- [x] Confirm existing retry policy parameters and re-plumb unchanged
- [x] Address R001 plan review by recording helper signatures, body lifecycle rules, retry budget split, oversize sentinel, and `RetryConfig` semantics

### Step 2: Implement the split

**Status:** ✅ Complete

- [x] `do(ctx) (*http.Response, error)` — single attempt, caller owns body close
- [x] `readBody(io.Reader) ([]byte, error)` — `io.LimitReader` + sentinel for oversize
- [x] `shouldRetry(*http.Response, error) bool` — classification only
- [x] Outer retry wrapper — no `defer` inside the loop
- [x] JSON decode from bounded buffer, not raw `resp.Body`

### Step 3: Replace `normalizeRetryConfig`

**Status:** ✅ Complete

- [x] Add `RetryConfig.WithDefaults` (or `NewRetryConfig`)
- [x] Update all call sites
- [x] Delete `normalizeRetryConfig` and the zero-value comparison

### Step 4: Tests

**Status:** ✅ Complete

- [x] Body-close accounting across all paths (success, retry-then-success, exhaustion, oversize, 4xx, ctx-cancel)
- [x] Oversize body trips `io.LimitReader` and returns sentinel
- [x] 429 retry path
- [x] 5xx retry path
- [x] 4xx no-retry path
- [x] Retry budget exhaustion
- [x] `RetryConfig.WithDefaults` behaviour (zero / partial)

### Step 5: Verify

**Status:** ✅ Complete

- [x] `make build` / `test` / `test-race` / `lint`
- [x] Grep confirms no `defer` inside a `for` block in `client.go`
- [x] `go doc ./internal/intervals` shows no public signature drift

---

## Decisions

- Helper split: keep public `doJSON` / `doJSONQuery` signatures unchanged. `doJSONQuery` owns the retry loop and passes each attempt to `do(ctx, query, pathParts...) (*http.Response, error)`, where `do` builds a GET request through `newRequest`, applies optional `url.Values`, sends it with the shared `*http.Client`, and returns a response whose body the caller owns. `readBody(io.Reader) ([]byte, error)` performs only bounded reads. `shouldRetry(*http.Response, error) bool` is classification-only: transport errors, 429, and >=500 are retryable; attempt budget, context state, `Retry-After`, sleeps, and body draining/closing remain in the outer loop.
- Body cap: use a package-level constant `maxResponseBodyBytes = 32 << 20` (32 MiB), matching the task recommendation and keeping config surface unchanged. `readBody` reads with `io.LimitReader(r, maxResponseBodyBytes+1)` and returns an exported sentinel `ErrResponseTooLarge` (defined with the other intervals errors) when the cap is exceeded, wrapped by the caller for context.
- Retry policy: preserve the hand-rolled policy; `go.mod` has no retry dependency to retain. Defaults remain max attempts `3`, base delay `200ms`, max delay `2s`, and default jitter `0.2` only for an entirely zero `RetryConfig`; partial configs keep explicit zero jitter unless they set `Jitter`. Retryable statuses remain HTTP `429` and `>=500`; transport errors retry only while `ctx.Err() == nil` and attempts remain. `Retry-After` seconds / HTTP-date handling remains capped by `MaxDelay`, and exponential backoff remains `BaseDelay << (attempt-1)` plus the existing jitter calculation.
- Body lifecycle: every non-nil `resp.Body` is drained as needed and closed in the same retry-loop iteration before any `continue` or `return`; no `defer resp.Body.Close()` inside the loop. Status-body close errors remain returned when no retry happens. Success-body close errors will be returned before bounded JSON decode, so decoding happens only after the bounded buffer is read and the response body is closed.
- Defaults constructor: replace `normalizeRetryConfig` with `RetryConfig.WithDefaults`. It can compute an explicit `allFieldsUnset` boolean from individual fields before filling defaults, but must not compare the whole struct to `RetryConfig{}`; this preserves zero-config default jitter and partial-config zero-jitter semantics.

## Notes

_Add notes as work progresses._

| 2026-05-15 14:58 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 14:58 | Step 1 started | Sketch the new function boundaries |
| 2026-05-15 15:01 | Review R001 | plan Step 1: UNKNOWN |
| 2026-05-15 15:04 | Review R002 | plan Step 1: APPROVE |
| 2026-05-15 15:06 | Review R003 | code Step 1: APPROVE |
| 2026-05-15 15:08 | Review R004 | plan Step 2: APPROVE |
| 2026-05-15 15:13 | Review R005 | code Step 2: APPROVE |
| 2026-05-15 15:17 | Review R006 | plan Step 3: APPROVE |
| 2026-05-15 15:21 | Review R007 | code Step 3: APPROVE |
| 2026-05-15 15:23 | Review R008 | plan Step 4: APPROVE |
| 2026-05-15 15:29 | Review R009 | code Step 4: APPROVE |

| 2026-05-15 16:01 | Worker iter 1 | done in 3772s, tools: 97 |
| 2026-05-15 16:01 | Step 5 started | Verify |

| 2026-05-15 16:49 | Worker iter 2 | done in 2891s, tools: 6 |
| 2026-05-15 17:03 | Step 5 grep evidence | `grep -n 'defer .*Body\\.Close\\|defer ' internal/intervals/client.go` found only `defer timer.Stop()`; Python for-block scan returned OK. |
| 2026-05-15 17:04 | Step 5 go doc evidence | `go doc ./internal/intervals` listed `Client` with `func NewClient(opts Options) (*Client, error)` and no exported `*Client` method signature drift versus the pre-change expectation. |
| 2026-05-15 17:07 | Exit intercept timeout | Supervisor did not respond within 60s — closing session |
| 2026-05-15 17:07 | ⚠️ Steering | You are on TP-045, not TP-043. Finish Step 5 verification now: run the grep/check that confirms there is no `defer` inside a `for` block in `internal/intervals/client.go`, run `go doc ./internal/inter |
| 2026-05-15 17:07 | Worker iter 3 | done in 1093s, tools: 7 |
| 2026-05-15 17:07 | Task complete | .DONE created |