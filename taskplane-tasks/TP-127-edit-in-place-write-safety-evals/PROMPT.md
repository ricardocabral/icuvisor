# Task: TP-127 - Edit-in-place write safety evals

**Created:** 2026-05-29
**Size:** M

## Review Level: 2 (Plan and Code)

**Assessment:** Touches write-tool safety guidance and adversarial/eval tests. It does not change credential handling, but it protects against destructive model behavior.
**Score:** 5/8 — Blast radius: 1, Pattern novelty: 1, Security: 1, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-127-edit-in-place-write-safety-evals/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Ensure assistants modify workouts/events in place instead of deleting and recreating when users ask for changes. Forum users and competing connectors are relying on prompt bias here; icuvisor should enforce this through explicit tool guidance, safety-mode behavior, and eval/adversarial coverage.

## Evidence from forum review

- Edit vs delete/recreate discussion: https://forum.intervals.icu/t/126632/229 through https://forum.intervals.icu/t/126632/233
- End-to-end existing-workout update avoided duplicate Garmin workout: https://forum.intervals.icu/t/126632/244

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `docs/prd/PRD-icuvisor.md` — destructive operations, registration-time gating, and workout write semantics.
- `docs/safety/adversarial-prompts.md` — existing safety-prompt coverage.
- `scripts/eval/README.md` — eval harness shape.
- `CONTRIBUTING.md` — test and Go workflow expectations.

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `internal/tools/add_or_update_event.go`
- `internal/tools/update_workout.go`
- `internal/tools/delete_workout.go`
- `internal/tools/delete_event.go`
- `internal/tools/*delete*_test.go`
- `docs/safety/adversarial-prompts.md`
- `scripts/eval/scenarios/cookbook_scenarios.json`
- `web/content/cookbook/build-workouts.md`
- `CHANGELOG.md`

## Steps

> **Hydration:** STATUS.md tracks outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

### Step 1: Audit write/delete guidance
- [ ] Inspect create/update/delete workout and event tool descriptions, schemas, and safety tests.
- [ ] Identify whether existing descriptions already prefer update/edit in place and where eval coverage is missing.
- [ ] Record the current safety contract and any token-budget tradeoff in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/tools`

**Artifacts:**
- `internal/tools/add_or_update_event.go`
- `internal/tools/update_workout.go`
- `internal/tools/delete_workout.go`
- `internal/tools/delete_event.go`

### Step 2: Add eval/adversarial coverage
- [ ] Add at least one eval/adversarial scenario where the user asks to change tomorrow’s workout and the assistant must choose update/edit tools, not delete/create.
- [ ] Assert safe-mode/delete-mode messaging remains short and actionable when deletion is unavailable.
- [ ] Run targeted tests: `make eval-validate` and `go test ./internal/tools`

**Artifacts:**
- `scripts/eval/scenarios/cookbook_scenarios.json`
- `docs/safety/adversarial-prompts.md`
- `internal/tools/*delete*_test.go`

### Step 3: Harden guidance if necessary
- [ ] Update concise tool descriptions or cookbook prompts only where tests show ambiguity.
- [ ] Do not add a model-controlled `confirm` flag or weaken registration-time gating.
- [ ] Run targeted tests: `go test ./internal/tools`

**Artifacts:**
- `internal/tools/add_or_update_event.go`
- `internal/tools/update_workout.go`
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
- `CHANGELOG.md` — note safety eval/guidance additions.

**Check If Affected:**
- `docs/prd/PRD-icuvisor.md` — update only if public write/delete behavior changes.
- `web/content/explain/safety-modes.md` — update if user-facing safety wording changes.

## Completion Criteria

- [ ] All steps complete
- [ ] All tests passing
- [ ] Documentation updated

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-127): complete Step N — description`
- **Bug fixes:** `fix(TP-127): description`
- **Tests:** `test(TP-127): description`
- **Hydration:** `hydrate: TP-127 expand Step N checkboxes`

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
