# TP-089 — Analyzer skeleton and mandatory `_meta` contract

**Created:** 2026-05-20
**Size:** M

## Review Level: 2

**Assessment:** Foundational scaffolding for multiple future tools; requires careful contract review.
**Score:** 5/8 — Blast radius: 2, Pattern novelty: 2, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-089-analyzer-skeleton-meta/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Create shared analyzer scaffolding so every analyzer emits `_meta.method`, `_meta.source_tools`, `_meta.n`, `_meta.missing_days`, `_meta.missing_action`, and `_meta.insufficient_sample` with golden-file coverage. This is the contract all v0.6 analyzer tools build on.

## Dependencies

- **Task:** TP-087 (`analysis_metric` enum exists)
- **Task:** TP-088 (`analysis-formulas` resource exists)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- internal/tools/* — existing tool handler patterns.
- internal/response/meta.go — shared meta behavior.
- docs/prd/PRD-icuvisor.md — analyzer design rules.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `internal/analysis/meta*.go`
- `internal/analysis/window*.go`
- `internal/tools/analyzer_common*.go`
- `internal/response/meta.go`
- `internal/tools/*_test.go`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Design shared meta structs

- [ ] Define typed meta structs/helpers for method/source_tools/n/missing/insufficient-sample fields.
- [ ] Decide default missing action (`skip`) and minimum-sample helper shape.
- [ ] Map formula refs to the resource added in TP-088.

### Step 2: Implement skeleton helpers

- [ ] Add reusable response/meta builders for analyzer tools.
- [ ] Add test utilities/golden fixtures for analyzer responses.
- [ ] Keep helpers small and internal.

### Step 3: Add a no-op/demo test path

- [ ] Create tests that prove a sample analyzer response contains all mandatory meta fields.
- [ ] Assert terse/full behavior and missing-day handling at helper level.
- [ ] Run targeted tests.

### Step 4: Verify

- [ ] Run full quality gate.
- [ ] Update CHANGELOG.md if public scaffolding is visible.
- [ ] Record conventions in STATUS.md for downstream tasks.

### Step 5: Testing & Verification

- [ ] Run targeted tests added/affected by this task
- [ ] Run FULL test suite: `make test`
- [ ] Build passes: `make build`
- [ ] Lint passes: `make lint`
- [ ] Fix all failures or document pre-existing unrelated failures in STATUS.md

### Step 6: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] `STATUS.md` discoveries and final notes updated
- [ ] Commit at step boundary with the task ID in the message

## Documentation Requirements

**Must Update:**
- CHANGELOG.md — record user-visible behavior under [Unreleased] if code or docs behavior changes.
- STATUS.md — keep execution state current.

**Check If Affected:**
- README.md — update if public setup/tool behavior changes.
- web/content/reference/tools.md — update if tool catalog descriptions or generated docs are affected.
- docs/prd/PRD-icuvisor.md — check only if behavior intentionally diverges from product scope.

## Completion Criteria

- Shared analyzer helpers produce mandatory `_meta` consistently.
- Golden tests pin the contract.
- Downstream analyzer tasks can reuse the helpers without duplicating meta code.
- `make test`, `make build`, and `make lint` pass.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-089` for traceability. Examples:

- `feat(TP-089): complete step 1 — scope current behavior`
- `fix(TP-089): repair regression found during analyzer tests`
- `test(TP-089): add golden coverage for roadmap behavior`
- `hydrate: TP-089 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not let any analyzer omit mandatory meta fields.
- Do not forward-fill missing days by default.

---

## Amendments

_Add amendments below this line only._
