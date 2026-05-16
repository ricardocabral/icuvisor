# TP-057 — `ctx` first on `shouldRetryWrite` + consolidate retry-decision logic (audit High)

## Mission

The 2026-05-16 Go audit flagged two related issues in `internal/intervals/`:

1. **`ctx` is not the first argument** in `internal/intervals/events.go:247`
   `func (c *Client) shouldRetryWrite(method string, ctx context.Context, attempt int) bool` — violates Go convention and CLAUDE.md ("every function that does I/O or blocks takes `ctx context.Context` as the first argument").

2. **Duplicated retry-decision logic** across `internal/intervals/client.go` and `internal/intervals/events.go`: `shouldRetry`, `shouldRetryTransport`, `shouldRetryStatus`, `shouldRetryWrite`, `shouldRetryWriteStatus`. Five near-identical helpers reading the same fields. Collapse into a single method on `*Client` that takes `(ctx, method, resp, err, attempt)` and returns `(retry bool, wait time.Duration)`.

Goal: one retry-decision method, `ctx`-first, with table-driven tests covering: idempotent GET vs non-idempotent POST/PATCH/DELETE, transport errors, 408/425/429/5xx status set, `Retry-After` honouring, attempt cap.

PRD anchors: §7.4 reliability.
CLAUDE.md hard rules: `ctx` first; one obvious place for each decision; sentinel errors for stable contract points.

Complexity: Blast radius 2 (touches every retry call site in the intervals package), Pattern novelty 1, Security 1, Reversibility 1 = 5 → Review Level 1. Size: S.

## Dependencies

- **TP-045** — coordinate. TP-045 splits `doJSONQuery` retry/transport/status concerns; this task takes the cleaned-up pieces and folds them under a single decision method. If TP-045 lands first, build on its primitives. If this lands first, leave breadcrumbs for TP-045 to pick up.
- **TP-056** — soft. Both touch `events.go`; sequence after TP-056 to keep the diff readable.

## Context to Read First

- `CLAUDE.md` — Go conventions section on `ctx` argument position.
- `internal/intervals/client.go:280-340` — `shouldRetry`, `shouldRetryTransport`, `shouldRetryStatus`, jitter helper.
- `internal/intervals/events.go:240-280` — `shouldRetryWrite`, `shouldRetryWriteStatus`.
- `internal/intervals/retry_test.go` (if present) — existing coverage.

## File Scope

- `internal/intervals/client.go` — define the consolidated `decideRetry(ctx, method, resp, err, attempt)`; delete or thin the four helpers.
- `internal/intervals/events.go` — delete `shouldRetryWrite`/`shouldRetryWriteStatus`; route through the consolidated method.
- `internal/intervals/retry_test.go` (new or extended) — table-driven coverage.
- `CHANGELOG.md`, `STATUS.md`.

Out of scope:
- Jitter generator (covered by TP-058).
- Body-close fix (covered by TP-056).
- `LimitReader` / response-size hardening (covered by TP-045).
- Changing retry counts, backoff base, or status-code set without explicit justification in `STATUS.md`.

## Steps

### Step 1: Capture current behaviour in a table-driven test
- [ ] Enumerate every `(method, status, err, attempt)` combination currently handled by the five helpers.
- [ ] Write a table that calls each existing helper and locks in the current truth table.
- [ ] Test must pass against `main` before refactoring.

### Step 2: Introduce the consolidated method
- [ ] Add `(c *Client) decideRetry(ctx context.Context, method string, resp *http.Response, err error, attempt int) (retry bool, wait time.Duration)` in `client.go`.
- [ ] Method must respect `RetryConfig`, `Retry-After`, idempotency by HTTP method, and `ctx.Done()`.
- [ ] `ctx` must be the first argument everywhere it appears.

### Step 3: Route all call sites through the new method
- [ ] Replace `shouldRetry*` call sites in `client.go` and `events.go`.
- [ ] Delete the now-unused helpers.
- [ ] Run the table-driven test from Step 1 — must still pass with the consolidated method.
- [ ] `make build` / `test` / `test-race` / `lint`.
- [ ] Commit: `TP-057 consolidate intervals retry decision into one method`.

### Step 4: Verify
- [ ] `grep -n "shouldRetry" internal/intervals/` shows only the consolidated method (or zero hits if renamed).
- [ ] `grep -n "ctx context.Context" internal/intervals/events.go` shows `ctx` always first.
- [ ] `make test-race` clean.

## Acceptance Criteria

- One retry-decision method (`decideRetry` or similar) replaces the five `shouldRetry*` helpers.
- `ctx context.Context` is the first parameter on every function that takes it in `internal/intervals/`.
- Behaviour is byte-identical on the truth table (locked-in by Step 1 test).
- `make` checks all pass.
- No public API change to `*Client` consumers.

## Do NOT

- Do not change retry policy (counts, base backoff, status-code set, idempotency rules) without an explicit `STATUS.md` note and reviewer sign-off.
- Do not absorb jitter logic here — TP-058 owns that.
- Do not generalize beyond the intervals package; this is not the time for a generic `httpretry` extraction.

## Documentation

- `STATUS.md`
- `CHANGELOG.md` `[Unreleased]` under "Changed".

## Git Commit Convention

Conventional Commits, prefixed `TP-057`.

---

## Amendments

_Add amendments below this line only._
