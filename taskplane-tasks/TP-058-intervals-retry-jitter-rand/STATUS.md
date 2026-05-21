# TP-058-intervals-retry-jitter-rand — Status

**Current Step:** Step 3: Verify
**Status:** ✅ Complete
**Last Updated:** 2026-05-16
**Review Level:** 1
**Review Counter:** 2
**Iteration:** 1
**Size:** XS

---

### Step 1: Statistical test

**Status:** ✅ Complete

- [x] Add a test that calls the jitter helper 1000 times and asserts uniqueness ≥ 900.
- [x] Confirm the test fails on `main` (current deterministic helper produces too many collisions when called in tight loop). Evidence: `go test ./internal/intervals -run TestAddJitterProducesUniqueSamples -count=1` failed with 115 unique samples.

### Step 2: Switch to `math/rand/v2`

**Status:** ✅ Complete

- [x] Replace the `time.Now().UnixNano()%...` expression with top-level `rand.Int64N(2*span+1) - span` from `math/rand/v2`.
- [x] Avoid package-local `*rand.Rand` initialization because `math/rand/v2` documents only top-level functions as concurrency-safe without extra synchronization.
- [x] `math/rand/v2`'s top-level functions are already concurrency-safe; using them directly is also acceptable and simpler. Prefer the top-level form unless deterministic seeding is needed for tests.

### Step 3: Verify

**Status:** ✅ Complete

- [x] New statistical test passes. Evidence: `go test ./internal/intervals -run TestAddJitterProducesUniqueSamples -count=1`.
- [x] `make build` / `test` / `test-race` / `lint`. Evidence: `make build && make test && make test-race && make lint` passed after adding a documented gosec suppression for non-cryptographic retry jitter.
- [x] No new `math/rand` (v1) import anywhere; only `math/rand/v2`. Evidence: grep found only `internal/intervals/client.go: "math/rand/v2"`.
- [x] Update `CHANGELOG.md` `[Unreleased]` under Changed and verify `go.mod` is Go 1.22+. Evidence: added Changed bullet; `go.mod` declares `go 1.25.10`.
- [x] Commit: `TP-058 use math/rand/v2 for retry jitter`. Evidence: committed `fix(TP-058): use math/rand/v2 for retry jitter` (`39df8ba`).

| 2026-05-16 22:58 | Task started | Runtime V2 lane-runner execution |
| 2026-05-16 22:58 | Step 1 started | Statistical test |
| 2026-05-16 23:02 | Review R001 | plan Step 1: APPROVE |
| 2026-05-16 23:06 | Review R002 | plan Step 2: APPROVE |

| 2026-05-16 23:10 | Worker iter 1 | done in 712s, tools: 65 |
| 2026-05-16 23:10 | Task complete | .DONE created |