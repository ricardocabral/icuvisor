# TP-045-intervals-client-dojsonquery-hardening — Status

**Current Step:** Step 1: Sketch the new function boundaries
**Status:** ⏳ Not started
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 0
**Iteration:** 0
**Size:** M

---

### Step 1: Sketch the new function boundaries

**Status:** ⏳ Not started

- [ ] Map current `doJSONQuery` into `do` / `readBody` / `shouldRetry` / outer loop
- [ ] Agree on body-size cap default (recommendation: 32 MiB)
- [ ] Confirm existing retry policy parameters and re-plumb unchanged

### Step 2: Implement the split

**Status:** ⏳ Not started

- [ ] `do(ctx) (*http.Response, error)` — single attempt, caller owns body close
- [ ] `readBody(io.Reader) ([]byte, error)` — `io.LimitReader` + sentinel for oversize
- [ ] `shouldRetry(*http.Response, error) bool` — classification only
- [ ] Outer retry wrapper — no `defer` inside the loop
- [ ] JSON decode from bounded buffer, not raw `resp.Body`

### Step 3: Replace `normalizeRetryConfig`

**Status:** ⏳ Not started

- [ ] Add `RetryConfig.WithDefaults` (or `NewRetryConfig`)
- [ ] Update all call sites
- [ ] Delete `normalizeRetryConfig` and the zero-value comparison

### Step 4: Tests

**Status:** ⏳ Not started

- [ ] Body-close accounting across all paths (success, retry-then-success, exhaustion, oversize, 4xx, ctx-cancel)
- [ ] Oversize body trips `io.LimitReader` and returns sentinel
- [ ] 429 retry path
- [ ] 5xx retry path
- [ ] 4xx no-retry path
- [ ] Retry budget exhaustion
- [ ] `RetryConfig.WithDefaults` behaviour (zero / partial)

### Step 5: Verify

**Status:** ⏳ Not started

- [ ] `make build` / `test` / `test-race` / `lint`
- [ ] Grep confirms no `defer` inside a `for` block in `client.go`
- [ ] `go doc ./internal/intervals` shows no public signature drift

---

## Decisions

_Record body-size cap default (recommendation 32 MiB) and whether it is configurable via `RetryConfig` / client config or a package constant — settle in Step 1._

## Notes

_Add notes as work progresses._
