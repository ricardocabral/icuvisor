# TP-099 — Upstream coverage audit for zone-time/load-balance analyzers

**Created:** 2026-05-20
**Size:** S

## Review Level: 1

**Assessment:** Audit/documentation task with optional small script.
**Score:** 2/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-099-upstream-coverage-audit/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Measure how often `compute_zone_time` and `compute_load_balance` can use precomputed per-activity zone times across the v0.2 fixture set instead of falling back to stream math. If fallback is too frequent, file/document the intervals.icu API gap.

## Dependencies

- **Task:** TP-093 (`compute_zone_time` and `compute_load_balance` exist)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- internal/intervals/testdata and internal/tools/testdata — v0.2 fixture corpus.
- docs/upstream-gaps/ — existing gap documentation pattern.
- docs/prd/PRD-icuvisor.md — upstream-coverage audit requirement.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `scripts/*`
- `internal/analysis/zone_time*.go`
- `internal/tools/compute_zone_time*.go`
- `internal/tools/compute_load_balance*.go`
- `docs/upstream-gaps/*`
- `docs/kr5-benchmark.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Define measurement method

- [ ] Identify fixture corpus and fields that count as precomputed zone times.
- [ ] Define output metrics: fixture count, precomputed count, fallback count, unknown count.
- [ ] Choose threshold or mark threshold as operator decision if not already agreed.

### Step 2: Implement/run audit

- [ ] Add a small script or Go test/helper to scan fixtures and report coverage.
- [ ] Run it against the v0.2 fixture set.
- [ ] Record results in STATUS.md and a durable doc if user-facing.

### Step 3: Document gap or close it

- [ ] If fallback exceeds threshold/looks risky, create `docs/upstream-gaps/zone-time-coverage.md` with evidence and feature-request text.
- [ ] If coverage is sufficient, document the result in `docs/kr5-benchmark.md` or STATUS.md.
- [ ] Do not change analyzer behavior except to fix discovered coverage bugs.

### Step 4: Verify

- [ ] Run relevant tests and full quality gate if code/scripts changed.
- [ ] Update CHANGELOG.md only for user-visible changes.

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

- Coverage measurement is reproducible from repository fixtures.
- Result states precomputed vs fallback coverage clearly.
- An upstream gap doc exists if coverage is inadequate.
- No secrets or live athlete data are required.
- Tests/build/lint pass if code changed.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-099` for traceability. Examples:

- `feat(TP-099): complete step 1 — scope current behavior`
- `fix(TP-099): repair regression found during analyzer tests`
- `test(TP-099): add golden coverage for roadmap behavior`
- `hydrate: TP-099 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not hit live intervals.icu from automated tests.
- Do not silently lower analyzer quality thresholds.

---

## Amendments

_Add amendments below this line only._
