# TP-030 — `ICUVISOR_TOOLSET` tiers + `icuvisor_list_advanced_capabilities`

## Mission

Cut per-session tool-description tokens (KR5) by exposing a curated `core` toolset (~17 tools) by default and gating the full surface behind `ICUVISOR_TOOLSET=full`. Ship a small `icuvisor_list_advanced_capabilities` tool in `core` so the LLM can discover what's hidden and tell the user how to enable it.

Roadmap items (ROADMAP.md v0.4):

- `ICUVISOR_TOOLSET` env var with `core` (default, ~17 tools) and `full` tiers.
- `icuvisor_list_advanced_capabilities` tool lives in `core` for discoverability when an advanced prompt arrives.

PRD anchors: §7.2.E toolset tiers, §7.2.D catalog, KR5 (§6). Builds on the registration-time filtering pattern established for `ICUVISOR_DELETE_MODE` (TP-018) — toolset tiering is a second, orthogonal registration filter.

Complexity: Blast radius 3 (touches every tool's registration), Pattern novelty 1 (extends TP-018 pattern), Security 1, Reversibility 2 = 7 → Review Level 2. Size: M.

## Dependencies

- **TP-018** — `ICUVISOR_DELETE_MODE` safety gate; reuse the registry plumbing and the env-var/`_meta` conventions. Toolset filtering composes with delete-mode filtering: a tool is registered only if **both** gates allow it.

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` §7.2.D catalog, §7.2.E toolset tiers, §6 KR5
- `ROADMAP.md` v0.4
- `internal/safety/` — the existing capability/registration pattern
- `internal/mcp/` — server wiring + tool registry
- `internal/config/` — env-var conventions

## File Scope

Expected files:

- `internal/safety/` (or a sibling `internal/toolset/`) — the `core`/`full` tier enum, env-var parsing, and a per-tool tier-membership decision API. Match the package boundary TP-018 chose; do not fork a parallel pattern.
- `internal/safety/*_test.go` (or sibling)
- `internal/mcp/` — registry plumbing reads the tier at startup and filters registration alongside the delete-mode filter
- `internal/tools/list_advanced_capabilities.go` + `_test.go` — the discoverability tool
- `internal/config/` — env-var loader entry
- `README.md` — documented env var
- `CHANGELOG.md`
- `taskplane-tasks/TP-030-toolset-tiers/STATUS.md`

## Steps

### Step 1: Tier enum and parsing

- [ ] Enum: `core` (default) and `full`; case-insensitive parsing; unknown/empty → `core`
- [ ] Log the resolved tier once at startup at INFO (count only — never leak tool names that hint roadmap state, consistent with TP-018)
- [ ] Decide and document the package boundary: extend `internal/safety` vs new `internal/toolset`. Record the choice and rationale in `STATUS.md`

### Step 2: Per-tool tier membership

- [ ] Each tool self-declares its tier (`core` or `full`); default for unmarked tools is `full` (opt-in to `core`)
- [ ] Curate the `core` set to the §7.2.E daily-use path: read activities/fitness/wellness/events, write events/wellness/messages, plus `icuvisor_list_advanced_capabilities`. Target ~17 tools; record the exact list in `STATUS.md`
- [ ] Test matrix: every tool's tier membership is asserted in a table-driven test so catalog drift is caught

### Step 3: Registry filtering composition

- [ ] Registration filters on tier **and** delete-mode; a tool appears only when both gates allow it
- [ ] Tools outside the active tier are **absent** from `tools/list`, not registered-and-erroring
- [ ] Startup INFO line reports registered/skipped counts per gate

### Step 4: `icuvisor_list_advanced_capabilities`

- [ ] Lives in `core`; returns the `full`-only tools with one-line summaries and the exact `ICUVISOR_TOOLSET=full` instruction to enable them
- [ ] Output is static/derived from the catalog — no upstream calls; terse by default
- [ ] When the tier is already `full`, it still works and says so

### Step 5: `_meta` surfacing + docs

- [ ] Add `_meta.toolset` to every response from the same chokepoint TP-018 used for `_meta.delete_mode`
- [ ] README: short section documenting `ICUVISOR_TOOLSET`, the two tiers, the default, and the discoverability tool
- [ ] CHANGELOG `[Unreleased]` entry

### Step 6: Verify

- [ ] `make test`, `make build`, `make lint`, `go test -race ./...`
- [ ] Manual: start the binary in `core` and `full`; confirm `tools/list` counts and that `icuvisor_list_advanced_capabilities` is present in `core`

## Reference Implementation Policy

- `hhopke/intervals-icu-mcp` (MIT) may be consulted for tier-curation ergonomics. Do not depend on it.
- GPL/copyleft implementation code is off limits — do not read, copy, paraphrase, or transliterate.

## Acceptance Criteria

- `ICUVISOR_TOOLSET` parsed once at startup; `core` is the default; invalid/empty → `core`.
- `core` exposes ~17 curated tools incl. `icuvisor_list_advanced_capabilities`; the exact list is pinned by a table-driven test.
- `full`-only tools are absent from `tools/list` under `core`.
- Toolset and delete-mode filters compose correctly.
- `_meta.toolset` appears on every response.
- README, CHANGELOG updated.

## Do NOT

- Do not let the tier be flipped at runtime from a tool call.
- Do not add a per-call argument that widens the catalog.
- Do not fork a parallel registration pattern — extend TP-018's.
- Do not log tool names in a way that leaks unreleased roadmap state.

## Documentation

Must update:

- `STATUS.md`
- `README.md`
- `CHANGELOG.md`

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-030`, for example: `TP-030 add ICUVISOR_TOOLSET tier enum and parsing`.

---

## Amendments

_Add amendments below this line only._
