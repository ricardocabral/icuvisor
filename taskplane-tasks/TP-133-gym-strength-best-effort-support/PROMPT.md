# Task: TP-133 - Gym and strength best-effort support plan

**Created:** 2026-05-29
**Size:** M

## Review Level: 1 (Plan Only)

**Assessment:** Mostly scopes/docs current best-effort behavior and future work. It should avoid prematurely implementing unsupported strength-training endpoints.
**Score:** 3/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-133-gym-strength-best-effort-support/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Address recurring demand for gym/strength support without overbuilding beyond current upstream/API readiness. Document what icuvisor can safely do today (e.g. schedule time blocks/notes if supported) and capture the upstream gaps or tool changes needed for first-class strength training later.

## Evidence from forum review

- LeCoach user requested gym support and would accept allocated Gym time: https://forum.intervals.icu/t/117602/381
- Maintainer said gym support is high on roadmap: https://forum.intervals.icu/t/117602/382

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `docs/prd/PRD-icuvisor.md` — out-of-scope and future strength-training/tool catalog boundaries.
- `ROADMAP.md` — v1.x strength training endpoint placeholder.
- `CONTRIBUTING.md` — docs/test workflow expectations.

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `docs/upstream-gaps/strength-training.md`
- `docs/prd/PRD-icuvisor.md`
- `ROADMAP.md`
- `web/content/cookbook/season-and-block-plan.md`
- `web/content/cookbook/build-workouts.md`
- `internal/prompts/testdata/weekly_planning.md`
- `CHANGELOG.md`

## Steps

> **Hydration:** STATUS.md tracks outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

### Step 1: Scope current support and upstream gaps
- [ ] Inspect current event/workout category handling and PRD/Roadmap strength-training mentions.
- [ ] Determine what can be represented today without inventing unsupported structured strength sets.
- [ ] Create or update an upstream-gap note for strength/gym support if missing.
- [ ] Run targeted checks/tests as relevant.

**Artifacts:**
- `docs/upstream-gaps/strength-training.md`
- `docs/prd/PRD-icuvisor.md`
- `ROADMAP.md`

### Step 2: Add best-effort prompt/docs guidance
- [ ] Update cookbook/prompt guidance to allow scheduling simple gym time blocks or notes when the user wants that, while explicitly saying detailed strength sets are future scope unless upstream support exists.
- [ ] Avoid adding a new write tool in this task unless upstream API support is already documented in this repository.
- [ ] Run targeted tests: `go test ./internal/prompts` if prompt fixtures change.

**Artifacts:**
- `web/content/cookbook/season-and-block-plan.md`
- `web/content/cookbook/build-workouts.md`
- `internal/prompts/testdata/weekly_planning.md`

### Step 3: Capture follow-up implementation criteria
- [ ] Record in docs what evidence is needed before adding first-class strength-training tools: upstream endpoints, schema fields, response shape, and safe write behavior.
- [ ] Update ROADMAP/PRD only if this clarifies existing future scope, not to expand v1 commitments.
- [ ] Run docs/test validation as available.

**Artifacts:**
- `docs/upstream-gaps/strength-training.md`
- `ROADMAP.md`
- `docs/prd/PRD-icuvisor.md`

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
- `docs/upstream-gaps/strength-training.md` — create/update strength/gym API gap and best-effort support.
- `CHANGELOG.md` — note docs/prompt guidance changes.

**Check If Affected:**
- `docs/prd/PRD-icuvisor.md` — update only to clarify existing future scope.
- `ROADMAP.md` — update only if best-effort/future strength scope changes.

## Completion Criteria

- [ ] All steps complete
- [ ] All tests passing
- [ ] Documentation updated

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-133): complete Step N — description`
- **Bug fixes:** `fix(TP-133): description`
- **Tests:** `test(TP-133): description`
- **Hydration:** `hydrate: TP-133 expand Step N checkboxes`

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
