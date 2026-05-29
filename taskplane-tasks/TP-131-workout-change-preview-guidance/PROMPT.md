# Task: TP-131 - Workout change preview guidance

**Created:** 2026-05-29
**Size:** M

## Review Level: 2 (Plan and Code)

**Assessment:** Touches write-flow prompts and potentially write-tool examples. It reduces risk before writes but must not add model-controlled confirmations that bypass safety gates.
**Score:** 4/8 — Blast radius: 1, Pattern novelty: 1, Security: 1, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-131-workout-change-preview-guidance/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Make assistants present understandable proposed workout changes before asking users to accept or writing to the calendar. Public feedback shows graph-only or vague proposed changes are hard to approve; icuvisor should guide assistants to summarize duration, intervals, intensity targets, load, and deltas using `validate_workout` where useful.

## Evidence from forum review

- IntervalCoach proposed change was hard to understand: https://forum.intervals.icu/t/120045/818
- IcuSync edit-in-place safety conversation also emphasized explicit approval before alternate processes: https://forum.intervals.icu/t/126632/233

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `docs/prd/PRD-icuvisor.md` — workout write semantics and user approval boundaries.
- `internal/resources/testdata/workout_syntax.md` — workout DSL behavior.
- `CONTRIBUTING.md` — test and Go workflow expectations.

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `internal/tools/validate_workout.go`
- `internal/tools/validate_workout_test.go`
- `internal/tools/add_or_update_event.go`
- `internal/tools/create_workout.go`
- `internal/tools/update_workout.go`
- `internal/prompts/testdata/weekly_planning.md`
- `web/content/cookbook/build-workouts.md`
- `CHANGELOG.md`

## Steps

> **Hydration:** STATUS.md tracks outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

### Step 1: Audit current pre-write guidance
- [ ] Inspect `validate_workout`, workout write tool descriptions/examples, and weekly-planning/build-workouts prompts.
- [ ] Identify whether assistants are instructed to summarize proposed changes before writes.
- [ ] Record current behavior and chosen changes in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/prompts`

**Artifacts:**
- `internal/tools/validate_workout.go`
- `internal/tools/add_or_update_event.go`
- `internal/prompts/testdata/weekly_planning.md`

### Step 2: Harden preview guidance
- [ ] Update prompts/tool examples so proposed changes include total duration, key steps, target intensities, load/distance/time changes, and what is being preserved.
- [ ] Recommend `validate_workout` preflight for uncertain DSL or structured workout changes.
- [ ] Do not introduce a model-controlled `confirm` override or bypass safety modes.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/prompts`

**Artifacts:**
- `internal/tools/validate_workout.go`
- `internal/tools/add_or_update_event.go`
- `internal/tools/create_workout.go`
- `internal/tools/update_workout.go`
- `internal/prompts/testdata/weekly_planning.md`

### Step 3: Update cookbook examples
- [ ] Update build-workouts cookbook with before/after preview language and approval workflow.
- [ ] Ensure examples distinguish prose description from structured `workout_doc`.
- [ ] Run targeted tests/docs validation as available.

**Artifacts:**
- `web/content/cookbook/build-workouts.md`

### Step 4: Testing & Verification

> ZERO test failures allowed. This step runs the FULL test suite as a quality gate. Earlier steps should use targeted tests for fast feedback.

- [ ] Run FULL test suite: `make test`
- [ ] Run lint: `make lint`
- [ ] Fix all failures or document pre-existing unrelated failures with exact command output
- [ ] Build passes: `make build`

### Step 5: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**
- `CHANGELOG.md` — note prompt/tool guidance changes.

**Check If Affected:**
- `docs/prd/PRD-icuvisor.md` — update only if write behavior changes.
- `web/content/reference/tools.md` — update via generated tooling only if tool descriptions change.

## Completion Criteria

- [ ] All steps complete
- [ ] All tests passing
- [ ] Documentation updated

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-131): complete Step N — description`
- **Bug fixes:** `fix(TP-131): description`
- **Tests:** `test(TP-131): description`
- **Hydration:** `hydrate: TP-131 expand Step N checkboxes`

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
