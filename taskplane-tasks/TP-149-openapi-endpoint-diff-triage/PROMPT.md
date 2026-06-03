# Task: TP-149 - OpenAPI endpoint-diff triage automation

**Created:** 2026-06-03
**Size:** M

## Review Level: 1 (Plan Only)

**Assessment:** Adds maintenance automation around upstream API discovery, touching scripts/workflows and docs but not runtime server behavior.
**Score:** 2/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-149-openapi-endpoint-diff-triage/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

OpenAPI endpoint-diff triage automation. This task comes from the 2026-06-03 review of public competing intervals.icu MCP server GitHub activity. Use the public issue/PR behavior signal only; do not copy competitor implementation code or depend on GPL/copyleft source.

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

- `.github/workflows/*`
- `scripts/**/*openapi*`
- `docs/**/*.md`
- `taskplane-tasks/CONTEXT.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm the task remains clean-room: public behavior signals are okay; competitor source copying is not

### Step 1: Design endpoint-diff triage workflow

- [ ] Inspect existing scripts/workflows and decide whether to add a standalone script, scheduled workflow, or documented manual command
- [ ] Ensure normal tests do not hit the network; any live fetch must be opt-in or confined to CI schedule/manual workflow
- [ ] Define output that highlights added/removed OpenAPI paths and creates a human-triage artifact without auto-implementing endpoints
- [ ] **Plan-review checkpoint**: get plan review before implementation

**Artifacts:**
- STATUS.md design notes

### Step 2: Implement OpenAPI diff tooling

- [ ] Add script or workflow that compares a pinned/baseline intervals.icu OpenAPI spec against latest fetched spec
- [ ] Add fixture-based tests for added path detection, removed path detection, and no-change output
- [ ] Document how maintainers triage new endpoints into Taskplane/backlog tasks
- [ ] Run targeted tests for the script/tooling

**Artifacts:**
- `.github/workflows/*` or `scripts/**/*openapi*` (new/modified)
- docs for triage process (modified if needed)

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
- A maintainer-facing doc or README section explaining how to run and triage the OpenAPI endpoint diff

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

- **Step completion:** `feat(TP-149): complete Step N — description`
- **Bug fixes:** `fix(TP-149): description`
- **Tests:** `test(TP-149): description`
- **Docs:** `docs(TP-149): description`
- **Hydration:** `hydrate: TP-149 expand Step N checkboxes`

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
