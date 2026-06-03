# Task: TP-148 - Public positioning for gear resolution and unit-safe output

**Created:** 2026-06-03
**Size:** S

## Review Level: 0 (None)

**Assessment:** Docs-only improvement with easy rollback and no production behavior.
**Score:** 0/8 — Blast radius: 0, Pattern novelty: 0, Security: 0, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-148-gear-units-positioning-docs/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Public positioning for gear resolution and unit-safe output. This task comes from the 2026-06-03 review of public competing intervals.icu MCP server GitHub activity. Use the public issue/PR behavior signal only; do not copy competitor implementation code or depend on GPL/copyleft source.

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

- `README.md`
- `docs/**/*.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm the task remains clean-room: public behavior signals are okay; competitor source copying is not

### Step 1: Improve public positioning copy

- [ ] Review README and docs surfaces that explain capabilities
- [ ] Add concise, accurate copy for gear-name resolution: bike/shoe names when upstream gear IDs can be resolved, with explicit unresolved status
- [ ] Add concise, accurate copy for unit-safe output: unit-labeled fields, calories burned vs consumed, scale legends
- [ ] Avoid unsupported claims about coaching quality, hosted features, or automatic calendar planning

**Artifacts:**
- `README.md` and/or relevant docs (modified)

### Step 2: Verify docs

- [ ] Inspect rendered Markdown structure or run available markdown/link checks
- [ ] Confirm claims match implemented tools and tests
- [ ] Run `make test` if any code changed; otherwise note docs-only verification in STATUS.md

**Artifacts:**
- STATUS.md verification notes

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
- `README.md` — add/clarify public positioning copy

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

- **Step completion:** `feat(TP-148): complete Step N — description`
- **Bug fixes:** `fix(TP-148): description`
- **Tests:** `test(TP-148): description`
- **Docs:** `docs(TP-148): description`
- **Hydration:** `hydrate: TP-148 expand Step N checkboxes`

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
