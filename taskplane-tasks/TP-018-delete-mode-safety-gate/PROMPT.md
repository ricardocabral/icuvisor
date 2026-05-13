# TP-018 — `ICUVISOR_DELETE_MODE` safety gate (registration-time filtering)

## Mission

Build the v0.3 safety model: a single env var, `ICUVISOR_DELETE_MODE`, that decides whether destructive tools are even registered with the MCP server. No per-call `confirm: true` arguments anywhere in the catalog. An LLM that cannot see a tool cannot be talked into calling it.

Roadmap items (ROADMAP.md v0.3):

- `ICUVISOR_DELETE_MODE` env var (`safe` default / `full` / `none`).
- Destructive tools are not *registered* in modes that forbid them.
- No per-call `confirm: true` arguments anywhere in the catalog.

PRD anchors: §7.2.D safety model, §7.4 write-path safety, CLAUDE.md "Destructive ops require explicit confirmation" — superseded for v0.3 by registration-time gating (update CLAUDE.md accordingly).

Complexity: Blast radius 3 (every v0.3 write tool depends on this), Pattern novelty 2, Security 3, Reversibility 2 = 10 → Review Level 2. Size: M.

## Dependencies

- **TP-007** — response shaping primitives (reuse `_meta.server_version` chokepoint; expose mode in `_meta` where helpful)

## Context to Read First

- `CLAUDE.md` (esp. destructive-op rule)
- `docs/prd/PRD-icuvisor.md` §7.2.C catalog, §7.2.D safety model, §7.4
- `ROADMAP.md` v0.3
- `internal/mcp/` server wiring + tool registry
- `internal/config/` for env-var conventions

## File Scope

Expected files:

- `internal/safety/` — new package owning the mode enum, env-var parsing, and the `Capability` decision API (`CanDelete`, `CanWrite`, etc.)
- `internal/safety/*_test.go`
- `internal/mcp/` — registry plumbing reads the mode at startup and filters tool registration
- `internal/config/` — env-var loader entry
- `README.md` — documented env var
- `CHANGELOG.md`
- `CLAUDE.md` — update the destructive-op rule to reference registration-time gating
- `taskplane-tasks/TP-018-delete-mode-safety-gate/STATUS.md`

Do **not** add write or delete tools here. Only the gate and the registry hook.

## Steps

### Step 1: Define the mode enum and parsing

- [ ] Enum: `safe` (default — writes allowed, deletes forbidden), `full` (writes + deletes), `none` (read-only; no write or delete tools registered)
- [ ] Env-var parsing case-insensitive; unknown / empty → `safe`; log the resolved mode once at startup at INFO
- [ ] Reject `confirm`-style per-call arguments at design time — the package exposes no helper for them

### Step 2: Capability API

- [ ] `safety.Capability` with `CanDelete() bool`, `CanWrite() bool`, `Mode() string`
- [ ] Single source of truth read once at startup; safe for concurrent reads
- [ ] Test matrix: every (mode × capability) pair

### Step 3: Registry filtering

- [ ] Tool registration takes a `safety.Capability`; destructive tools self-declare `RequiresDelete()` / `RequiresWrite()`
- [ ] Tools whose requirement the mode forbids are **not** registered (not registered-and-erroring — absent from the catalog entirely)
- [ ] Log a single INFO line at startup listing registered/skipped tool counts; never log tool names that could leak roadmap state — counts only

### Step 4: `_meta` surfacing

- [ ] Add `_meta.delete_mode` to every response from a single chokepoint in `internal/response`
- [ ] No per-call override; mode is process-global

### Step 5: Docs + CLAUDE.md update

- [ ] README: a short section documenting `ICUVISOR_DELETE_MODE` with the three modes and the default
- [ ] CLAUDE.md: replace the per-call `confirm: true` rule with the registration-time gate rule; cross-link to `internal/safety`
- [ ] CHANGELOG `[Unreleased]` entry

### Step 6: Verify

- [ ] `make test`, `make build`, `make lint`, `go test -race ./...`
- [ ] Manual: start the binary in each of `safe`, `full`, `none`; confirm the registered-tool count in logs and via `tools/list` if a smoke harness exists

## Reference Implementation Policy

- `hhopke/intervals-icu-mcp` (MIT) may be consulted for env-var ergonomics. Do not depend on it.
- `mvilanova/intervals-mcp-server` is GPLv3 — do not read, copy, paraphrase, or transliterate.

## Acceptance Criteria

- `internal/safety` exists and is the only place modes are interpreted.
- `safe` is the default; missing/invalid env values resolve to `safe`.
- Destructive tools in `safe` mode are *absent* from `tools/list`, not registered-and-erroring.
- No tool in the catalog accepts a `confirm` argument.
- `_meta.delete_mode` appears on every response.
- README, CLAUDE.md, CHANGELOG updated.

## Do NOT

- Do not implement any write or delete tool here.
- Do not introduce a `confirm: true` parameter on any tool, now or ever.
- Do not let the mode be flipped at runtime from a tool call.
- Do not log API keys or athlete IDs when logging the resolved mode.

## Documentation

Must update:

- `STATUS.md`
- `README.md`
- `CLAUDE.md`
- `CHANGELOG.md`

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-018`, for example: `TP-018 add ICUVISOR_DELETE_MODE enum and parsing`.

---

## Amendments

_Add amendments below this line only._
