# Task: TP-128 - Plan health review prompt

**Created:** 2026-05-29
**Size:** M

## Review Level: 2 (Plan and Code)

**Assessment:** Adds or modifies user-facing MCP prompt guidance using existing analyzer tools. It must avoid pretending to be an autonomous coach or hidden physiology model.
**Score:** 4/8 — Blast radius: 1, Pattern novelty: 2, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-128-plan-health-review-prompt/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Create a transparent plan-health review workflow that uses existing deterministic tools to explain planned-vs-completed adherence, load/form trajectory, deload/ramp caveats, and race-date risk. Competitors are shipping plan-health dashboards; icuvisor can offer a local, formula-transparent prompt instead of a black-box score.

## Evidence from forum review

- LeCoach weekly review / Plan Health release: https://forum.intervals.icu/t/117602/372
- User report of AI-created deload criticized by the plan-health feature: https://forum.intervals.icu/t/117602/381
- Plan-health scoring refinements: https://forum.intervals.icu/t/117602/383 and https://forum.intervals.icu/t/117602/390
- Montis weekly overview contract: https://forum.intervals.icu/t/117856/476

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `docs/prd/PRD-icuvisor.md` — prompts, analyzer tools, and coaching-scope boundaries.
- `ROADMAP.md` — future plan-filler and science-guardrail direction.
- `internal/prompts/testdata/weekly_review.md` — existing weekly-review behavior.
- `CONTRIBUTING.md` — test and Go workflow expectations.

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `internal/prompts/catalog.go`
- `internal/prompts/catalog_test.go`
- `internal/prompts/testdata/*.md`
- `web/content/cookbook/weekly-review.md`
- `web/content/cookbook/season-and-block-plan.md`
- `web/content/explain/fitness-projection.md`
- `CHANGELOG.md`

## Steps

> **Hydration:** STATUS.md tracks outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

### Step 1: Design plan-health prompt contract
- [ ] Inspect existing `weekly_review`, `weekly_planning`, `race_week_taper`, analyzer tools, and cookbook pages.
- [ ] Decide whether to add a new `plan_health_review` prompt or extend `weekly_review` without duplicating TP-122 season-planning scope.
- [ ] Define required tool sequence: events/training plan, fitness/projection, planned-vs-completed compliance, recent wellness, and caveats for deload/recovery weeks.
- [ ] Run targeted tests: `go test ./internal/prompts`

**Artifacts:**
- `internal/prompts/catalog.go`
- `internal/prompts/testdata/*.md`
- `STATUS.md Discoveries`

### Step 2: Implement prompt and golden tests
- [ ] Add or update prompt text with transparent scoring/caveats, explicit missing-data handling, and no hidden black-box score unless computed from surfaced values.
- [ ] Add/update prompt registry golden tests.
- [ ] Ensure prompt asks for a reviewed proposal before any calendar writes.
- [ ] Run targeted tests: `go test ./internal/prompts`

**Artifacts:**
- `internal/prompts/catalog.go`
- `internal/prompts/catalog_test.go`
- `internal/prompts/testdata/*.md`

### Step 3: Document cookbook workflow
- [ ] Add cookbook guidance showing when to use weekly review vs plan-health review vs season planning.
- [ ] Include caveats for deload weeks, planned races, and incomplete wellness/readiness data.
- [ ] Run targeted tests: `make test` or relevant docs validation if available.

**Artifacts:**
- `web/content/cookbook/weekly-review.md`
- `web/content/cookbook/season-and-block-plan.md`
- `web/content/explain/fitness-projection.md`

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
- `CHANGELOG.md` — note new/changed prompt and cookbook workflow.

**Check If Affected:**
- `docs/prd/PRD-icuvisor.md` — update only if prompt catalog scope changes materially.
- `ROADMAP.md` — update only if this changes future phase assumptions.

## Completion Criteria

- [ ] All steps complete
- [ ] All tests passing
- [ ] Documentation updated

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-128): complete Step N — description`
- **Bug fixes:** `fix(TP-128): description`
- **Tests:** `test(TP-128): description`
- **Hydration:** `hydrate: TP-128 expand Step N checkboxes`

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
