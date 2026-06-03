# Task: TP-142 - Readiness provenance labels and recovery wording guardrails

**Created:** 2026-06-03
**Size:** S

## Review Level: 1 (Plan Only)

**Assessment:** Focused wellness response/prompt label hardening. Low blast radius, but important because mislabeled recovery/readiness scores can produce misleading training advice.
**Score:** 3/8 — Blast radius: 1, Pattern novelty: 0, Security: 0, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-142-readiness-provenance-labels/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Ensure Icuvisor never presents Garmin Body Battery, Oura readiness, Polar freshness, or generic upstream `readiness` as an unlabeled universal “recovery” score. Icuvisor already carries wellness provenance metadata; this task adds tests and wording guardrails so LLM-facing labels stay clear.

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

- `internal/tools/get_wellness_data.go`
- `internal/tools/get_wellness_data_test.go`
- `internal/tools/schema_snapshot/get_wellness_data.json`
- `internal/prompts/catalog.go`
- `internal/prompts/catalog_test.go`
- `internal/prompts/testdata/recovery_check.md`
- `internal/prompts/testdata/weekly_review.md`
- `CHANGELOG.md`

## Steps

> **Hydration:** STATUS.md tracks outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

### Step 0: Preflight
- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

### Step 1: Audit readiness/recovery wording
- [ ] Inspect wellness provenance shaping for Garmin, Oura, Polar, WHOOP, and unknown readiness sources.
- [ ] Inspect recovery/weekly prompts for wording that could collapse provider-native readiness into generic recovery.
- [ ] Record missing labels or ambiguous terms in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/prompts`.

**Artifacts:**
- `internal/tools/get_wellness_data.go`
- `internal/tools/get_wellness_data_test.go`
- `internal/tools/schema_snapshot/get_wellness_data.json`
- `internal/prompts/catalog.go`
- `internal/prompts/catalog_test.go`
- `internal/prompts/testdata/recovery_check.md`

### Step 2: Add provenance and prompt regressions
- [ ] Add or strengthen tests that `_meta.provenance.readiness.native_scale` is provider-specific and visible when readiness is present.
- [ ] Update prompt wording/golden tests so assistants cite provider/source and do not invent a readiness score when missing or stale.
- [ ] Ensure terse defaults remain compact and null stripping does not remove required provenance.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/prompts`.

**Artifacts:**
- `internal/tools/get_wellness_data.go`
- `internal/tools/get_wellness_data_test.go`
- `internal/tools/schema_snapshot/get_wellness_data.json`
- `internal/prompts/catalog.go`
- `internal/prompts/catalog_test.go`
- `internal/prompts/testdata/recovery_check.md`

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
- `CHANGELOG.md` — note readiness provenance wording/test hardening.

**Check If Affected:**
- `docs/prd/PRD-icuvisor.md` — update only if wellness response semantics change materially.
- `docs/dogfood/v0.2-prompts.md` — update if readiness provenance dogfood coverage should change.

## Completion Criteria

- [ ] All steps complete
- [ ] All tests passing
- [ ] Documentation updated

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-142): complete Step N — description`
- **Bug fixes:** `fix(TP-142): description`
- **Tests:** `test(TP-142): description`
- **Hydration:** `hydrate: TP-142 expand Step N checkboxes`

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
