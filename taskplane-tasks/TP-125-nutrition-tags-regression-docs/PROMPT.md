# Task: TP-125 - Activity tags and fueling regression/docs pass

**Created:** 2026-05-29
**Size:** S

## Review Level: 1 (Plan Only)

**Assessment:** Primarily protects and documents existing behavior for activity tags and fueling fields. Low novelty, but user-visible response shape requires careful tests.
**Score:** 3/8 — Blast radius: 1, Pattern novelty: 0, Security: 0, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-125-nutrition-tags-regression-docs/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Lock in the forum-validated expectation that normal activity reads include upstream tags and activity fueling fields (`carbs_ingested_g`, `carbs_used_g`) without requiring `include_full:true`. This prevents regressions and lets docs market icuvisor as nutrition-aware and tag-aware.

## Evidence from forum review

- Tags request/fix signal: https://forum.intervals.icu/t/126632/241 and https://forum.intervals.icu/t/126632/242
- Carbs request/fix signal: https://forum.intervals.icu/t/126632/246 and https://forum.intervals.icu/t/126632/247

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `docs/prd/PRD-icuvisor.md` — terse-by-default response and activity tool catalog expectations.
- `CONTRIBUTING.md` — test and Go workflow expectations.

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `internal/tools/get_activities_test.go`
- `internal/tools/get_activity_details_test.go`
- `internal/tools/get_today_test.go`
- `web/content/reference/tools.md`
- `web/content/cookbook/activity-retrospective.md`
- `web/content/cookbook/readiness-check.md`
- `CHANGELOG.md`

## Steps

> **Hydration:** STATUS.md tracks outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

### Step 1: Audit existing coverage
- [ ] Verify terse `get_activities` and `get_activity_details` tests cover present tags, empty tags, and fueling fields.
- [ ] Verify `get_today` preserves tags for completed activities and planned events.
- [ ] Record any missing coverage in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/tools`

**Artifacts:**
- `internal/tools/get_activities_test.go`
- `internal/tools/get_activity_details_test.go`
- `internal/tools/get_today_test.go`

### Step 2: Fill regression or docs gaps
- [ ] Add missing regression tests rather than changing already-correct behavior.
- [ ] Update user-facing docs/cookbook text to mention tag-aware and fueling-aware activity reads where useful.
- [ ] Avoid changing raw upstream field names; keep disambiguated grams suffixes.
- [ ] Run targeted tests: `go test ./internal/tools`

**Artifacts:**
- `internal/tools/*_test.go`
- `web/content/cookbook/activity-retrospective.md`
- `web/content/cookbook/readiness-check.md`

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
- `CHANGELOG.md` — note added regression coverage/docs if changed.

**Check If Affected:**
- `web/content/reference/tools.md` — regenerate/update only if generated tool docs need refresh.

## Completion Criteria

- [ ] All steps complete
- [ ] All tests passing
- [ ] Documentation updated

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-125): complete Step N — description`
- **Bug fixes:** `fix(TP-125): description`
- **Tests:** `test(TP-125): description`
- **Hydration:** `hydrate: TP-125 expand Step N checkboxes`

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
