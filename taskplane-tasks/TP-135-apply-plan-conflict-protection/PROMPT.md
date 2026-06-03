# Task: TP-135 - Apply training plan conflict protection for non-workout calendar items

**Created:** 2026-06-03
**Size:** M

## Review Level: 2 (Plan and Code)

**Assessment:** Refines destructive conflict handling in an existing write tool. Risk is moderate because an incorrect implementation can delete or overwrite races, notes, or unavailable blocks.
**Score:** 4/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-135-apply-plan-conflict-protection/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Protect non-workout calendar items during plan application. Competing tools showed that `UNAVAILABLE` days and other calendar annotations can accidentally collapse or wipe planned weeks; Icuvisor should make conflicts explicit and avoid replacing races, notes, and unavailable blocks unless the user has intentionally scoped that behavior.

## Evidence from forum review

- Public intervals.icu forum comments from 2026-05-30 through 2026-06-03 reported this behavior in adjacent AI coach/MCP products. Use these as behavior signals only; do not open, copy, or infer from competitor source code.

## Dependencies

- **Task:** TP-134 (calendar write duplicate/idempotency behavior should land first so this task can build on the final conflict contract)

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

- `internal/tools/apply_training_plan.go`
- `internal/tools/apply_training_plan_test.go`
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

### Step 1: Audit conflict shape and replace policy
- [ ] Inspect `fetchApplyTrainingPlanConflicts` and `replace_existing` deletion behavior.
- [ ] Confirm how event category/type/name/date are available from existing event rows and upstream raw fields.
- [ ] Define a safe conflict taxonomy in STATUS.md: workout conflicts vs protected annotations/races/unavailable items.
- [ ] Run targeted tests: `go test ./internal/tools`.

**Artifacts:**
- ``internal/tools/apply_training_plan.go`
- ``internal/tools/apply_training_plan_test.go`
- ``internal/tools/get_events.go`
- ``internal/tools/get_events_training_plan_test.go`
- ``internal/intervals/*`
- ``CHANGELOG.md`

### Step 2: Add protected-conflict behavior and tests
- [ ] Extend conflict output to include enough category/type/name information for LLMs to explain why a day was skipped.
- [ ] Ensure `replace_existing` deletes only intended workout conflicts; protected NOTE, RACE, and UNAVAILABLE-like events are skipped/reported unless a clearly named server-side policy is added.
- [ ] Add tests for mixed calendar days containing a workout plus NOTE/race/unavailable event.
- [ ] Run targeted tests: `go test ./internal/tools`.

**Artifacts:**
- ``internal/tools/apply_training_plan.go`
- ``internal/tools/apply_training_plan_test.go`
- ``internal/tools/get_events.go`
- ``internal/tools/get_events_training_plan_test.go`
- ``internal/intervals/*`
- ``CHANGELOG.md`

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
- `CHANGELOG.md` — note safer plan-application conflict behavior.

**Check If Affected:**
- `docs/prd/PRD-icuvisor.md` — update only if `apply_training_plan` user-visible contract changes.
- `docs/dogfood/v0.3-prompts.md` — update only if dogfood should verify protected conflicts.

## Completion Criteria

- [ ] All steps complete
- [ ] All tests passing
- [ ] Documentation updated

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-135): complete Step N — description`
- **Bug fixes:** `fix(TP-135): description`
- **Tests:** `test(TP-135): description`
- **Hydration:** `hydrate: TP-135 expand Step N checkboxes`

## Do NOT

- Expand task scope — add tech debt to CONTEXT.md instead
- Skip tests
- Modify framework/standards docs without explicit user approval
- Load docs not listed in "Context to Read First"
- Open, copy, paraphrase, or transliterate GPL/copyleft competitor source
- Add first-party Strava/TrainingPeaks ingestion or hosted SaaS behavior
- Commit without the task ID prefix in the commit message

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if issues discovered during execution.
     Format:
     ### Amendment N — YYYY-MM-DD HH:MM
     **Issue:** [what was wrong]
     **Resolution:** [what was changed] -->
