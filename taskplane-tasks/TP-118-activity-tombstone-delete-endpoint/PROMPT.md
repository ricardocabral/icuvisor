# Task: TP-118 - Activity tombstone delete endpoint

**Created:** 2026-05-29
**Size:** S

## Review Level: 2 (Plan and Code)

**Assessment:** This touches a destructive activity-delete path and must preserve icuvisor's registration-time delete-mode safety contract. The implementation should be small, but the behavior affects athlete data deletion semantics.
**Score:** 4/8 — Blast radius: 1, Pattern novelty: 1, Security: 2, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-118-activity-tombstone-delete-endpoint/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Verify whether the current Intervals.icu activity deletion endpoint should use the newly observed OpenAPI path `/api/v1/activity/{id}/tombstone` rather than the existing direct activity delete path. Update icuvisor's activity-delete client behavior and regression tests if needed, while preserving delete-mode gating and terse, actionable user errors.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `docs/prd/PRD-icuvisor.md` — destructive-operation safety and activity tool behavior.
- `ROADMAP.md` — current release priorities and delete/write scope.
- `CONTRIBUTING.md` — test and Go workflow expectations.

## Environment

- **Workspace:** repository root
- **Services required:** None; tests must use httptest/stubs only and must not hit intervals.icu.

## File Scope

- `internal/intervals/delete.go`
- `internal/intervals/delete_test.go`
- `internal/tools/delete_activity.go`
- `internal/tools/delete_tools_test.go`
- `internal/tools/schema_snapshot/delete_activity.json`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public upstream API behavior and existing icuvisor code/tests.

### Step 1: Determine the correct activity deletion contract

- [ ] Inspect existing activity delete implementation and tests to identify current request path, method, safety checks, and response behavior.
- [ ] Check the repository's own OpenAPI/docs/test fixtures if present for activity tombstone semantics; do not load competitor implementation code.
- [ ] Decide whether `DeleteActivity` should call `/activity/{id}/tombstone`, keep `/activity/{id}`, or support a documented fallback; record the decision in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/intervals ./internal/tools -run 'DeleteActivity|delete_activity|DeleteTools'`

**Artifacts:**
- `internal/intervals/delete.go` (modified if endpoint changes)
- `internal/intervals/delete_test.go` (modified)
- `internal/tools/delete_activity.go` (modified if tool metadata needs endpoint wording)

### Step 2: Implement and lock the endpoint behavior

- [ ] Update the client path/method only if Step 1 proves the tombstone endpoint is the correct public API contract.
- [ ] Add or update httptest coverage asserting the exact method/path for activity deletion, including target-athlete mismatch safety behavior.
- [ ] If public tool metadata mentions the source endpoint, update tool/schema snapshot expectations to match.
- [ ] Run targeted tests: `go test ./internal/intervals ./internal/tools -run 'DeleteActivity|delete_activity|DeleteTools|Schema'`

**Artifacts:**
- `internal/intervals/delete.go` (modified if needed)
- `internal/intervals/delete_test.go` (modified)
- `internal/tools/delete_activity.go` (modified if needed)
- `internal/tools/schema_snapshot/delete_activity.json` (modified if needed)

### Step 3: Testing & Verification

- [ ] Run FULL test suite: `make test`
- [ ] Run lint if source changed: `make lint`
- [ ] Fix all failures
- [ ] Build passes: `make build`

### Step 4: Documentation & Delivery

- [ ] `CHANGELOG.md` updated if behavior changes or a regression test documents an upstream API change.
- [ ] `ROADMAP.md` checked for any needed adjustment; update only if this affects public delete/write scope.
- [ ] Discoveries logged in STATUS.md, including the final endpoint decision and evidence source.

## Documentation Requirements

**Must Update:**
- `CHANGELOG.md` — add an Unreleased note if delete endpoint behavior or test coverage changes materially.

**Check If Affected:**
- `ROADMAP.md` — update only if the behavior changes roadmap-visible write/delete scope.
- `docs/prd/PRD-icuvisor.md` — update only if destructive-operation semantics need product wording.

## Completion Criteria

- [ ] Activity deletion endpoint decision is explicit and covered by tests.
- [ ] Delete-mode gating remains registration-time only; no `confirm` argument is introduced.
- [ ] All targeted and full verification commands pass.
- [ ] Documentation/discoveries updated as required.

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `fix(TP-118): complete Step N — description`
- **Bug fixes:** `fix(TP-118): description`
- **Tests:** `test(TP-118): description`
- **Hydration:** `hydrate: TP-118 expand Step N checkboxes`

## Do NOT

- Do not read, copy, paraphrase, or port GPL/copyleft competitor source code.
- Do not hit the live Intervals.icu API from tests.
- Do not add model-controlled delete confirmation arguments.
- Do not broaden delete-mode behavior beyond this endpoint decision.
- Do not skip tests.

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if issues discovered during execution.
     Format:
     ### Amendment N — YYYY-MM-DD HH:MM
     **Issue:** [what was wrong]
     **Resolution:** [what was changed] -->
