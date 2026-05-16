# TP-058 — Non-deterministic retry jitter using `math/rand/v2` (audit High)

## Mission

`internal/intervals/client.go:317` computes jitter as

```go
offset := time.Now().UnixNano()%(2*span+1) - span
```

This is deterministic per nanosecond. Multiple goroutines retrying within the same nanosecond (likely on a 429 burst) get correlated offsets, defeating the entire point of jitter (de-correlating retries to avoid thundering-herd).

Fix: use `math/rand/v2` with a package-local `*rand.Rand` seeded once. Generate `rand.Int64N(2*span+1) - span` per call. `math/rand/v2` is safe for concurrent use as of Go 1.22.

Audit ref: 2026-05-16 Go audit, "High" severity.

PRD anchors: §7.4 reliability (no thundering herd on rate-limit storms).

Complexity: Blast radius 1, Pattern novelty 1, Security 2 (defends against rate-limit feedback loops), Reversibility 1 = 5 → Review Level 1. Size: XS.

## Dependencies

- **TP-045** — soft. Both touch the retry helpers. If TP-045's `WithDefaults` constructor lands first, route default-jitter configuration through it. Otherwise, keep this isolated to the jitter generator only.
- **TP-057** — sequence after TP-057 so the jitter generator sits cleanly under the consolidated `decideRetry`.

## Context to Read First

- `internal/intervals/client.go:300-340` — the current jitter helper.
- `internal/intervals/client.go:48-65` — `WithDefaults()` / `RetryConfig` (related, but not changed here).
- Go stdlib docs for `math/rand/v2` — concurrency safety, `Int64N`.

## File Scope

- `internal/intervals/client.go` — replace the deterministic jitter expression with a `*rand.Rand`-backed generator. Seed once at package init using `rand.NewPCG(uint64(time.Now().UnixNano()), uint64(os.Getpid()))` (or equivalent).
- `internal/intervals/client_test.go` — add a statistical test: 1000 jitter samples must produce ≥ 90% unique values within the span.
- `CHANGELOG.md`, `STATUS.md`.

Out of scope:
- Changing the jitter span / formula. Same span, just a real RNG.
- Renaming `RetryConfig` fields.
- The jitter-default-only-when-all-fields-zero quirk (covered by TP-045's `WithDefaults`).

## Steps

### Step 1: Statistical test
- [ ] Add a test that calls the jitter helper 1000 times and asserts uniqueness ≥ 900.
- [ ] Confirm the test fails on `main` (current deterministic helper produces too many collisions when called in tight loop).

### Step 2: Switch to `math/rand/v2`
- [ ] Replace the `time.Now().UnixNano()%...` expression with `pkgRand.Int64N(2*span+1) - span`.
- [ ] Initialize `pkgRand` once at package level using a `sync.Once` or `var` init.
- [ ] `math/rand/v2`'s top-level functions are already concurrency-safe; using them directly is also acceptable and simpler. Prefer the top-level form unless deterministic seeding is needed for tests.

### Step 3: Verify
- [ ] New statistical test passes.
- [ ] `make build` / `test` / `test-race` / `lint`.
- [ ] No new `math/rand` (v1) import anywhere; only `math/rand/v2`.
- [ ] Commit: `TP-058 use math/rand/v2 for retry jitter`.

## Acceptance Criteria

- Jitter helper uses `math/rand/v2`.
- No `time.Now()` is the only entropy source for jitter.
- Statistical test gates against regressions.
- `make` checks pass.
- Go module requires Go 1.22+ (verify `go.mod`; if not, bump it).

## Do NOT

- Do not import `math/rand` (v1).
- Do not change the jitter span or the backoff base.
- Do not introduce a third-party RNG library.

## Documentation

- `STATUS.md`
- `CHANGELOG.md` `[Unreleased]` under "Changed".

## Git Commit Convention

Conventional Commits, prefixed `TP-058`.

---

## Amendments

_Add amendments below this line only._
