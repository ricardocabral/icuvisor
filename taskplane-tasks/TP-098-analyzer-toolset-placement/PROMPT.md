# TP-098 — Analyzer toolset placement and core-promotion gate

**Created:** 2026-05-20
**Size:** S

## Review Level: 1

**Assessment:** Catalog/tier policy change dependent on benchmark evidence.
**Score:** 3/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-098-analyzer-toolset-placement/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Place the analyzer family in `full` by default and add a benchmark-gated path to promote `analyze_trend`, `compute_zone_time`, and `compute_baseline` to `core` only after KR5 shows net token savings versus fetch-and-reduce baselines.

## Dependencies

- **Task:** TP-030 (`ICUVISOR_TOOLSET` tiers exist)
- **Task:** TP-091 (`analyze_trend` exists)
- **Task:** TP-093 (`compute_zone_time` and `compute_baseline` exist)
- **Task:** TP-100 (benchmark harness extension provides promotion evidence)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- internal/safety/toolset.go — toolset tier policy.
- internal/tools/catalog_tiers_test.go — catalog tier tests.
- docs/kr5-benchmark.md — benchmark reporting.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `internal/safety/toolset*.go`
- `internal/tools/catalog*.go`
- `internal/tools/catalog_tiers_test.go`
- `internal/toolcatalog/*`
- `docs/kr5-benchmark.md`
- `web/content/reference/tools.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Audit analyzer tier placement

- [ ] List all analyzer-family tools and current registered tiers.
- [ ] Confirm all are `full` before benchmark evidence.
- [ ] Define core-promotion acceptance note tied to KR5 benchmark results.

### Step 2: Enforce placement in tests

- [ ] Add/adjust catalog-tier tests so analyzer family defaults to `full`.
- [ ] Add conditional/policy test or doc note for future promotion of the three candidate core tools.
- [ ] Ensure `icuvisor_list_advanced_capabilities` advertises hidden analyzer capabilities clearly.

### Step 3: Apply promotion if evidence exists

- [ ] If TP-100 benchmark results are present and positive, promote only `analyze_trend`, `compute_zone_time`, and `compute_baseline` to `core`.
- [ ] If evidence is absent/negative, leave all analyzers in `full` and document why.
- [ ] Update docs accordingly.

### Step 4: Verify

- [ ] Run catalog/toolset tests and full quality gate.
- [ ] Update CHANGELOG.md.

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

- Analyzer family is in `full` by default unless benchmark evidence justifies selected core promotion.
- Tier tests encode the policy.
- Advanced-capabilities discoverability remains accurate.
- `make test`, `make build`, and `make lint` pass.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-098` for traceability. Examples:

- `feat(TP-098): complete step 1 — scope current behavior`
- `fix(TP-098): repair regression found during analyzer tests`
- `test(TP-098): add golden coverage for roadmap behavior`
- `hydrate: TP-098 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not promote analyzer tools to `core` without benchmark evidence.
- Do not hide analyzers from `full`.

---

## Amendments

_Add amendments below this line only._
