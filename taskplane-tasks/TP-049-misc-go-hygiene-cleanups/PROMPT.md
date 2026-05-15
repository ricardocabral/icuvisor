# TP-049 — Misc Go hygiene cleanups (audit P2 bundle)

## Mission

A grab-bag of small Go-hygiene fixes flagged by the 2026-05-15 audit. Land as a single small PR; each item is mechanical and individually uninteresting, but together they pay down low-grade noise across the tree. No item changes user-visible behaviour.

The five items:

1. **Extract panic→error recover helper in MCP server.** `internal/mcp/server.go:260-267, 296-300, 383-387` wraps SDK calls in `defer recover()` for panic-to-error conversion in three near-identical places. Extract one helper (e.g., `withPanicRecovery(name string, fn func() error) error`) with a short doc comment on _why_ it exists (defending against SDK panics at the protocol boundary).
2. **Propagate `ctx` through toolchecks `Register`.** `internal/toolchecks/schema_stability.go:46` and `internal/toolchecks/confusable_names.go:38` pass `context.Background()` into `Register`. CLAUDE.md forbids `context.TODO`; `Background()` is fine at program root but these are libraries called from the registry. Plumb `ctx` through from the caller.
3. **Move env read out of `internal/app`.** `internal/app/app.go:67` calls `response.DebugMetadataFromEnv()` inside `Run`; environment lookups belong in `internal/config` (`config.Load`). Move the resolution; have `Run` consume the resolved value from `config.Config`.
4. **Long-line constructor formatting.** `internal/tools/get_fitness.go:210-224` four `newXxxTool` constructors exceed 200 columns each; wrap arguments across multiple lines per gofmt-friendly style.
5. **Fix misleading error message.** `internal/tools/registry.go:67-70` error string references `getAthleteProfileName` hard-coded for non-profile failures. (May overlap with TP-042 — if TP-042 lands first, this item can be a no-op; flag in `STATUS.md`.)

PRD anchors: §7.4 reliability. CLAUDE.md hard rules: `ctx` first on every I/O function; no env reads outside `internal/config`.

ROADMAP positioning: pure maintenance / debt paydown; independent of any milestone.

Complexity: Blast radius 1 (small touches, no protocol changes), Pattern novelty 1, Security 1, Reversibility 1 = 4 → Review Level 1. Size: S.

## Dependencies

- **TP-042** — soft coordination on item 5 (the error-message fix). If TP-042 lands first the fix may already be in. Not a blocker; flag in `STATUS.md`.

## Context to Read First

- `CLAUDE.md` — Go conventions, `ctx` first, no env reads outside `config`, no `panic` outside `main`.
- `internal/mcp/server.go` — the three recover blocks at lines 260-267, 296-300, 383-387.
- `internal/toolchecks/schema_stability.go`, `internal/toolchecks/confusable_names.go` — the `context.Background()` call sites and the `Register` signatures.
- `internal/app/app.go` — the env-read inside `Run`.
- `internal/config/config.go` — where the env read should move.
- `internal/response/shaper.go` — `DebugMetadataFromEnv` definition.
- `internal/tools/get_fitness.go` — the over-long constructor lines.
- `internal/tools/registry.go` — the misleading error string.

## File Scope

- `internal/mcp/server.go` — extract the recover helper; collapse three call sites to use it.
- `internal/toolchecks/schema_stability.go`, `internal/toolchecks/confusable_names.go`, and their tests — accept and propagate `ctx`.
- `internal/app/app.go`, `internal/config/config.go`, `internal/response/shaper.go` — move env read into config; remove `DebugMetadataFromEnv` (or relocate it to `internal/config`). Plumb the resolved value through `config.Config`.
- `internal/tools/get_fitness.go` — reformat the four constructors.
- `internal/tools/registry.go` — fix the error string (skip if TP-042 already fixed it).
- `CHANGELOG.md`, `STATUS.md`.

Out of scope:
- Any item not in the five above.
- Wider style sweeps (line-length, ctx-plumbing, env-read audits) beyond the cited sites.
- Changing `Config` exported fields beyond what item 3 strictly requires.
- Touching tool schemas or names.

File follow-ups in `CONTEXT.md` "Technical Debt" if more hygiene issues surface during the work.

## Steps

### Step 1: Recover helper (item 1)

