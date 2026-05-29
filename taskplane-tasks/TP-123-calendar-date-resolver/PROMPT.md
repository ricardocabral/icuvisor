# Task: TP-123 - Calendar date resolver and future date anchors

**Created:** 2026-05-29
**Size:** M

## Review Level: 2 (Plan and Code)

**Assessment:** Adds or hardens a user-facing deterministic date surface used by planning conversations. The blast radius is moderate because it touches tool/catalog and date semantics, but it is read-only and easily reverted.
**Score:** 4/8 — Blast radius: 1, Pattern novelty: 2, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-123-calendar-date-resolver/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Prevent day/date hallucinations in training-plan conversations by giving assistants deterministic athlete-local calendar anchors for today and future dates. Public forum feedback showed models inventing wrong weekday/date pairings even in fresh chats; icuvisor should make the correct local calendar mechanically available instead of relying on model arithmetic.

## Evidence from forum review

- IcuSync timezone/date issues: https://forum.intervals.icu/t/126632/220 and https://forum.intervals.icu/t/126632/225
- Explicit check_dates suggestion: https://forum.intervals.icu/t/126632/227
- LeCoach tomorrow/date confusion: https://forum.intervals.icu/t/117602/381

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `docs/prd/PRD-icuvisor.md` — date/timezone, tool catalog, terse-response, and planning boundaries.
- `ROADMAP.md` — current release phase and prompt/tool scope.
- `CONTRIBUTING.md` — test and Go workflow expectations.

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `internal/tools/*calendar*`
- `internal/tools/catalog.go`
- `internal/tools/catalog_test.go`
- `internal/toolcatalog/*`
- `internal/tools/schema_snapshot/*`
- `scripts/eval/scenarios/cookbook_scenarios.json`
- `web/content/guides/claude-project-instructions.md`
- `CHANGELOG.md`

## Steps

> **Hydration:** STATUS.md tracks outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

### Step 1: Design deterministic date surface
- [ ] Inspect existing `_meta.as_of`, `get_today`, `get_activities`, `get_events`, and prompt guidance for date anchors.
- [ ] Decide whether to add a small read-only tool such as `resolve_calendar_dates` or to harden existing date metadata/prompts without a new tool.
- [ ] Document the chosen surface and non-goals in STATUS.md Discoveries, including why it avoids model date arithmetic.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/toolcatalog`

**Artifacts:**
- `internal/tools/*calendar* (new/modified if tool is added)`
- `internal/tools/get_today.go (inspected/modified if metadata is reused)`
- `STATUS.md Discoveries`

### Step 2: Implement date anchors and tests
- [ ] Implement the chosen deterministic date anchor behavior using athlete timezone, local date, weekday, and offsets.
- [ ] Add tests covering current day, future day offsets, timezone boundaries, and invalid input if a new tool is added.
- [ ] Update catalog/schema snapshots if the public tool surface changes.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/toolcatalog`

**Artifacts:**
- `internal/tools/*calendar*`
- `internal/tools/*_test.go`
- `internal/tools/schema_snapshot/*`

### Step 3: Add activation guidance and eval coverage
- [ ] Add or update eval/cookbook prompt text so prompts that mention future weeks or “tomorrow” use the deterministic date anchor.
- [ ] Add an eval scenario for a known-bad weekday/date pairing such as “Monday May 26” when the local date says otherwise.
- [ ] Ensure guidance does not ask the assistant to infer dates from UTC.
- [ ] Run targeted tests: `make eval-validate`

**Artifacts:**
- `scripts/eval/scenarios/cookbook_scenarios.json`
- `web/content/guides/claude-project-instructions.md`

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
- `CHANGELOG.md` — note the new or hardened date-resolution behavior.

**Check If Affected:**
- `docs/prd/PRD-icuvisor.md` — update only if the public tool catalog changes.
- `web/content/reference/tools.md` — update via generated tooling only if tool catalog changes.

## Completion Criteria

- [ ] All steps complete
- [ ] All tests passing
- [ ] Documentation updated

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-123): complete Step N — description`
- **Bug fixes:** `fix(TP-123): description`
- **Tests:** `test(TP-123): description`
- **Hydration:** `hydrate: TP-123 expand Step N checkboxes`

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
