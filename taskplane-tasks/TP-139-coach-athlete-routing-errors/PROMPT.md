# Task: TP-139 - Coach-mode athlete routing and authorization errors

**Created:** 2026-06-03
**Size:** M

## Review Level: 2 (Plan and Code)

**Assessment:** Hardens coach-mode routing and explicit authorization errors across config/coach/tool registration paths. It touches athlete scoping and access-control behavior, so plan and code review are warranted.
**Score:** 5/8 — Blast radius: 1, Pattern novelty: 1, Security: 1, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-139-coach-athlete-routing-errors/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Make coach-mode athlete routing failures explicit and testable before broader coach-mode rollout. Forum users of competing tools hit wrong-athlete memory and authorization confusion; Icuvisor should normalize athlete IDs, distinguish local-athlete from coached-athlete flows, and return actionable unauthorized-athlete errors rather than timeout-like or generic failures.

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

- `internal/coach/*`
- `internal/config/athlete.go`
- `internal/config/athlete_test.go`
- `internal/config/coach_load_test.go`
- `internal/tools/list_athletes.go`
- `internal/tools/select_athlete.go`
- `internal/tools/*athlete*_test.go`
- `internal/mcp/*`
- `CHANGELOG.md`

## Steps

> **Hydration:** STATUS.md tracks outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

### Step 0: Preflight
- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

### Step 1: Audit coach/local athlete routing
- [ ] Inspect `internal/coach`, athlete ID normalization, `list_athletes`, and `select_athlete` behavior.
- [ ] Identify where unauthorized coached-athlete access currently becomes generic upstream errors or ambiguous state.
- [ ] Define expected public error messages that do not leak credentials or raw sensitive identifiers.
- [ ] Run targeted tests: `go test ./internal/coach ./internal/config ./internal/tools`.

**Artifacts:**
- ``internal/coach/*`
- ``internal/config/athlete.go`
- ``internal/config/athlete_test.go`
- ``internal/config/coach_load_test.go`
- ``internal/tools/list_athletes.go`
- ``internal/tools/select_athlete.go`

### Step 2: Add routing/error tests and hardening
- [ ] Add tests for normalized `i123`/numeric athlete IDs, unauthorized coached-athlete selection, and local-athlete fallback when coach mode is not active.
- [ ] Implement explicit authorization/routing errors where tests reveal ambiguity.
- [ ] Ensure tool catalog/ACL behavior still hides disallowed tools and does not accept API keys in chat/tool parameters.
- [ ] Run targeted tests: `go test ./internal/coach ./internal/config ./internal/tools ./internal/mcp`.

**Artifacts:**
- ``internal/coach/*`
- ``internal/config/athlete.go`
- ``internal/config/athlete_test.go`
- ``internal/config/coach_load_test.go`
- ``internal/tools/list_athletes.go`
- ``internal/tools/select_athlete.go`

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
- `CHANGELOG.md` — note coach-mode routing/error hardening.

**Check If Affected:**
- `docs/prd/PRD-icuvisor.md` — update only if coach-mode behavior or error contract changes materially.
- `ROADMAP.md` — update only if this unblocks or changes the coach-mode milestone wording.

## Completion Criteria

- [ ] All steps complete
- [ ] All tests passing
- [ ] Documentation updated

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-139): complete Step N — description`
- **Bug fixes:** `fix(TP-139): description`
- **Tests:** `test(TP-139): description`
- **Hydration:** `hydrate: TP-139 expand Step N checkboxes`

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
