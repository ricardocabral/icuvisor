# Task: TP-138 - Weekly report timezone and stale-data guardrails

**Created:** 2026-06-03
**Size:** S

## Review Level: 1 (Plan Only)

**Assessment:** Prompt and test hardening around existing timezone/as-of behavior. Low implementation risk, but important for preventing misleading weekly reports.
**Score:** 3/8 — Blast radius: 1, Pattern novelty: 0, Security: 0, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-138-weekly-report-timezone-guardrails/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Harden Icuvisor's weekly-review/report guidance so assistants do not mix partial current-day metrics into prior-week summaries or silently reuse stale briefing/report data. Competing tools recently showed wrong-date weekly and daily report bugs; Icuvisor should make athlete-local date windows and `_meta.as_of` caveats explicit in prompts and tests.

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

- `internal/prompts/catalog.go`
- `internal/prompts/catalog_test.go`
- `internal/prompts/testdata/weekly_review.md`
- `internal/prompts/testdata/plan_health_review.md`
- `internal/tools/as_of_test.go`
- `internal/tools/get_wellness_data_test.go`
- `internal/tools/get_today_test.go`
- `CHANGELOG.md`

## Steps

> **Hydration:** STATUS.md tracks outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

### Step 0: Preflight
- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

### Step 1: Audit weekly/date-window safeguards
- [ ] Inspect prompt text and tests for weekly review, plan-health review, wellness reads, and `_meta.as_of`.
- [ ] Identify whether prompts explicitly forbid including wellness after the requested report window.
- [ ] Record any missing stale-date guardrails in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/prompts ./internal/tools`.

**Artifacts:**
- ``internal/prompts/catalog.go`
- ``internal/prompts/catalog_test.go`
- ``internal/prompts/testdata/weekly_review.md`
- ``internal/prompts/testdata/plan_health_review.md`
- ``internal/tools/as_of_test.go`
- ``internal/tools/get_wellness_data_test.go`

### Step 2: Add prompt and regression coverage
- [ ] Update weekly/plan-health prompt guidance to anchor all report windows in athlete-local dates and treat current-day `_meta.as_of` as partial-day context only.
- [ ] Add or strengthen golden tests so stale/current-day caveats are preserved in prompt output.
- [ ] Add targeted tool tests only if an existing `_meta.as_of` edge case is uncovered.
- [ ] Run targeted tests: `go test ./internal/prompts ./internal/tools`.

**Artifacts:**
- ``internal/prompts/catalog.go`
- ``internal/prompts/catalog_test.go`
- ``internal/prompts/testdata/weekly_review.md`
- ``internal/prompts/testdata/plan_health_review.md`
- ``internal/tools/as_of_test.go`
- ``internal/tools/get_wellness_data_test.go`

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
- `CHANGELOG.md` — note prompt/test hardening for weekly report date windows.

**Check If Affected:**
- `docs/prd/PRD-icuvisor.md` — update only if prompt catalog behavior changes materially.
- `docs/dogfood/v0.2-prompts.md` — update if dogfood weekly-review scenarios should check stale-date behavior.

## Completion Criteria

- [ ] All steps complete
- [ ] All tests passing
- [ ] Documentation updated

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-138): complete Step N — description`
- **Bug fixes:** `fix(TP-138): description`
- **Tests:** `test(TP-138): description`
- **Hydration:** `hydrate: TP-138 expand Step N checkboxes`

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
