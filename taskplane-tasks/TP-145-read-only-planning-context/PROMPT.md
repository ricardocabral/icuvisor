# Task: TP-145 - Read-only planning context tool

**Created:** 2026-06-03
**Size:** M

## Review Level: 1 (Plan Only)

**Assessment:** Adds a new MCP tool by composing existing read paths, so the implementation spans registry/catalog/tests but should not introduce writes or credentials changes.
**Score:** 2/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-145-read-only-planning-context/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Read-only planning context tool. This task comes from the 2026-06-03 review of public competing intervals.icu MCP server GitHub activity. Use the public issue/PR behavior signal only; do not copy competitor implementation code or depend on GPL/copyleft source.

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

- `internal/tools/get_planning_context.go`
- `internal/tools/get_planning_context_test.go`
- `internal/tools/catalog.go`
- `internal/toolcatalog/catalog.go`
- `internal/toolcatalog/catalog_test.go`
- `README.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm the task remains clean-room: public behavior signals are okay; competitor source copying is not

### Step 1: Design the read-only contract

- [ ] Inventory existing patterns in `get_today`, `get_training_plan`, `get_events`, `get_fitness`, and planning prompts
- [ ] Define a terse default response with `_meta.source_tools`, timezone/as-of, week window, and explicit caveats
- [ ] Confirm the tool performs no create/update/delete and does not synthesize an ATP plan
- [ ] **Plan-review checkpoint**: get plan review before implementation

**Artifacts:**
- Implementation notes in STATUS.md discoveries

### Step 2: Implement `get_planning_context`

- [ ] Add the tool using existing intervals client methods and response shaping patterns
- [ ] Return week events/workouts, active training-plan summary, current/recent fitness context, upcoming race context, and caveats without calendar writes
- [ ] Register the tool in the catalog/toolcatalog with appropriate core/full tier placement
- [ ] Add input and output schema descriptions that clearly distinguish planning context from ATP creation

**Artifacts:**
- `internal/tools/get_planning_context.go` (new)
- catalog/registry files (modified)

### Step 3: Test and document

- [ ] Add table-driven handler tests for terse default, include_full behavior, source_tools metadata, timezone/week window handling, and empty-data caveats
- [ ] Add catalog/registration tests if needed
- [ ] Update CHANGELOG and README/catalog docs if user-visible
- [ ] Run targeted tests: `go test ./internal/tools ./internal/toolcatalog`

**Artifacts:**
- `internal/tools/get_planning_context_test.go` (new)
- `README.md` / `CHANGELOG.md` (modified if needed)

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
- `CHANGELOG.md` — record the new tool under Unreleased
- `README.md` — update public tool/capability summary if it lists tools

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

- **Step completion:** `feat(TP-145): complete Step N — description`
- **Bug fixes:** `fix(TP-145): description`
- **Tests:** `test(TP-145): description`
- **Docs:** `docs(TP-145): description`
- **Hydration:** `hydrate: TP-145 expand Step N checkboxes`

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
