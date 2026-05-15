# TP-045 — Harden `internal/intervals/client.go` `doJSONQuery` (audit P1)

## Mission

`internal/intervals/client.go:115-163` `doJSONQuery` mixes retry, transport, status handling, body close, and JSON decode into a single loop. Two specific hazards from the 2026-05-15 Go audit:

1. `defer resp.Body.Close()` sits inside `for attempt := 1; ;` (line ~157). In the current control flow it executes once, but the structure is fragile — a future `continue` added after that line would leak the body. `defer` inside a loop is a known footgun.
2. The successful-path body is not bounded with `io.LimitReader` before `json.NewDecoder(resp.Body).Decode`, allowing a misbehaving upstream to OOM the process.

Plus a smaller item from the same audit: `normalizeRetryConfig` (~line 165-183) uses the subtle `useDefaultJitter := cfg == (RetryConfig{})` early-capture trick; replace with an explicit `WithDefaults` constructor (or method on `RetryConfig`).

Split `doJSONQuery` into:

- `do(ctx) (*http.Response, error)` — single attempt + body lifecycle.
- `readBody(io.Reader) ([]byte, error)` — bounded read via `io.LimitReader`.
- `shouldRetry(*http.Response, error) bool` — classification only.
- An outer retry wrapper that owns backoff + jitter.

Keep the existing public API of `*intervals.Client` unchanged.

PRD anchors: §7.4 reliability / no panics on upstream misbehaviour.

ROADMAP positioning: maintenance / hardening; independent of any version milestone. Land before v0.5 dogfood so a flaky or malicious upstream cannot take down a forum-recruited tester.

Complexity: Blast radius 2 (all upstream HTTP calls go through this), Pattern novelty 1 (standard Go HTTP plumbing), Security 2 (resource-exhaustion fix), Reversibility 2 = 7 → Review Level 2. Size: M.

## Dependencies

- None blocking. Pure internal refactor + a bounded-body safety fix.

## Context to Read First

- `CLAUDE.md` — error wrapping, no panic outside main, `context` first, shared `*http.Client`, structured tests, no logging API keys.
- `internal/intervals/client.go` — the file under refactor.
- `internal/intervals/client_test.go` (if present) — existing retry / transport tests to mirror.
- `go.mod` — confirm the existing retry/backoff dep (`httpretry` or equivalent) and its version; do not swap it.
- Any caller in `internal/tools/` that exercises retry behaviour, to confirm the public surface stays stable.

## File Scope

- `internal/intervals/client.go` — refactor `doJSONQuery` into `do` / `readBody` / `shouldRetry` + outer retry loop; replace `normalizeRetryConfig` with `RetryConfig.WithDefaults` (or equivalent constructor).
- `internal/intervals/client_test.go` — add tests for: oversize body trips `io.LimitReader`; 429 retry; 5xx retry; 4xx no-retry; body always closed across every code path (use `httptest.Server` + a counter); retry budget exhaustion returns a wrapped error.
- `CHANGELOG.md` — `[Unreleased]` entry under "Changed" (internal hardening) and "Fixed" (unbounded body read).
- `taskplane-tasks/TP-045-intervals-client-dojsonquery-hardening/STATUS.md`.

Out of scope:

- Switching the retry library.
- Changing the retry policy (same backoff + jitter, just plumbed through cleanly).
- Changing response-body size limits anywhere other than this function. Pick a generous default (32 MiB) and make it configurable via existing `RetryConfig` / client config only if trivial — otherwise hardcode and note in `STATUS.md`.

## Steps

### Step 1: Sketch the new function boundaries

- [ ] Map the current `doJSONQuery` into the four pieces (`do`, `readBody`, `shouldRetry`, outer loop).
- [ ] Agree on a body-size cap default. Default recommendation: 32 MiB. Document the choice in `STATUS.md`.
- [ ] Confirm the existing retry policy parameters (max attempts, backoff base, jitter, retryable status set) and re-plumb them unchanged.

