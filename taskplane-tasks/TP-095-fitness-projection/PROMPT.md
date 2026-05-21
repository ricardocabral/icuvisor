# TP-095 — `get_fitness_projection` analyzer-family tool

**Created:** 2026-05-20
**Size:** M

## Review Level: 2

**Assessment:** New public analyzer-family tool with deterministic modeling assumptions.
**Score:** 5/8 — Blast radius: 1, Pattern novelty: 2, Security: 0, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-095-fitness-projection/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Add `get_fitness_projection` so projection and analysis land together in v0.6. It should simulate forward CTL/ATL/TSB under documented assumptions such as ramp percentage, recovery-week cadence, and horizon, then return the projected curve plus modeled assumptions in `_meta`.

## Dependencies

- **Task:** TP-010 (`get_fitness` exists)
- **Task:** TP-089 (analyzer meta skeleton exists)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- internal/tools/get_fitness.go and internal/intervals/fitness.go — current fitness response shape.
- docs/upstream-gaps/periodization-parameters.md — upstream planning-parameter gap context.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `internal/analysis/fitness_projection*.go`
- `internal/tools/get_fitness_projection*.go`
- `internal/tools/get_fitness*.go`
- `internal/intervals/fitness*.go`
- `internal/tools/catalog.go`
- `internal/tools/catalog_test.go`
- `web/content/reference/tools.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Define projection assumptions

- [ ] Define request fields for horizon date/days, ramp %, recovery-week cadence, and optional planned load input.
- [ ] Document calculation assumptions and boundaries in `_meta`.
- [ ] Reject unsupported free-form physiology models.

### Step 2: Implement projection engine

- [ ] Use current fitness values as starting point.
- [ ] Project CTL/ATL/TSB deterministically over the horizon.
- [ ] Return terse summary by default and curve series behind `include_full` if needed.

### Step 3: Register and test

- [ ] Add tool registration in the analyzer family/toolset.
- [ ] Add golden tests for standard ramp, recovery week, invalid inputs, and insufficient current fitness data.
- [ ] Assert mandatory analyzer meta.

### Step 4: Docs and verification

- [ ] Update docs/reference with assumptions.
- [ ] Update CHANGELOG.md.
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

- Tool produces deterministic projections with assumptions in `_meta`.
- Invalid ramp/horizon inputs are rejected clearly.
- Terse/full behavior is tested.
- `make test`, `make build`, and `make lint` pass.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-095` for traceability. Examples:

- `feat(TP-095): complete step 1 — scope current behavior`
- `fix(TP-095): repair regression found during analyzer tests`
- `test(TP-095): add golden coverage for roadmap behavior`
- `hydrate: TP-095 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not claim predictive certainty; label modeled assumptions clearly.
- Do not require hidden upstream periodization fields.

---

## Amendments

_Add amendments below this line only._
