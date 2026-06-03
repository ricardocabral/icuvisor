# Task: TP-143 - Workout repeat header syntax regression

**Created:** 2026-06-03
**Size:** S

## Review Level: 1 (Plan Only)

**Assessment:** Adds focused WorkoutDoc serializer/parser/validation regression coverage. Low blast radius and no auth/data risk, but workout writes break visibly if repeat headers serialize incorrectly.
**Score:** 2/8 — Blast radius: 1, Pattern novelty: 0, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-143-workout-repeat-syntax-regression/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Lock down WorkoutDoc repeat-header serialization so repeat blocks emit canonical `3x` / `Main Set 3x` headers, never malformed step-like lines such as `-3 x` or `- 3x`. A competitor bug showed this breaks workout builders; Icuvisor should have explicit regression coverage for repeat headers before relying on structured write tools.

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

- `internal/workoutdoc/serialize.go`
- `internal/workoutdoc/parse.go`
- `internal/workoutdoc/workoutdoc_test.go`
- `internal/workoutdoc/validate_test.go`
- `internal/tools/validate_workout_test.go`
- `internal/tools/add_or_update_event_test.go`
- `internal/tools/create_workout_test.go`
- `CHANGELOG.md`

## Steps

> **Hydration:** STATUS.md tracks outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

### Step 0: Preflight
- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

### Step 1: Audit repeat serialization and validation
- [ ] Inspect WorkoutDoc serialize/parse/validate tests for repeat headers with and without descriptions.
- [ ] Confirm write-tool tests exercise repeat blocks through validation and event/workout serialization.
- [ ] Record any missing edge cases in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/workoutdoc ./internal/tools`.

**Artifacts:**
- `internal/workoutdoc/serialize.go`
- `internal/workoutdoc/parse.go`
- `internal/workoutdoc/workoutdoc_test.go`
- `internal/workoutdoc/validate_test.go`
- `internal/tools/validate_workout_test.go`
- `internal/tools/add_or_update_event_test.go`

### Step 2: Add repeat syntax regressions
- [ ] Add tests asserting repeat headers serialize as `3x` or `<description> 3x` without a leading dash.
- [ ] Add parse/validate coverage rejecting or warning on malformed `-3 x` / `- 3x` style lines when appropriate.
- [ ] Add at least one write-tool regression showing a repeat workout_doc produces canonical DSL.
- [ ] Run targeted tests: `go test ./internal/workoutdoc ./internal/tools`.

**Artifacts:**
- `internal/workoutdoc/serialize.go`
- `internal/workoutdoc/parse.go`
- `internal/workoutdoc/workoutdoc_test.go`
- `internal/workoutdoc/validate_test.go`
- `internal/tools/validate_workout_test.go`
- `internal/tools/add_or_update_event_test.go`

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
- `CHANGELOG.md` — note workout repeat syntax regression coverage.

**Check If Affected:**
- `docs/prd/PRD-icuvisor.md` — update only if WorkoutDoc contract changes materially.
- `docs/dogfood/v0.3-prompts.md` — update only if write-path dogfood should include repeat syntax explicitly.

## Completion Criteria

- [ ] All steps complete
- [ ] All tests passing
- [ ] Documentation updated

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-143): complete Step N — description`
- **Bug fixes:** `fix(TP-143): description`
- **Tests:** `test(TP-143): description`
- **Hydration:** `hydrate: TP-143 expand Step N checkboxes`

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
