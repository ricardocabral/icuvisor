# TP-032 — MCP Prompts (curated entrypoints)

## Mission

Ship a curated set of MCP Prompts so users on clients that surface prompts get a "what can this thing do?" entrypoint without learning the tool catalog. Five prompts: training analysis, recovery check, weekly planning, race-week taper, coach roster triage.

Roadmap items (ROADMAP.md v0.4):

- MCP Prompts: training analysis, recovery check, weekly planning, race-week taper, coach roster triage.

PRD anchors: §7.2.G MCP Resources and Prompts, §7.2.C catalog, §7.4 #13 (clients honoring Prompts). Prompts orchestrate existing read tools and resources — they add no new upstream surface.

Complexity: Blast radius 1 (additive; no tool/transport changes), Pattern novelty 3 (first MCP Prompts), Security 1, Reversibility 1 = 6 → Review Level 1. Size: S/M.

## Dependencies

- **TP-031** — MCP Resources; prompts reference resource URIs (`icuvisor://workout-syntax`, `icuvisor://athlete-profile`) rather than inlining long-form content.
- **TP-010** — fitness / metrics reads; `training analysis` and `race-week taper` lean on these.
- **TP-011** — wellness reads; `recovery check` leans on these.
- **TP-012** — events / training-plan reads; `weekly planning` and `race-week taper` lean on these.

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` §7.2.C, §7.2.G, §7.4 #13
- `ROADMAP.md` v0.4
- `internal/mcp/` — server wiring; check the Go SDK's `prompts/list` + `prompts/get` support
- `internal/resources/` (TP-031) — resource URIs the prompts cite
- `internal/tools/` — the read tools each prompt is meant to orchestrate
- Go SDK docs for MCP Prompts (record the canonical link in `STATUS.md`)

## File Scope

Expected files:

- `internal/mcp/prompts.go` (or `internal/prompts/`) — prompt registration + handlers for `prompts/list` and `prompts/get`
- `internal/prompts/*_test.go`
- `internal/prompts/testdata/` — golden rendered prompt text
- `README.md` — document the five prompts
- `CHANGELOG.md`
- `taskplane-tasks/TP-032-mcp-prompts/STATUS.md`

## Steps

### Step 1: Prompt registration plumbing

- [ ] Wire `prompts/list` and `prompts/get` into the MCP server via the Go SDK
- [ ] One greppable registration per prompt, mirroring the tool/resource registry pattern
- [ ] Decide which prompts take arguments (e.g. date range, athlete_id for coach triage) and define minimal typed argument schemas with LLM-readable descriptions

### Step 2: Author the five prompts

- [ ] `training analysis` — guides the LLM through fitness/effort reads for a training-load + trend readout
- [ ] `recovery check` — wellness-led; surfaces sleep dual-scale, HRV/staleness, readiness
- [ ] `weekly planning` — events + training-plan reads; structures a week against planned vs completed
- [ ] `race-week taper` — events + fitness reads; taper-week framing
- [ ] `coach roster triage` — per-athlete scan; takes `athlete_id` selection, respects coach-mode conventions (the `athlete_id` argument is a selector, not a credential)
- [ ] Each prompt cites resources (`icuvisor://...`) and names the tools it expects, but does not inline long-form schema content

### Step 3: Token discipline

- [ ] Prompt bodies stay terse — they orchestrate, they do not lecture; long-form content lives in Resources (TP-031)
- [ ] Golden-file lock each rendered prompt so wording changes are deliberate

### Step 4: Docs

- [ ] README: document the five prompts and which clients surface them
- [ ] CHANGELOG `[Unreleased]` entry

### Step 5: Verify

- [ ] `make test`, `make build`, `make lint`, `go test -race ./...`
- [ ] Manual: `prompts/list` shows five; `prompts/get` renders each; confirm at least one MCP client surfaces them (note any client that ignores `prompts/list` in `STATUS.md` per §7.4 #13)

## Reference Implementation Policy

- `hhopke/intervals-icu-mcp` (MIT) ships Prompts; may be consulted for ergonomics only. Do not copy.
- `mvilanova/intervals-mcp-server` is GPLv3 — do not read, copy, paraphrase, or transliterate.

## Acceptance Criteria

- Five prompts registered and retrievable via `prompts/list` / `prompts/get`.
- Prompts that need parameters expose minimal typed argument schemas with readable descriptions.
- Prompt bodies are terse and cite Resources rather than inlining long-form content.
- Each rendered prompt is golden-file locked.
- `coach roster triage` treats `athlete_id` as a selector, not a credential.
- README, CHANGELOG updated.

## Do NOT

- Do not inline workout DSL, event-category, or custom-item schema text into prompt bodies — cite the TP-031 resources.
- Do not add new tools or upstream calls here; prompts orchestrate the existing surface.
- Do not let a prompt argument carry an API key.

## Documentation

Must update:

- `STATUS.md`
- `README.md`
- `CHANGELOG.md`

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-032`, for example: `TP-032 wire MCP prompt registration plumbing`.

---

## Amendments

_Add amendments below this line only._
