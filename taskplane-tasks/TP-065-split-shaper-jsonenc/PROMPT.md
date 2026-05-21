# TP-065 — Split `internal/response/shaper.go` into focused files (audit God-module)

## Mission

`internal/response/shaper.go` is 771 LOC and tangles three responsibilities:

- (a) Hand-rolled reflection-based JSON marshaller (lines ~86-365) that partly duplicates `encoding/json`.
- (b) JSON tree walker (`walkJSON`, lines ~387-425) — the target of TP-047's consolidation.
- (c) MCP response shaping / meta enrichment (lines ~427-611).

Even after TP-043 (remove global state) and TP-047 (consolidate walkers) land, the file remains a god module. Goal: finish the job by splitting into focused files (or a small subpackage where it pays for itself):

- `internal/response/marshal.go` — `toJSONValue` / reflection-based marshaller. Consider extracting to subpackage `internal/response/jsonenc/` **only if** the marshaller survives TP-047 (TP-047 may drop the marshal round-trip entirely; in that case this step is a no-op).
- `internal/response/walk.go` — single tree-walker primitive.
- `internal/response/shape.go` — public shaping API surface.
- `internal/response/meta.go` — catalog/version/scale/unit `_meta` enrichment.

This is the final cleanup pass on `shaper.go`. No behaviour change.

Audit ref: 2026-05-16 Go audit, God-module section.

PRD anchors: §7.2.D response shaping invariants.
CLAUDE.md hard rules: small packages, small files, one obvious concern per file.

Complexity: Blast radius 3 (touches the response package's internal layout), Pattern novelty 2 (subpackage decision), Security 1, Reversibility 2 = 8 → Review Level 2. Size: M.

## Dependencies

- **TP-043** — must land first. TP-043 removes global mutable state; splitting before that lands risks duplicating the state-removal work.
- **TP-047** — must land first. TP-047 consolidates walkers and may drop the marshal round-trip. The set of files needed here depends on TP-047's outcome.
- **TP-051** (tool catalog generator) — soft. If the catalog generator references `response.SetCatalogMeta` (or similar), keep the public API stable across the split.

## Context to Read First

- TP-043 and TP-047 PROMPT.md + STATUS.md — understand what they leave behind.
- `internal/response/shaper.go` — full file (post-TP-043, post-TP-047 if both have landed).
- `internal/response/shaper_test.go` — golden tests that must keep passing.
- Any external caller of `internal/response.*` — verify the public surface remains.

## File Scope

- Split `internal/response/shaper.go` into the files listed above, inside the same package.
- Decide on `internal/response/jsonenc/` subpackage **only if** the marshaller survived TP-047 and exceeds ~200 LOC. Otherwise keep `marshal.go` in `internal/response/`.
- No public API change. Same exported names, same signatures.
- Test files split alongside the source split where it improves clarity.
- `CHANGELOG.md`, `STATUS.md`.

Out of scope:
- New marshalling strategies (covered by TP-047).
- Removing further globals (covered by TP-043).
- Changing `_meta` contents or shape.
- Changing pagination / unit / scale outputs.

## Steps

### Step 1: Re-baseline after TP-043 + TP-047
- [ ] Confirm both have landed; if not, block this task.
- [ ] Inventory every top-level decl in the post-merge `shaper.go` and assign each to a target file. Record in `STATUS.md`.

### Step 2: Decide on subpackage
- [ ] If `toJSONValue` survives TP-047 and > 200 LOC: create `internal/response/jsonenc/`. Otherwise: keep in `internal/response/marshal.go`.
- [ ] Document the decision in `STATUS.md` with the line-count evidence.

### Step 3: Mechanical split
- [ ] Move decls into the new files without logic changes.
- [ ] Run the full test suite after each file move.

### Step 4: Verify
- [ ] `make build` / `test` / `test-race` / `lint`.
- [ ] `scripts/snapshot_tool_schemas.go` diff empty.
- [ ] `wc -l internal/response/*.go` — each focused file ≤ ~300 LOC.

## Acceptance Criteria

- `internal/response/shaper.go` is either deleted or contains only the package doc + the public shaping entry point.
- Marshal / walk / shape / meta concerns each have their own file (or `jsonenc/` subpackage).
- Public API unchanged.
- All `make` checks pass.
- No new global state, no `init()` functions.

## Do NOT

- Do not start before TP-043 and TP-047 have merged.
- Do not change `_meta` semantics or wire shape.
- Do not introduce a third-party JSON library.
- Do not expose new public types/functions.

## Documentation

- `STATUS.md`
- `CHANGELOG.md` `[Unreleased]` under "Changed" (internal refactor).

## Git Commit Convention

Conventional Commits, prefixed `TP-065`. One commit per file extraction is fine; reviewers will want diffs they can read.

---

## Amendments

_Add amendments below this line only._
