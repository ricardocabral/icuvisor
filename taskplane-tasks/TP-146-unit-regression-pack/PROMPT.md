# Task: TP-146 - Unit regression pack for work, calories, and hydration

**Created:** 2026-06-03
**Size:** S

## Review Level: 0 (None)

**Assessment:** Primarily adds regression tests around existing unit-safe behavior; any production fix should stay minimal and test-driven.
**Score:** 1/8 — Blast radius: 1, Pattern novelty: 0, Security: 0, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-146-unit-regression-pack/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Unit regression pack for work, calories, and hydration. This task comes from the 2026-06-03 review of public competing intervals.icu MCP server GitHub activity. Use the public issue/PR behavior signal only; do not copy competitor implementation code or depend on GPL/copyleft source.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — repository rules, clean-room constraints, test expectations
- `docs/prd/PRD-icuvisor.md` — product/tool-catalog behavior if changing tool output or docs
- `ROADMAP.md` — phasing if this changes roadmap-visible scope

## Environment

- **Workspace:** `/Users/jusbrasil/prj/icuvisor`
- **Services required:** None

## File Scope

- `internal/tools/*_test.go`
- `internal/response/*_test.go`
- `internal/analysis/*_test.go`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm the task remains clean-room: public behavior signals are okay; competitor source copying is not

### Step 1: Audit current unit coverage

- [ ] Locate existing tests for extended metrics Joules/kJ, wellness kcal/hydration, activity calories semantics, and unit metadata
- [ ] Identify missing regression assertions without duplicating existing coverage
- [ ] Record any already-covered items in STATUS.md discoveries

**Artifacts:**
- STATUS.md discoveries (modified)

### Step 2: Add unit regression tests

- [ ] Add or tighten tests for raw Joules emitted only as explicit kJ-derived fields where applicable
- [ ] Add or tighten tests for wellness `kcalConsumed` and `hydrationVolume` unit semantics
- [ ] Assert zero values are preserved and ambiguous raw field names are not emitted in terse responses
- [ ] Run targeted tests: `go test ./internal/tools ./internal/response ./internal/analysis -run 'Joule|Calories|Hydration|Unit'`

**Artifacts:**
- Relevant `_test.go` files (modified)

### Step 99: Testing & Verification

- [ ] Run targeted tests from implementation steps
- [ ] Run FULL test suite: `make test`
- [ ] Run build if code changed: `make build`
- [ ] Fix all failures

### Step 100: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**
- None

**Check If Affected:**
- `CHANGELOG.md` — update if behavior or user-visible docs change
- `README.md` — update if public capabilities or examples changed
- `docs/prd/PRD-icuvisor.md` — update only if product scope changes; otherwise leave unchanged
- `ROADMAP.md` — update only if phasing changes; otherwise leave unchanged

## Completion Criteria

- [ ] All steps complete
- [ ] All tests passing
- [ ] Documentation updated or explicitly deemed unaffected

## Git Commit Convention

Commits happen at step boundaries. All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-146): complete Step N — description`
- **Bug fixes:** `fix(TP-146): description`
- **Tests:** `test(TP-146): description`
- **Docs:** `docs(TP-146): description`
- **Hydration:** `hydrate: TP-146 expand Step N checkboxes`

## Do NOT

- Expand task scope — add tech debt to CONTEXT.md instead
- Skip tests
- Modify framework/standards docs without explicit user approval
- Load docs not listed in "Context to Read First"
- Commit without the task ID prefix in the commit message
- Paste, paraphrase, transliterate, or port competitor source code

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if issues discovered during execution. -->
