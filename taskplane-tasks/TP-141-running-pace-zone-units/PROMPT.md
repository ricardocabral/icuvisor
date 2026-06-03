# Task: TP-141 - Running pace-zone unit and label audit

**Created:** 2026-06-03
**Size:** S

## Review Level: 1 (Plan Only)

**Assessment:** Focused read/write test audit for existing sport-settings pace behavior. It touches profile and sport-setting tests with low safety risk when zones remain delete-mode gated.
**Score:** 3/8 — Blast radius: 1, Pattern novelty: 0, Security: 0, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-141-running-pace-zone-units/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Verify Icuvisor handles running threshold pace and pace zones with clear units and labels on both read and write paths. Adjacent products recently fixed running threshold pace conversion and zone wording; Icuvisor should have regression tests proving seconds-per-km/seconds-per-mile conversion and LLM-facing labels are unambiguous.

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

- `internal/tools/get_athlete_profile.go`
- `internal/tools/get_athlete_profile_test.go`
- `internal/tools/update_sport_settings.go`
- `internal/tools/update_sport_settings_test.go`
- `internal/tools/update_sport_settings_zones.go`
- `internal/tools/update_sport_settings_zones_test.go`
- `internal/units/*`
- `CHANGELOG.md`

## Steps

> **Hydration:** STATUS.md tracks outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

### Step 0: Preflight
- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

### Step 1: Audit run pace read/write coverage
- [ ] Inspect athlete-profile and sport-settings tests for threshold pace, pace units, and pace-zone names.
- [ ] Confirm tests cover both `seconds_per_km` and `seconds_per_mile` inputs and upstream pace unit output.
- [ ] Record any ambiguous LLM-facing wording or missing scale/unit metadata in STATUS.md.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/units`.

**Artifacts:**
- `internal/tools/get_athlete_profile.go`
- `internal/tools/get_athlete_profile_test.go`
- `internal/tools/update_sport_settings.go`
- `internal/tools/update_sport_settings_test.go`
- `internal/tools/update_sport_settings_zones.go`
- `internal/tools/update_sport_settings_zones_test.go`

### Step 2: Add pace-zone regressions and wording fixes
- [ ] Add missing tests for Run threshold pace conversion and pace zone boundary/name round trips.
- [ ] Update schema descriptions or response labels if they could be misread as speed rather than pace seconds per distance.
- [ ] Ensure zone overwrite behavior remains gated by `ICUVISOR_DELETE_MODE=full` where applicable.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/units`.

**Artifacts:**
- `internal/tools/get_athlete_profile.go`
- `internal/tools/get_athlete_profile_test.go`
- `internal/tools/update_sport_settings.go`
- `internal/tools/update_sport_settings_test.go`
- `internal/tools/update_sport_settings_zones.go`
- `internal/tools/update_sport_settings_zones_test.go`

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
- `CHANGELOG.md` — note running pace-unit regression coverage or wording fix.

**Check If Affected:**
- `docs/prd/PRD-icuvisor.md` — update only if sport-settings response or input contracts change materially.

## Completion Criteria

- [ ] All steps complete
- [ ] All tests passing
- [ ] Documentation updated

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-141): complete Step N — description`
- **Bug fixes:** `fix(TP-141): description`
- **Tests:** `test(TP-141): description`
- **Hydration:** `hydrate: TP-141 expand Step N checkboxes`

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
