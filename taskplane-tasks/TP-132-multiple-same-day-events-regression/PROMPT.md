# Task: TP-132 - Multiple same-day events regression pack

**Created:** 2026-05-29
**Size:** S

## Review Level: 1 (Plan Only)

**Assessment:** Adds targeted tests around existing event/today read behavior. Low novelty, but regressions would directly affect daily planning reliability.
**Score:** 3/8 — Blast radius: 1, Pattern novelty: 0, Security: 0, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-132-multiple-same-day-events-regression/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Ensure `get_today` and `get_events` reliably surface multiple workouts/events on the same athlete-local date, preserving enough order and identity for assistants to avoid telling users the wrong thing about today or tomorrow.

## Evidence from forum review

- IntervalCoach multiple workouts on same day did not surface: https://forum.intervals.icu/t/120045/820

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `docs/prd/PRD-icuvisor.md` — `get_today`, `get_events`, and athlete-local date behavior.
- `CONTRIBUTING.md` — test and Go workflow expectations.

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `internal/tools/get_today_test.go`
- `internal/tools/get_today.go`
- `internal/tools/get_events_training_plan_test.go`
- `internal/tools/get_events.go`
- `scripts/eval/scenarios/cookbook_scenarios.json`
- `CHANGELOG.md`

## Steps

> **Hydration:** STATUS.md tracks outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

### Step 1: Audit same-day event handling
- [ ] Inspect `get_today` and `get_events` shaping/tests for multiple same-day planned workouts, notes, and races.
- [ ] Confirm athlete-local date filtering does not collapse rows by date.
- [ ] Record missing cases in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/tools`

**Artifacts:**
- `internal/tools/get_today.go`
- `internal/tools/get_events.go`
- `internal/tools/get_today_test.go`
- `internal/tools/get_events_training_plan_test.go`

### Step 2: Add regression coverage
- [ ] Add tests with at least two WORKOUT events on the same date plus optional NOTE/race annotations.
- [ ] Assert both entries are present, separately identifiable, and not overwritten by map/date grouping.
- [ ] If useful, add an eval scenario for “what is on tomorrow?” with two planned sessions.
- [ ] Run targeted tests: `go test ./internal/tools` and `make eval-validate` if eval changed.

**Artifacts:**
- `internal/tools/get_today_test.go`
- `internal/tools/get_events_training_plan_test.go`
- `scripts/eval/scenarios/cookbook_scenarios.json`

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
- `CHANGELOG.md` — note regression coverage or behavior fix.

**Check If Affected:**
- `docs/prd/PRD-icuvisor.md` — update only if response behavior changes materially.

## Completion Criteria

- [ ] All steps complete
- [ ] All tests passing
- [ ] Documentation updated

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-132): complete Step N — description`
- **Bug fixes:** `fix(TP-132): description`
- **Tests:** `test(TP-132): description`
- **Hydration:** `hydrate: TP-132 expand Step N checkboxes`

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
