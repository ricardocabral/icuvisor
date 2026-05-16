# TP-063 — Split `internal/tools/get_fitness.go` into one-file-per-tool (audit God-module, CLAUDE.md rule violation)

## Mission

`internal/tools/get_fitness.go` (673 LOC) packs **four distinct tools** — `get_fitness`, `get_best_efforts`, `get_power_curves`, `get_training_summary` — into a single file. This is a direct violation of CLAUDE.md line 50:

> "Add new tools as `internal/tools/<tool_name>.go` with a matching `_test.go`."

Goal: split into four tool files (`get_fitness.go`, `get_best_efforts.go`, `get_power_curves.go`, `get_training_summary.go`), with a sibling `fitness_shared.go` for genuinely shared helpers (bucketing, curve smoothing). Tests follow the same split (`get_best_efforts_test.go`, etc.).

This is a mechanical refactor with **no behaviour change** — tool names, schemas, registration order, and outputs must all be byte-identical.

Audit ref: 2026-05-16 Go audit, God-module section.

PRD anchors: §7.2.C tool catalog (no name/schema change).
CLAUDE.md hard rules: one file per tool.

Complexity: Blast radius 2 (touches the tools package internally; no external consumers), Pattern novelty 1, Security 1, Reversibility 1 = 5 → Review Level 1. Size: S.

## Dependencies

- **TP-049** Step 4 — reformat long constructor lines in `get_fitness.go`. Sequence after TP-049 so reformatting and splitting don't conflict.
- **TP-048** — soft. If `tools.DecodeStrict`/`tools.TextResult` exist, each new file should use them.

## Context to Read First

- `CLAUDE.md` line 50 — the rule being enforced.
- `internal/tools/get_fitness.go` — full file.
- `internal/tools/get_fitness_test.go` — locked-in golden tests.
- `internal/tools/registry.go` — where the four tools are registered.
- One or two simpler one-file-per-tool examples for the target shape (e.g., `internal/tools/get_athlete_profile.go`).

## File Scope

- New: `internal/tools/get_best_efforts.go`, `get_power_curves.go`, `get_training_summary.go`. Move the input/output types, schema, handler, and constructor for each tool into its own file.
- Keep: `internal/tools/get_fitness.go` for the `get_fitness` tool only.
- New (if needed): `internal/tools/fitness_shared.go` for helpers used by 2+ of the split files. Resist over-extraction.
- Split tests correspondingly: `get_best_efforts_test.go`, etc.
- No change to `internal/tools/registry.go` order or call sites — registration order is observable via the catalog.
- `CHANGELOG.md`, `STATUS.md`.

Out of scope:
- Behaviour changes — schemas, defaults, outputs all byte-identical.
- Schema rewrites; even if a schema is wonky, do not touch it here.
- Renaming tools.

## Steps

### Step 1: Map the contents of `get_fitness.go`
- [ ] Catalog every top-level decl by tool ownership; record in `STATUS.md`.
- [ ] Identify genuinely shared helpers (used by 2+ tools).

### Step 2: Split
- [ ] Move per-tool types/schemas/handlers/constructors into their own files.
- [ ] Move shared helpers into `fitness_shared.go` only if used by ≥ 2 of the new files. Otherwise duplicate-and-rename or inline — three similar lines is better than premature abstraction.
- [ ] Mirror the test split.

### Step 3: Verify byte-identical behaviour
- [ ] Run the existing test suite — every test must still pass without modification (other than file moves).
- [ ] `make build` / `test` / `test-race` / `lint`.
- [ ] `bin/icuvisor` `list_advanced_capabilities` (or the equivalent catalog dump) — diff against pre-refactor must be empty.
- [ ] `scripts/snapshot_tool_schemas.go` output diff must be empty.

### Step 4: Wrap up
- [ ] Commit per tool extraction (4 commits) for review clarity, prefixed `TP-063`.
- [ ] Final commit: `TP-063 finalize get_fitness split` (the registry/shared helper cleanups, if any).

## Acceptance Criteria

- `internal/tools/get_fitness.go` contains only the `get_fitness` tool.
- Three new files exist, one per remaining tool.
- Optional `fitness_shared.go` only if used by ≥ 2 files.
- Tool catalog, schemas, and outputs are byte-identical to pre-refactor (verified via `scripts/snapshot_tool_schemas.go`).
- All `make` checks pass.

## Do NOT

- Do not change tool names, descriptions, or schemas.
- Do not change registration order in `registry.go`.
- Do not introduce a "fitness handler" abstraction layer to share code — keep helpers small and explicit.
- Do not start factoring other multi-tool files in this task.

## Documentation

- `STATUS.md`
- `CHANGELOG.md` `[Unreleased]` under "Changed" (internal refactor; no user-visible change).

## Git Commit Convention

Conventional Commits, prefixed `TP-063`. One commit per extracted tool plus a finalize commit.

---

## Amendments

_Add amendments below this line only._
