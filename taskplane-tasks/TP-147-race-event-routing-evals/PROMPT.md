# Task: TP-147 - Race-event routing evals for add_or_update_event

**Created:** 2026-06-03
**Size:** S

## Review Level: 0 (None)

**Assessment:** Evaluation/test-only task using existing routing fixture patterns; no runtime behavior should change unless tests reveal a schema/description issue.
**Score:** 1/8 — Blast radius: 1, Pattern novelty: 0, Security: 0, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-147-race-event-routing-evals/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Race-event routing evals for add_or_update_event. This task comes from the 2026-06-03 review of public competing intervals.icu MCP server GitHub activity. Use the public issue/PR behavior signal only; do not copy competitor implementation code or depend on GPL/copyleft source.

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

- `internal/toolrouting/*`
- `internal/prompts/*`
- `internal/tools/add_or_update_event.go`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm the task remains clean-room: public behavior signals are okay; competitor source copying is not

### Step 1: Add race-event routing cases

- [ ] Find the existing tool-routing fixture/test pattern
- [ ] Add prompts for creating A/B/C races and assert the expected first tool is `add_or_update_event`
- [ ] Include a negative assertion or fixture note that a separate `add_race_event` tool should not be required
- [ ] If routing fails, first improve examples/descriptions minimally; do not add a dedicated tool in this task
- [ ] Run targeted tests: `go test ./internal/toolrouting ./internal/prompts -run 'Race|Routing|Fixture'`

**Artifacts:**
- Tool-routing fixtures/tests (modified)
- `internal/tools/add_or_update_event.go` (modified only if descriptions/examples need minimal routing help)

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

- **Step completion:** `feat(TP-147): complete Step N — description`
- **Bug fixes:** `fix(TP-147): description`
- **Tests:** `test(TP-147): description`
- **Docs:** `docs(TP-147): description`
- **Hydration:** `hydrate: TP-147 expand Step N checkboxes`

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
