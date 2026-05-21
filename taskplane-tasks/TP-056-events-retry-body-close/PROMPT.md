# TP-056 — Close response body inside `events.go` retry loop (audit Critical)

## Mission

`internal/intervals/events.go:239` places `defer resp.Body.Close()` inside the `for attempt := 1; ; attempt++` retry loop. Because `defer` schedules execution at function return — not loop-iteration end — every non-final iteration leaks an open body until the function eventually returns. On a hot retry path (5xx storm, network blip) this accumulates file descriptors and TCP connections.

Compare with the correct pattern at `internal/intervals/client.go:236-243`, which closes the body inline (drain + close) before continuing the loop.

Fix: replace the in-loop `defer` with explicit drain (`io.Copy(io.Discard, resp.Body)`) + `resp.Body.Close()` at the end of each non-success iteration; keep `defer` only outside the loop or on the terminal success branch.

This was identified in the 2026-05-16 Go audit follow-up as the single Critical finding.

PRD anchors: §7.4 reliability.
CLAUDE.md hard rules: "Always close response bodies."

Complexity: Blast radius 1 (single file, tightly scoped), Pattern novelty 1, Security 2 (resource leak), Reversibility 1 = 5 → Review Level 1. Size: XS.

## Dependencies

- None. Coordinate loosely with **TP-045** (which hardens the analogous loop in `client.go`) — if TP-045 lands first, mirror its exact close pattern here for consistency.

## Context to Read First

- `CLAUDE.md` — HTTP / resource-handling rules.
- `internal/intervals/events.go:200-280` — the offending retry loop.
- `internal/intervals/client.go:200-260` — the correct inline-close pattern to mirror.
- `internal/intervals/events_test.go` — existing retry tests to verify nothing regresses.

## File Scope

- `internal/intervals/events.go` — fix the defer placement.
- `internal/intervals/events_test.go` — add a regression test that retries 3+ times and asserts no body leak (e.g., count `Body.Close()` calls on a test `io.ReadCloser`).
- `CHANGELOG.md`, `STATUS.md`.

Out of scope:
- Wider refactor of the retry loop (covered by TP-045 for the sibling file; do not generalize here).
- Touching retry policy, jitter, or status-decision logic.
- Renaming or moving the function.

## Steps

### Step 1: Reproduce the leak in a test

- [ ] Add a test using an `httptest.Server` that returns 503 three times then 200. Wrap the response body in a counting `io.ReadCloser` that records `Close()` calls. Assert the count equals the number of attempts.
- [ ] Confirm the test fails on `main`.

### Step 2: Fix

- [ ] Replace in-loop `defer resp.Body.Close()` with explicit drain + close on the retry branch; keep a single `defer` after the loop on the success path (or close-and-return inline).
- [ ] Mirror the exact pattern from `internal/intervals/client.go:236-243`.
- [ ] `make build` / `test` / `test-race` / `lint`.
- [ ] Commit: `TP-056 close response body inside events retry loop`.

### Step 3: Verify

- [ ] New test passes.
- [ ] `make test-race` clean.
- [ ] `grep -n "defer resp.Body.Close" internal/intervals/` shows no `defer` inside any retry loop.

## Acceptance Criteria

- No `defer resp.Body.Close()` remains inside any `for attempt := …` loop in `internal/intervals/`.
- Regression test asserts `Body.Close()` is called once per attempt.
- All `make` checks pass.
- No user-visible behaviour change.

## Do NOT

- Do not refactor the surrounding retry logic; that's TP-045 + TP-057.
- Do not change retry counts, backoff, or status-decision rules.
- Do not introduce `io.LimitReader` here (that's TP-045's scope for the other client method).

## Documentation

- `STATUS.md`
- `CHANGELOG.md` `[Unreleased]` under "Fixed".

## Git Commit Convention

Conventional Commits, prefixed `TP-056`. Example: `TP-056 close response body inside events retry loop`.

---

## Amendments

_Add amendments below this line only._
