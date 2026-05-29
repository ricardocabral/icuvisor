# Task: TP-124 - Activity date resolution and detail-routing evals

**Created:** 2026-05-29
**Size:** M

## Review Level: 2 (Plan and Code)

**Assessment:** Adds test/eval coverage for multi-tool routing across activities, details, and intervals. It should not change core API semantics but may harden descriptions/prompts.
**Score:** 4/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-124-activity-date-detail-routing-evals/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Make “analyze my race last Sunday” reliable by ensuring assistants resolve athlete-local dates to activity IDs and then fetch details or intervals when the user asks for race/rep/split analysis. Public feedback shows models sometimes claim no activity exists or claim detail endpoints are unavailable unless the user manually supplies an ID.

## Evidence from forum review

- Activity-by-date failure: https://forum.intervals.icu/t/126632/236
- Claude missed detail/interval capability: https://forum.intervals.icu/t/126632/249
- Maintainer clarified list→detail flow: https://forum.intervals.icu/t/126632/250

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `docs/prd/PRD-icuvisor.md` — activity/detail/interval tool expectations and terse/full behavior.
- `scripts/eval/README.md` — eval harness shape and validation requirements.
- `CONTRIBUTING.md` — test and Go workflow expectations.

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `scripts/eval/scenarios/cookbook_scenarios.json`
- `scripts/eval/README.md`
- `internal/tools/get_activities.go`
- `internal/tools/get_activity_details.go`
- `internal/tools/get_activity_details_test.go`
- `internal/prompts/testdata/*.md`
- `web/content/cookbook/activity-retrospective.md`
- `CHANGELOG.md`

## Steps

> **Hydration:** STATUS.md tracks outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

### Step 1: Map current routing hints
- [ ] Inspect `get_activities`, `get_activity_details`, `get_activity_intervals`, cookbook prompts, and eval scenarios.
- [ ] Identify where prompts/tool descriptions fail to instruct list-by-date before detail/interval fetch.
- [ ] Record any gaps and chosen changes in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/prompts`

**Artifacts:**
- `internal/tools/get_activities.go`
- `internal/tools/get_activity_details.go`
- `internal/prompts/testdata/*.md`

### Step 2: Add eval scenarios
- [ ] Add at least one eval scenario for “analyze my race last Sunday” that requires activity-date lookup then detail/interval fetch.
- [ ] Add one scenario for “show/compare lap splits or reps for my run on [date]” that must not stop at session summaries.
- [ ] Validate expected tool-use ordering and grounding rubric.
- [ ] Run targeted tests: `make eval-validate`

**Artifacts:**
- `scripts/eval/scenarios/cookbook_scenarios.json`
- `scripts/eval/README.md`

### Step 3: Harden descriptions or cookbook guidance
- [ ] If gaps are found, update tool descriptions or cookbook text to make the list→detail/interval path explicit.
- [ ] Avoid adding broad tool-description tokens unless the eval requires them; prefer concise activation hints.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/prompts` and `make eval-validate`

**Artifacts:**
- `internal/tools/get_activities.go`
- `internal/tools/get_activity_details.go`
- `web/content/cookbook/activity-retrospective.md`

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
- `CHANGELOG.md` — document new eval/prompt hardening or tool-description changes.

**Check If Affected:**
- `web/content/cookbook/activity-retrospective.md` — update if user-facing prompt guidance changes.
- `web/content/reference/tools.md` — update via generated tooling only if tool catalog text changes.

## Completion Criteria

- [ ] All steps complete
- [ ] All tests passing
- [ ] Documentation updated

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-124): complete Step N — description`
- **Bug fixes:** `fix(TP-124): description`
- **Tests:** `test(TP-124): description`
- **Hydration:** `hydrate: TP-124 expand Step N checkboxes`

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
