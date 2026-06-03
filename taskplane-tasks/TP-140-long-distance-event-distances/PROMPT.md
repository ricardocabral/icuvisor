# Task: TP-140 - Long-distance event distance regression coverage

**Created:** 2026-06-03
**Size:** S

## Review Level: 1 (Plan Only)

**Assessment:** Adds focused event read/write tests and removes any arbitrary local distance cap if found. Low novelty, but user-visible race planning can be wrong if long events are rejected or truncated.
**Score:** 3/8 — Blast radius: 1, Pattern novelty: 0, Security: 0, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-140-long-distance-event-distances/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Ensure Icuvisor accepts and preserves long endurance event distances such as 1200 km randonneuring events. Forum users noted that 1000 km caps are too low; Icuvisor should avoid arbitrary local caps and should clearly report any true upstream refusal instead of silently truncating or implying automatic load calculation.

## Evidence from forum review

- Public intervals.icu forum comments from 2026-05-30 through 2026-06-03 reported this behavior in adjacent AI coach/MCP products. Use these as behavior signals only; do not open, copy, or infer from competitor source code.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `docs/prd/PRD-icuvisor.md` — relevant product/tool contract and safety constraints.
- `CONTRIBUTING.md` — test and Go workflow expectations.

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `internal/tools/add_or_update_event.go`
- `internal/tools/add_or_update_event_test.go`
- `internal/tools/get_events.go`
- `internal/tools/get_events_training_plan_test.go`
- `internal/intervals/*`
- `CHANGELOG.md`

## Steps

> **Hydration:** STATUS.md tracks outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

### Step 0: Preflight
- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

### Step 1: Audit event distance handling
- [ ] Inspect event write/read validation for distance limits, units, and load/target-load wording.
- [ ] Confirm whether any Icuvisor-local cap below 1200 km exists; record upstream-only constraints in STATUS.md.
- [ ] Check response wording does not imply Icuvisor auto-calculates load from distance/duration unless it actually does.
- [ ] Run targeted tests: `go test ./internal/tools`.

**Artifacts:**
- `internal/tools/add_or_update_event.go`
- `internal/tools/add_or_update_event_test.go`
- `internal/tools/get_events.go`
- `internal/tools/get_events_training_plan_test.go`
- `internal/intervals/*`
- `CHANGELOG.md`

### Step 2: Add long-distance regression tests
- [ ] Add tests for creating/updating and reading a 1200 km event or workout/race distance as meters.
- [ ] Remove or relax any arbitrary Icuvisor-local cap below randonneuring distances, preserving upstream error passthrough as actionable user errors.
- [ ] Assert no truncation and no false auto-load claim in response metadata/rows.
- [ ] Run targeted tests: `go test ./internal/tools`.

**Artifacts:**
- `internal/tools/add_or_update_event.go`
- `internal/tools/add_or_update_event_test.go`
- `internal/tools/get_events.go`
- `internal/tools/get_events_training_plan_test.go`
- `internal/intervals/*`
- `CHANGELOG.md`

### Step 3: Testing & Verification

> ZERO test failures allowed. This step runs the FULL test suite as a quality gate. Earlier steps should use targeted tests for fast feedback.

- [ ] Run FULL test suite: `make test`
- [ ] Run lint: `make lint`
- [ ] Fix all failures or document pre-existing unrelated failures with exact command output
- [ ] Build passes: `make build`

### Step 4: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**
- `CHANGELOG.md` — note long-distance event regression coverage or validation fix.

**Check If Affected:**
- `docs/prd/PRD-icuvisor.md` — update only if event distance contract changes materially.

## Completion Criteria

- [ ] All steps complete
- [ ] All tests passing
- [ ] Documentation updated

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-140): complete Step N — description`
- **Bug fixes:** `fix(TP-140): description`
- **Tests:** `test(TP-140): description`
- **Hydration:** `hydrate: TP-140 expand Step N checkboxes`

## Do NOT

- Expand task scope — add tech debt to CONTEXT.md instead
- Skip tests
- Modify framework/standards docs without explicit user approval
- Load docs not listed in "Context to Read First"
- Open, copy, paraphrase, or transliterate GPL/copyleft competitor source
- Add arbitrary event distance caps not required by upstream
- Commit without the task ID prefix in the commit message

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if issues discovered during execution.
     Format:
     ### Amendment N — YYYY-MM-DD HH:MM
     **Issue:** [what was wrong]
     **Resolution:** [what was changed] -->
