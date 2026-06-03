# Task: TP-134 - Calendar write idempotency and duplicate prevention

**Created:** 2026-06-03
**Size:** M

## Review Level: 2 (Plan and Code)

**Assessment:** Hardens existing calendar write paths and adds regression coverage for retry/concurrency behavior. Blast radius is one tool cluster, but calendar writes can create user-visible duplicate workouts.
**Score:** 4/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-134-calendar-write-idempotency/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Ensure Icuvisor calendar writes are safe under retries and near-concurrent invocations. Users of competing AI coaches reported duplicate workouts and write loops when two planning processes wrote at the same time; Icuvisor should make repeated `apply_training_plan` and same-day event writes as idempotent and observable as the upstream API allows.

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

- `internal/tools/apply_training_plan.go`
- `internal/tools/apply_training_plan_test.go`
- `internal/tools/add_or_update_event.go`
- `internal/tools/add_or_update_event_test.go`
- `internal/intervals/*`
- `CHANGELOG.md`

## Steps

> **Hydration:** STATUS.md tracks outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

### Step 0: Preflight
- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

### Step 1: Audit write retry and duplicate behavior
- [ ] Inspect `apply_training_plan` and `add_or_update_event` for retry, repeated-call, and concurrent-call behavior.
- [ ] Identify whether duplicate detection can be done deterministically from existing event fields before writes.
- [ ] Record the chosen idempotency contract and any upstream limitations in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/tools`.

**Artifacts:**
- ``internal/tools/apply_training_plan.go`
- ``internal/tools/apply_training_plan_test.go`
- ``internal/tools/add_or_update_event.go`
- ``internal/tools/add_or_update_event_test.go`
- ``internal/intervals/*`
- ``CHANGELOG.md`

### Step 2: Implement duplicate prevention or explicit duplicate warnings
- [ ] Add tests for repeated `apply_training_plan` calls against the same plan/date range and for duplicate same-day planned events.
- [ ] Implement deduplication, stable skip behavior, idempotency keys/metadata, or explicit duplicate warnings using existing upstream fields only.
- [ ] Ensure dry-run output makes potential duplicate/conflict outcomes clear before any write.
- [ ] Run targeted tests: `go test ./internal/tools`.

**Artifacts:**
- ``internal/tools/apply_training_plan.go`
- ``internal/tools/apply_training_plan_test.go`
- ``internal/tools/add_or_update_event.go`
- ``internal/tools/add_or_update_event_test.go`
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
- `CHANGELOG.md` — note duplicate-prevention/idempotency behavior or regression coverage.

**Check If Affected:**
- `docs/prd/PRD-icuvisor.md` — update only if the calendar write contract changes materially.
- `docs/dogfood/v0.3-prompts.md` — update only if write smoke-test prompts should exercise repeated writes.

## Completion Criteria

- [ ] All steps complete
- [ ] All tests passing
- [ ] Documentation updated

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-134): complete Step N — description`
- **Bug fixes:** `fix(TP-134): description`
- **Tests:** `test(TP-134): description`
- **Hydration:** `hydrate: TP-134 expand Step N checkboxes`

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
