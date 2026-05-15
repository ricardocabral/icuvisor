# TP-043 — Remove process-global mutable state from `internal/response/shaper.go` (audit P1)

## Mission

`internal/response/shaper.go:18-24` keeps delete-mode and toolset in package-level `atomic.Value` slots initialized via `init()` and written by `SetDeleteMode` / `SetToolset` (called from `main`). Read deep inside `addCommonMeta` (~line 291). This is hidden global mutable state that:

- makes tests interfere across packages (one test's `SetDeleteMode` leaks into another's),
- violates CLAUDE.md's preferences for explicit dependencies and `ctx`-threaded state,
- and is the only `init()` in `./internal` (per the audit grep).

`response.Options` already exists as the natural carrier. Thread delete-mode and toolset through it, populate at the call site in `internal/app/app.go` / registry construction, and delete the globals + `init()` + `Set*` setters.

PRD anchors: §7.4 #6 (clear safety-mode signaling in responses — must be preserved bit-for-bit), §7.2.E `_meta` envelope.

Complexity: Blast radius 2 (touches every `_meta`-emitting tool, but mechanically), Pattern novelty 1, Security 1, Reversibility 2 = 6 → Review Level 2. Size: S.

## Dependencies

- None blocking. Coordinate with TP-047 if both run in parallel (both touch `shaper.go`); land TP-043 first if so.

## Context to Read First

- `CLAUDE.md` — explicit deps, no globals, `ctx` first.
- `internal/response/shaper.go` — the file.
- `internal/response/doc.go` — pkg overview.
- `internal/response/shaper_test.go` — current tests; check for any that depend on the global being set.
- `internal/app/app.go` — call site for `SetDeleteMode`/`SetToolset`.
- `internal/safety/mode.go`, `internal/safety/toolset.go` — sources of truth.

## File Scope

- `internal/response/shaper.go` — remove globals, `init()`, `SetDeleteMode`, `SetToolset`. Add fields to `Options`. Update `addCommonMeta` to read from `Options`.
- `internal/response/shaper_test.go` — remove any `Set*` test helpers; pass via `Options`.
- `internal/app/app.go` — pass values via `Options` (or via the registry, which then constructs `Options`).
- `internal/tools/registry.go` — if it's the place `Options` is built, plumb through there.
- `CHANGELOG.md`.

## Steps

### Step 1: Audit reads

- [ ] `grep -rn "processDeleteMode\|processToolset\|SetDeleteMode\|SetToolset" internal/` to find every reader/writer.
- [ ] Decide where `Options` is constructed (registry, per-call, or shared singleton inside the registry).

### Step 2: Refactor

- [ ] Add `DeleteMode` and `Toolset` fields to `response.Options` with their existing types.
- [ ] Update `addCommonMeta` (and any other reader) to consume from `Options`.
- [ ] Delete the package-level `atomic.Value` slots, the `init()`, and the `Set*` setters.
- [ ] Update call sites to pass values through `Options`.

### Step 3: Tests

- [ ] Run existing shaper tests; fix any that relied on globals.
- [ ] Add a test asserting two `Options` with different delete-mode produce the expected divergent `_meta` (regression guard against globals creeping back).

### Step 4: Verify

- [ ] `make build`, `make test`, `make test-race`, `make lint`.
- [ ] Confirm `grep -rn "init()" internal/` returns no hits in `internal/response/`.
- [ ] `_meta` output byte-identical for the same inputs (compare snapshot fixtures if present).

## Acceptance Criteria

- No `init()` in `internal/response/`.
- No package-level mutable state in `internal/response/shaper.go`.
- `response.Options` carries delete-mode and toolset; consumers read from there.
- `_meta` output for the same logical inputs is bit-identical to pre-change.
- Tests no longer call `SetDeleteMode` / `SetToolset`; values arrive via `Options`.

## Do NOT

- Do not change the `_meta` JSON shape.
- Do not introduce a different global (e.g., `context.WithValue` carrying delete-mode through every call) — pass via `Options`.
- Do not move delete-mode resolution out of `internal/safety`; only its **consumption** by response shaping moves.

## Documentation

- `STATUS.md`
- `CHANGELOG.md` `[Unreleased]` under "Changed".

## Git Commit Convention

`TP-043 thread delete-mode/toolset through response.Options`, etc.

---

## Amendments

_Add amendments below this line only._
