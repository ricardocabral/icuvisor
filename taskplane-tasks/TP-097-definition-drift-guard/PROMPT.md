# TP-097 — Definition-drift guard for canonical formulas

**Created:** 2026-05-20
**Size:** M

## Review Level: 2

**Assessment:** Test/guard task over canonical public math definitions.
**Score:** 4/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-097-definition-drift-guard/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Pin canonical formula definitions and analyzer outputs so decoupling, drift, polarization, EF, VI, and z-score cannot silently change. Future changes to definitions must be treated as breaking/product decisions, not incidental refactors.

## Dependencies

- **Task:** TP-088 (`analysis-formulas` resource exists)
- **Task:** TP-091 (`analyze_*` tools exist)
- **Task:** TP-093 (`compute_*` tools exist)
- **Task:** TP-094 (`compute_activity_segment_stats` exists)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- internal/resources/analysis_formulas*.go — formula text/refs.
- internal/analysis/* — computation implementations.
- internal/toolchecks/schema_stability_test.go — precedent for stability guards.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `internal/analysis/**/*`
- `internal/resources/analysis_formulas*.go`
- `internal/toolchecks/*`
- `internal/tools/*_test.go`
- `testdata/analysis/**/*`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Inventory formula-sensitive code

- [ ] List all formulas and analyzer tools that depend on each.
- [ ] Record stable formula IDs/refs and expected outputs.
- [ ] Decide golden fixture layout.

### Step 2: Add golden-file guards

- [ ] Add tests pinning formulas/resource text hashes or stable IDs.
- [ ] Add analyzer golden outputs for decoupling, drift, polarization, EF, VI, and z-score.
- [ ] Ensure tests fail loudly when formulas drift.

### Step 3: Document breaking-change policy

- [ ] Add code notes or docs explaining that formula definition changes are breaking.
- [ ] Update contributing/tool catalog docs if needed.
- [ ] Update CHANGELOG.md only if behavior changes during this task.

### Step 4: Verify

- [ ] Run targeted analysis/resource/toolcheck tests.
- [ ] Run full quality gate.

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

- Golden tests pin canonical formula behavior.
- Formula refs and analyzer outputs cannot drift silently.
- Policy is documented for future contributors.
- `make test`, `make build`, and `make lint` pass.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-097` for traceability. Examples:

- `feat(TP-097): complete step 1 — scope current behavior`
- `fix(TP-097): repair regression found during analyzer tests`
- `test(TP-097): add golden coverage for roadmap behavior`
- `hydrate: TP-097 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not change formula definitions while adding guards unless explicitly required and documented.

---

## Amendments

_Add amendments below this line only._
