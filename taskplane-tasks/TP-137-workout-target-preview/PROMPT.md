# Task: TP-137 - Resolved workout target previews for planned workouts

**Created:** 2026-06-03
**Size:** M

## Review Level: 2 (Plan and Code)

**Assessment:** Adds derived presentation fields across workout/event read shapes using existing profile thresholds. Moderate novelty because it must avoid bloating terse responses or misrepresenting targets without enough profile data.
**Score:** 4/8 — Blast radius: 2, Pattern novelty: 1, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-137-workout-target-preview/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Provide human-friendly resolved targets for structured planned workouts where possible, such as `% FTP` targets converted to watts and threshold-pace percentages converted to the athlete's preferred pace units. Forum users asked to see actual watts alongside percentage FTP in workout details; Icuvisor should add compact, clearly labeled previews without exposing raw heavy workout docs by default.

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

- `internal/tools/get_events.go`
- `internal/tools/get_events_training_plan_test.go`
- `internal/tools/get_workout_library.go`
- `internal/tools/get_workout_library_test.go`
- `internal/tools/typed_json_bodies.go`
- `internal/tools/get_athlete_profile.go`
- `internal/workoutdoc/*`
- `internal/units/*`
- `CHANGELOG.md`

## Steps

> **Hydration:** STATUS.md tracks outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

### Step 0: Preflight
- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

### Step 1: Design compact resolved-target shape
- [ ] Audit event/workout read rows and `workout_doc_summary` to find the least-bloated place for target previews.
- [ ] Use athlete profile thresholds/units only when already available or cheaply fetchable; avoid extra heavy calls or raw payload expansion.
- [ ] Record unsupported target cases and null/omission rules in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/workoutdoc`.

**Artifacts:**
- ``internal/tools/get_events.go`
- ``internal/tools/get_events_training_plan_test.go`
- ``internal/tools/get_workout_library.go`
- ``internal/tools/get_workout_library_test.go`
- ``internal/tools/typed_json_bodies.go`
- ``internal/tools/get_athlete_profile.go`

### Step 2: Implement target previews and tests
- [ ] Add tests for `% FTP` planned workout targets resolving to watts from profile FTP.
- [ ] Add tests or explicit omissions for HR threshold, pace threshold, missing profile threshold, and non-numeric/text targets.
- [ ] Implement compact preview fields while preserving terse-by-default and `include_full` behavior.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/workoutdoc`.

**Artifacts:**
- ``internal/tools/get_events.go`
- ``internal/tools/get_events_training_plan_test.go`
- ``internal/tools/get_workout_library.go`
- ``internal/tools/get_workout_library_test.go`
- ``internal/tools/typed_json_bodies.go`
- ``internal/tools/get_athlete_profile.go`

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
- `CHANGELOG.md` — note resolved workout target preview behavior.

**Check If Affected:**
- `docs/prd/PRD-icuvisor.md` — update if the read response shape gains a new documented field.
- `docs/internal/forum-announcement.md` — update only if this becomes a launch talking point.

## Completion Criteria

- [ ] All steps complete
- [ ] All tests passing
- [ ] Documentation updated

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-137): complete Step N — description`
- **Bug fixes:** `fix(TP-137): description`
- **Tests:** `test(TP-137): description`
- **Hydration:** `hydrate: TP-137 expand Step N checkboxes`

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
