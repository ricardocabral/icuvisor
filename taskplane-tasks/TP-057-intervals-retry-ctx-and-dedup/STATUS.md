# TP-057-intervals-retry-ctx-and-dedup — Status

**Current Step:** Step 4: Verify
**Status:** ✅ Complete
**Last Updated:** 2026-05-16
**Review Level:** 1
**Review Counter:** 5
**Iteration:** 3
**Size:** S

---

### Step 1: Capture current behaviour in a table-driven test

**Status:** ✅ Complete

- [x] Enumerate every `(method, status, err, attempt)` combination currently handled by the five helpers.
- [x] Include effective read-path and no-body DELETE call-site decisions, with active/canceled context and attempt-cap cases.
- [x] Lock current status behaviour, including 408/425 as non-retryable and 429/5xx as retryable where the existing policy allows.
- [x] Cover deterministic retry delays and `Retry-After` handling with jitter disabled.
- [x] Write a shared truth table with a Step 1 adapter for existing helpers that can later be switched to the consolidated method.
- [x] Test must pass against `main` before refactoring.

### Step 2: Introduce the consolidated method

**Status:** ✅ Complete

- [x] Add consolidated `decideRetry(ctx, method, resp, err, attempt)` in `client.go` returning retry and wait.
- [x] Preserve current retry policy for transport errors, status codes, HTTP methods, attempt caps, and context cancellation.
- [x] Return deterministic waits through existing retry-delay/Retry-After logic without changing jitter behaviour.
- [x] Keep every `ctx context.Context` parameter first in the new method signature.

### Step 3: Route all call sites through the new method

**Status:** ✅ Complete

- [x] Replace retry decision call sites in `client.go`, `events.go`, and `workout_library.go` with `decideRetry` while preserving no-body DELETE behaviour.
- [x] Route all retrying call sites through the `wait` returned by `decideRetry`, preserving response close/drain ordering and context-cancellation error behaviour.
- [x] Retarget the Step 1 truth-table adapter to validate `decideRetry` without references to old helpers.
- [x] Delete the now-unused `shouldRetry*` helpers.
- [x] Update `CHANGELOG.md` under `[Unreleased]` for the retry-decision consolidation.
- [x] Run the Step 1 table-driven retry test successfully against the consolidated method.
- [x] Run `make build`, `make test`, `make test-race`, and `make lint` successfully.

### Step 4: Verify

**Status:** ✅ Complete

- [x] Verify `shouldRetry` references are removed from `internal/intervals/`.
- [x] Verify every `ctx context.Context` parameter in `internal/intervals/events.go` is first.
- [x] Run `make test-race` successfully as the final verification gate.

| 2026-05-16 21:00 | Task started | Runtime V2 lane-runner execution |
| 2026-05-16 21:00 | Step 1 started | Capture current behaviour in a table-driven test |
| 2026-05-16 21:03 | Review R001 | plan Step 1: UNKNOWN |
| 2026-05-16 21:05 | Review R002 | plan Step 1: APPROVE |

| 2026-05-16 21:54 | Worker iter 1 | done in 3240s, tools: 28 |
| 2026-05-16 22:29 | Review R003 | plan Step 2: APPROVE |

| 2026-05-16 22:48 | Worker iter 2 | done in 3264s, tools: 45 |
| 2026-05-16 22:48 | Step 3 started | Route all call sites through the new method |
| 2026-05-16 22:51 | Review R004 | plan Step 3: UNKNOWN |
| 2026-05-16 22:52 | Review R005 | plan Step 3: APPROVE |

| 2026-05-16 22:58 | Worker iter 3 | done in 622s, tools: 61 |
| 2026-05-16 22:58 | Task complete | .DONE created |