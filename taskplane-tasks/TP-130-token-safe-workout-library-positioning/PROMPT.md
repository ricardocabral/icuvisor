# Task: TP-130 - Token-safe workout library positioning

**Created:** 2026-05-29
**Size:** S

## Review Level: 1 (Plan Only)

**Assessment:** Primarily documentation and eval guidance around existing workout-library tools. Low security risk, but it affects user expectations and token-efficiency positioning.
**Score:** 3/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-130-token-safe-workout-library-positioning/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Highlight and protect icuvisor’s token-safe workout-library access as an advantage over approaches that try to dump entire workout libraries into the model. Competitor feedback calls full-library ingestion a mammoth data challenge; icuvisor should document folder/pagination workflows and avoid context blowouts.

## Evidence from forum review

- Montis workout-library scale note: https://forum.intervals.icu/t/117856/475

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `docs/prd/PRD-icuvisor.md` — KR5 token efficiency and workout-library tool catalog.
- `docs/kr5-benchmark.md` — token-efficiency positioning.
- `CONTRIBUTING.md` — test and Go workflow expectations.

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `internal/tools/get_workout_library.go`
- `internal/tools/get_workouts_in_folder.go`
- `internal/tools/get_workout_library_test.go`
- `web/content/cookbook/build-workouts.md`
- `web/content/explain/terse-by-default.md`
- `web/content/reference/tools.md`
- `CHANGELOG.md`

## Steps

> **Hydration:** STATUS.md tracks outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

### Step 1: Audit workout-library response shape
- [ ] Inspect workout-library tools/tests for pagination, terse default, and folder scoping.
- [ ] Record whether existing tests protect against huge raw payloads and `include_full` behavior.
- [ ] Run targeted tests: `go test ./internal/tools`

**Artifacts:**
- `internal/tools/get_workout_library.go`
- `internal/tools/get_workouts_in_folder.go`
- `internal/tools/get_workout_library_test.go`

### Step 2: Add docs/eval hardening
- [ ] Update build-workouts cookbook to recommend folder-filtered and paginated library queries instead of dumping all templates.
- [ ] Add or update tests only if audit finds missing token-safety coverage.
- [ ] Mention the local/token-safe advantage without naming or disparaging competitors.
- [ ] Run targeted tests: `go test ./internal/tools`

**Artifacts:**
- `web/content/cookbook/build-workouts.md`
- `web/content/explain/terse-by-default.md`
- `internal/tools/get_workout_library_test.go`

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
- `CHANGELOG.md` — note docs/test hardening if changed.

**Check If Affected:**
- `docs/kr5-benchmark.md` — update only if measured token claims change.
- `web/content/reference/tools.md` — update via generated tooling only if tool docs change.

## Completion Criteria

- [ ] All steps complete
- [ ] All tests passing
- [ ] Documentation updated

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-130): complete Step N — description`
- **Bug fixes:** `fix(TP-130): description`
- **Tests:** `test(TP-130): description`
- **Hydration:** `hydrate: TP-130 expand Step N checkboxes`

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
