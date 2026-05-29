# Task: TP-129 - Readiness fallback guidance for null upstream readiness

**Created:** 2026-05-29
**Size:** M

## Review Level: 2 (Plan and Code)

**Assessment:** Mostly prompt/docs and possibly wellness-response tests. It is user-visible because recovery advice must not overclaim from missing readiness data.
**Score:** 4/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-129-readiness-fallback-guidance/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Teach icuvisor prompts/docs to handle null readiness scores gracefully: if Intervals readiness is absent, assistants should state that plainly and use HRV, resting HR, sleep, fatigue/soreness/stress, and available native provider fields as cautious supporting evidence. Garmin users often have useful wellness data without a single readiness number.

## Evidence from forum review

- IntervalCoach readiness ribbon empty/null: https://forum.intervals.icu/t/120045/820 and https://forum.intervals.icu/t/120045/823
- User request to point readiness at custom/wellness fields: https://forum.intervals.icu/t/120045/822
- Garmin does not directly populate readiness for that user: https://forum.intervals.icu/t/120045/824

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `docs/prd/PRD-icuvisor.md` — wellness fields, native provider sidecars, scale labels, and terse/full behavior.
- `internal/prompts/testdata/recovery_check.md` — existing recovery prompt behavior.
- `CONTRIBUTING.md` — test and Go workflow expectations.

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `internal/tools/get_wellness_data.go`
- `internal/tools/get_wellness_data_test.go`
- `internal/prompts/testdata/recovery_check.md`
- `internal/prompts/testdata/weekly_review.md`
- `web/content/cookbook/readiness-check.md`
- `web/content/reference/tools.md`
- `CHANGELOG.md`

## Steps

> **Hydration:** STATUS.md tracks outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

### Step 1: Audit wellness readiness semantics
- [ ] Inspect wellness shaping, provenance metadata, native provider fields, and recovery/weekly prompt text.
- [ ] Identify whether null readiness already appears in missing_fields and whether prompts instruct cautious fallback.
- [ ] Record available fallback fields and non-goals in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/prompts`

**Artifacts:**
- `internal/tools/get_wellness_data.go`
- `internal/tools/get_wellness_data_test.go`
- `internal/prompts/testdata/recovery_check.md`

### Step 2: Add fallback tests or prompt guidance
- [ ] Add tests if missing for null readiness with present HRV/RHR/sleep/native fields.
- [ ] Update recovery/weekly prompts so assistants do not invent readiness scores and explain missingness before fallback interpretation.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/prompts`

**Artifacts:**
- `internal/tools/get_wellness_data_test.go`
- `internal/prompts/testdata/recovery_check.md`
- `internal/prompts/testdata/weekly_review.md`

### Step 3: Update cookbook docs
- [ ] Update readiness-check cookbook with Garmin/null-readiness fallback examples and language.
- [ ] Keep scale labels explicit and avoid device-specific claims not backed by response fields.
- [ ] Run targeted tests/docs validation as available.

**Artifacts:**
- `web/content/cookbook/readiness-check.md`

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
- `CHANGELOG.md` — note prompt/docs/test changes.

**Check If Affected:**
- `web/content/reference/tools.md` — update via generated tooling only if response/tool docs change.

## Completion Criteria

- [ ] All steps complete
- [ ] All tests passing
- [ ] Documentation updated

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-129): complete Step N — description`
- **Bug fixes:** `fix(TP-129): description`
- **Tests:** `test(TP-129): description`
- **Hydration:** `hydrate: TP-129 expand Step N checkboxes`

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