- [ ] Add `withPanicRecovery(name string, fn func() error) error` (or equivalent) in `internal/mcp/server.go` (or a small new file in the same package). Doc comment must say _why_ it exists (SDK protocol-boundary panics).
- [ ] Replace each of the three `defer recover()` blocks with a call to the helper.
- [ ] `make build` / `test` / `test-race` / `lint`.
- [ ] Commit: `TP-049 extract MCP panic-recovery helper`.

### Step 2: Propagate `ctx` through toolchecks (item 2)

- [ ] Change `toolchecks.Register` (in both `schema_stability.go` and `confusable_names.go`) to accept `ctx context.Context` as the first argument.
- [ ] Update the registry call site to pass through the existing `ctx`.
- [ ] Update tests.
- [ ] `make build` / `test` / `test-race` / `lint`.
- [ ] Commit: `TP-049 propagate ctx through toolchecks Register`.

### Step 3: Move env read into config (item 3)

- [ ] Resolve `DebugMetadata` inside `config.Load` (read the env var there, store on `config.Config`).
- [ ] Have `app.Run` consume the resolved value from `Config` instead of calling `response.DebugMetadataFromEnv()`.
- [ ] Remove or relocate `response.DebugMetadataFromEnv`; if relocated, it lives in `internal/config`.
- [ ] `grep -rn "os.Getenv\b" internal/ | grep -v "^internal/config/"` should be empty (or remaining hits justified in `STATUS.md`).
- [ ] `make build` / `test` / `test-race` / `lint`.
- [ ] Commit: `TP-049 move DebugMetadata env read into config`.

### Step 4: Reformat long constructor lines (item 4)

- [ ] Wrap the four `newXxxTool` calls at `internal/tools/get_fitness.go:210-224` across multiple lines in gofmt-friendly style.
- [ ] Verify `gofmt`/`goimports` leaves the result clean.
- [ ] `make build` / `test` / `lint`.
- [ ] Commit: `TP-049 wrap long constructor lines in get_fitness`.

### Step 5: Fix registry error message (item 5)

- [ ] If TP-042 already fixed `internal/tools/registry.go:67-70`, mark this step a no-op in `STATUS.md` and skip.
- [ ] Otherwise, replace the hardcoded `getAthleteProfileName` reference with the failing tool's actual name.
- [ ] `make build` / `test` / `lint`.
- [ ] Commit: `TP-049 fix misleading registry error message` (if not skipped).

### Step 6: Verify

- [ ] `make build` / `make test` / `make test-race` / `make lint` all green.
- [ ] `grep -rn "os.Getenv\b" internal/ | grep -v "^internal/config/"` empty (or each remaining hit justified in `STATUS.md`).
- [ ] `grep -rn "context.Background()" internal/toolchecks/` empty.
- [ ] `git diff --stat` shows small, focused touches — no unrelated churn.

## Acceptance Criteria

- The three `defer recover()` blocks in `internal/mcp/server.go` collapse to a single helper with a doc comment explaining purpose.
- `toolchecks.Register` (both variants) accepts `ctx`; no `context.Background()` remains in `internal/toolchecks/`.
- No `os.Getenv` outside `internal/config/` (or each remaining hit explicitly justified in `STATUS.md`).
- `response.DebugMetadataFromEnv` is removed or moved into `internal/config`; `internal/app` no longer reads env directly.
- The four long-line constructors in `internal/tools/get_fitness.go` are wrapped within gofmt-friendly style.
- The registry error message at `internal/tools/registry.go:67-70` no longer mis-attributes failures to `getAthleteProfileName` (or noted as already fixed by TP-042).
- All checks pass: `make build` / `test` / `test-race` / `lint`.
- No tool schema, tool name, or user-visible behaviour change.

## Do NOT

- Do not convert this into a wider style / ctx / env-read sweep. Five items, scoped as listed.
- Do not change `Config`'s exported fields beyond what item 3 requires.
- Do not touch tool schemas, names, or `_meta` output.
- Do not panic anywhere; preserve error wrapping with `%w`.
- Do not introduce new dependencies.

## Documentation

- `STATUS.md`
- `CHANGELOG.md` `[Unreleased]` under "Changed" (internal refactor; no user-visible change).

## Git Commit Convention

One commit per item, each prefixed `TP-049`. Examples in Steps 1-5 above.

---

## Amendments

_Add amendments below this line only._