### Step 2: Implement the split

- [ ] Implement `do(ctx) (*http.Response, error)` — builds the request (including `User-Agent`), sends via the shared `*http.Client`, returns the response without consuming the body. Caller owns close.
- [ ] Implement `readBody(io.Reader) ([]byte, error)` — wraps the reader with `io.LimitReader(r, maxBodyBytes+1)` and returns an explicit "response too large" error (sentinel) when the cap is tripped. Wrap with `%w`.
- [ ] Implement `shouldRetry(*http.Response, error) bool` — classification only; no I/O, no sleep.
- [ ] Implement the outer retry wrapper that owns backoff + jitter and calls `do` per attempt. Close the body in the same scope that opened it, no `defer` inside the loop.
- [ ] Decode JSON from the bounded buffer, not directly from `resp.Body`.

### Step 3: Replace `normalizeRetryConfig`

- [ ] Add `func (c RetryConfig) WithDefaults() RetryConfig` (or a free `NewRetryConfig` constructor) that returns a config with explicit defaults filled in for zero-value fields. No `cfg == (RetryConfig{})` comparison.
- [ ] Update all call sites; delete `normalizeRetryConfig`.

### Step 4: Tests

- [ ] Body-close accounting: `httptest.Server` handler increments a counter on body consumption / connection close; assert the counter for success, retry-then-success, retry-budget-exhaustion, oversize-body, and context-cancelled paths.
- [ ] Oversize body: server emits `Content-Length: huge` (or chunked stream > cap); assert sentinel error and that `resp.Body` was closed.
- [ ] 429 retry: server returns 429 N times then 200; assert N+1 attempts, final value decoded.
- [ ] 5xx retry: same pattern for 503.
- [ ] 4xx no-retry: server returns 400; assert exactly one attempt and a wrapped error.
- [ ] Retry budget exhaustion: server always 503; assert max attempts hit and a wrapped error referencing the last status.
- [ ] `RetryConfig.WithDefaults`: zero value yields documented defaults; partially set value preserves explicit fields.

### Step 5: Verify

- [ ] `make build`, `make test`, `make test-race`, `make lint`.
- [ ] `git diff --stat` should show `client.go` net-roughly-similar (split into helpers), `client_test.go` growing.
- [ ] Grep for `defer resp.Body.Close()` inside any `for` block in the file — must be zero.
- [ ] Confirm no public method signature on `*intervals.Client` changed (`go doc ./internal/intervals` diff).

## Acceptance Criteria

- No `defer` inside the retry loop in `internal/intervals/client.go`.
- `io.LimitReader` (or equivalent bounded read) is used on the success-path body before JSON decode; oversize responses return a wrapped sentinel error rather than allocating unbounded memory.
- `RetryConfig` uses an explicit defaults constructor / method; no zero-value struct comparison anywhere in the package.
- Body-close-count test passes for every code path (success, retry-then-success, exhaustion, oversize, 4xx, context-cancel).
- Existing `*intervals.Client` public API preserved (no signature changes on exported methods).
- Retry policy unchanged: same max attempts, same backoff base, same jitter, same retryable status set.

## Do NOT

- Do not change public method signatures on `*intervals.Client`.
- Do not introduce a new HTTP library; reuse the existing shared `*http.Client` and existing retry dep.
- Do not log API keys or any header that might contain them.
- Do not remove the `User-Agent: icuvisor/<version>` header.
- Do not panic on oversize / malformed bodies — return wrapped errors.
- Do not move the body-size cap into a global var; pass it through the client config or keep it a package-level constant with a comment explaining the choice.

## Documentation

Must update:

- `STATUS.md`
- `CHANGELOG.md` — `[Unreleased]` under "Changed" (internal split) and "Fixed" (unbounded body read).
- A short note in the client's doc comment naming the body-size cap and pointing at the sentinel error.

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-045`, for example: `TP-045 split doJSONQuery into do/readBody/shouldRetry`.

---

## Amendments

_Add amendments below this line only._
