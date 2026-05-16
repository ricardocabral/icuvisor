# TP-060 — Buffer MCP close channel + wrap recovered panics with `%w` (audit Medium)

## Mission

Two small but real defects in `internal/mcp/server.go`:

1. **Unbuffered close channel** at `internal/mcp/server.go:165-173`. `closed := make(chan error)` is unbuffered. The current control flow happens to work because the goroutine always sends, but a future early-return inside the goroutine — or a panic that bypasses the send — would hang the parent on `<-closed`. Make it `make(chan error, 1)` defensively.

2. **`withPanicRecovery` uses `%v`, not `%w`** at `internal/mcp/server.go:298-305` (introduced by TP-049). Recovered panics are converted to `fmt.Errorf("%s: %v", name, r)` — but if `r` is itself an `error`, the original is unwrappable from the recovered side. Also: no `slog` log + `debug.Stack()` capture, so on-call has no way to see what panicked. Fix: if `r` is an `error`, wrap with `%w`; either way log a `slog.Error` with `debug.Stack()` (info only, not returned to the LLM).

Audit ref: 2026-05-16 Go audit, "Medium" severity.

PRD anchors: §7.4 reliability; errors back to LLM "short, actionable, free of internal stack traces" — stack stays in logs only.

CLAUDE.md hard rules: no `panic` outside `main`; wrap with `%w`.

Complexity: Blast radius 1, Pattern novelty 1, Security 1, Reversibility 1 = 4 → Review Level 1. Size: XS.

## Dependencies

- **TP-049** — sequence after. TP-049 already extracted `withPanicRecovery` into a helper; this task hardens its body. If TP-049 has merged, just edit the helper.
- **TP-063** — soft. TP-063 splits `mcp/server.go`; if TP-063 lands first, edit the new location of `withPanicRecovery` (likely `mcp/recover.go`).

## Context to Read First

- `internal/mcp/server.go:160-180` — the close channel pattern.
- `internal/mcp/server.go:290-310` — the `withPanicRecovery` helper.
- `runtime/debug.Stack` docs.
- CLAUDE.md "Logging" section — `log/slog` rules, never log API keys.

## File Scope

- `internal/mcp/server.go` (or wherever TP-063 has moved these) — the two fixes only.
- `internal/mcp/server_test.go` — add (a) a test that the close-path doesn't hang when the inner goroutine returns without sending, (b) a test that a panicking handler produces a wrapped error AND a logged stack.
- `CHANGELOG.md`, `STATUS.md`.

Out of scope:
- Wider transport refactor (TP-063).
- Changing the recover scope (still only at the SDK protocol boundary).
- Capturing stacks for non-panic errors.

## Steps

### Step 1: Buffer the close channel
- [ ] Change `make(chan error)` to `make(chan error, 1)` at the cited line.
- [ ] Add a test that the parent does not block when the inner goroutine returns early.

### Step 2: Harden `withPanicRecovery`
- [ ] If `r` is `error`: `err = fmt.Errorf("%s: %w", name, rErr)`. Otherwise: `err = fmt.Errorf("%s: %v", name, r)` (keep `%v` only for non-error values).
- [ ] Log via `slog.Default().Error("panic recovered", "scope", name, "stack", string(debug.Stack()))`. **Never** log raw API keys or athlete IDs; log only the helper's `name` argument plus the stack.
- [ ] Add a test that asserts: (a) returned error wraps the panic when it was an error, (b) `debug.Stack()` is captured in the log (use a test handler).

### Step 3: Verify
- [ ] `make build` / `test` / `test-race` / `lint`.
- [ ] Commit: `TP-060 buffer mcp close channel and wrap recovered panics`.

## Acceptance Criteria

- `closed := make(chan error, 1)` (or equivalent buffered pattern).
- `withPanicRecovery` returns wrapped errors when the panic value was an error.
- Stack is captured into structured logs at `slog.Error`.
- No API key, token, or athlete ID is ever logged.
- All `make` checks pass.

## Do NOT

- Do not return the stack to the LLM — keep error messages short.
- Do not widen recover scope; still only at the SDK protocol boundary.
- Do not introduce a panic-handler library.

## Documentation

- `STATUS.md`
- `CHANGELOG.md` `[Unreleased]` under "Fixed" and "Changed".

## Git Commit Convention

Conventional Commits, prefixed `TP-060`.

---

## Amendments

_Add amendments below this line only._
