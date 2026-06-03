# Task: TP-136 - Activity interval-source clarity in details and routing

**Created:** 2026-06-03
**Size:** S

## Review Level: 1 (Plan Only)

**Assessment:** Small read-path clarity task adapting existing interval-source classification. It touches activity detail/interval tests and tool descriptions, with low safety risk.
**Score:** 3/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-136-activity-interval-source-clarity/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Make lap/interval-source evidence clear when assistants analyze completed workouts. Forum users specifically asked for Garmin/device laps instead of auto-generated Intervals.icu intervals; Icuvisor already classifies `get_activity_intervals`, but `get_activity_details` and tool-routing wording should prevent LLMs from making interval-execution claims without checking the interval source.

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

- `internal/tools/get_activity_details.go`
- `internal/tools/get_activity_details_test.go`
- `internal/analysis/interval_source.go`
- `internal/analysis/interval_source_test.go`
- `internal/tools/schema_snapshot/*activity*`
- `CHANGELOG.md`

## Steps

> **Hydration:** STATUS.md tracks outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

### Step 0: Preflight
- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

### Step 1: Audit current interval-source exposure
- [ ] Inspect `get_activity_details`, `get_activity_intervals`, interval-source tests, and tool descriptions.
- [ ] Decide whether to expose interval-source metadata on `get_activity_details`, strengthen descriptions only, or both.
- [ ] Record the decision and any upstream limitation in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/analysis`.

**Artifacts:**
- ``internal/tools/get_activity_details.go`
- ``internal/tools/get_activity_details_test.go`
- ``internal/analysis/interval_source.go`
- ``internal/analysis/interval_source_test.go`
- ``internal/tools/schema_snapshot/*activity*`
- ``CHANGELOG.md`

### Step 2: Implement clarity and regression coverage
- [ ] Add or update tests showing device-lap/auto-lap/structured-workout source metadata is surfaced or routed correctly.
- [ ] Update tool descriptions/schema snapshots as needed so assistants know to call `get_activity_intervals` before analyzing laps/reps.
- [ ] Ensure terse defaults stay compact and `include_full` remains the raw-payload opt-in.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/analysis`.

**Artifacts:**
- ``internal/tools/get_activity_details.go`
- ``internal/tools/get_activity_details_test.go`
- ``internal/analysis/interval_source.go`
- ``internal/analysis/interval_source_test.go`
- ``internal/tools/schema_snapshot/*activity*`
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
- `CHANGELOG.md` — note interval-source clarity or routing regression coverage.

**Check If Affected:**
- `docs/prd/PRD-icuvisor.md` — update only if response fields or tool contracts change materially.
- `docs/dogfood/v0.2-prompts.md` — update only if interval-source dogfood prompt coverage should change.

## Completion Criteria

- [ ] All steps complete
- [ ] All tests passing
- [ ] Documentation updated

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-136): complete Step N — description`
- **Bug fixes:** `fix(TP-136): description`
- **Tests:** `test(TP-136): description`
- **Hydration:** `hydrate: TP-136 expand Step N checkboxes`

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
