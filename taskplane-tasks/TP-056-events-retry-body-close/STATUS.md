# TP-056-events-retry-body-close — Status

**Current Step:** Step 3: Verify
**Status:** ✅ Complete
**Last Updated:** 2026-05-16
**Review Level:** 1
**Review Counter:** 4
**Iteration:** 1
**Size:** XS

---

### Step 1: Reproduce the leak in a test

**Status:** ✅ Complete

- [x] Add a regression test for a retried PUT event-write path using a client-side wrapping transport that returns 503 three times then 200, asserts each prior retry body is closed before the next attempt starts, and asserts final attempts/Close counts.
- [x] Run the new test against the current pre-fix checkout and record whether it fails or already passes due the existing non-2xx inline close behavior.

### Step 2: Fix

**Status:** ✅ Complete

- [x] Replace the success-path `defer resp.Body.Close()` inside the `doJSONBody` retry loop with explicit `decodeErr` / `closeErr` handling so the body closes before every return, including decode errors, without calling `readBody` or adding response-size limiting.
- [x] Preserve the existing non-2xx retry-branch drain-and-close pattern from `internal/intervals/client.go` without changing retry eligibility, `Retry-After`, or ignored-close-error-on-retry behavior.
- [x] Add a decode-error regression case with a counting body that proves a 200 invalid-JSON response is closed when `AddOrUpdateEvent` returns an error.
- [x] Run `make build`, `make test`, `make test-race`, and `make lint` successfully.
- [x] Update `CHANGELOG.md` under `[Unreleased]` / `Fixed` for the response-body close fix.

### Step 3: Verify

**Status:** ✅ Complete

- [x] Run the new event body-close tests and confirm they pass.
- [x] Run `make test-race` clean as the final race check.
- [x] Confirm `grep -n "defer resp.Body.Close" internal/intervals/` shows no defers inside retry loops.

| 2026-05-16 20:44 | Task started | Runtime V2 lane-runner execution |
| 2026-05-16 20:44 | Step 1 started | Reproduce the leak in a test |
| 2026-05-16 20:44 | Step 1 verification | `go test ./internal/intervals -run TestAddOrUpdateEventClosesRetryResponseBodyBeforeNextAttempt -count=1` passed on the current checkout; review R001 noted events.go already has inline drain+close on non-2xx retry responses, so the retry-branch leak is not reproducible here. |
| 2026-05-16 20:44 | Step 2 started | Fix |
| 2026-05-16 20:44 | Step 2 checks | `make build && make test && make test-race && make lint` passed. |
| 2026-05-16 20:44 | Step 3 started | Verify |
| 2026-05-16 20:44 | Step 3 targeted tests | `go test ./internal/intervals -run 'TestAddOrUpdateEventCloses(RetryResponseBodyBeforeNextAttempt|ResponseBodyOnDecodeError)' -count=1` passed. |
| 2026-05-16 20:44 | Step 3 race tests | `make test-race` passed. |
| 2026-05-16 20:44 | Step 3 grep | `grep -n "defer resp.Body.Close" internal/intervals/*.go || true` returned no matches. |
| 2026-05-16 20:47 | Review R001 | plan Step 1: REVISE |
| 2026-05-16 20:50 | Review R002 | plan Step 1: APPROVE |
| 2026-05-16 20:54 | Review R003 | plan Step 2: REVISE |
| 2026-05-16 20:55 | Review R004 | plan Step 2: APPROVE |

| 2026-05-16 21:00 | Worker iter 1 | done in 933s, tools: 71 |
| 2026-05-16 21:00 | Task complete | .DONE created |