# TP-093 — `compute_zone_time`, `compute_load_balance`, `compute_baseline`, `compute_compliance_rate`

**Created:** 2026-05-20
**Size:** L

## Review Level: 3

**Assessment:** Large public compute-tool cluster with multiple source tools and formula contracts.
**Score:** 6/8 — Blast radius: 2, Pattern novelty: 2, Security: 0, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-093-compute-tools-cluster/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Implement deterministic compute tools for zone time, load balance, baseline, and compliance rate. They should aggregate existing read outputs first, report missing/insufficient data explicitly, and avoid raw-stream math when upstream/precomputed zone times are available.

## Dependencies

- **Task:** TP-087 (`analysis_metric` enum exists)
- **Task:** TP-088 (`analysis-formulas` resource exists)
- **Task:** TP-089 (analyzer meta skeleton exists)
- **Task:** TP-090 (auto-lap signal exists for interval-consuming compliance behavior)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- internal/tools/get_training_summary.go, get_extended_metrics.go, get_fitness.go, get_wellness_data.go, get_events.go — likely source tools.
- internal/tools/link_activity_to_event.go — compliance pairing context.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `internal/analysis/zone_time*.go`
- `internal/analysis/load_balance*.go`
- `internal/analysis/baseline*.go`
- `internal/analysis/compliance*.go`
- `internal/tools/compute_zone_time*.go`
- `internal/tools/compute_load_balance*.go`
- `internal/tools/compute_baseline*.go`
- `internal/tools/compute_compliance_rate*.go`
- `internal/tools/catalog.go`
- `internal/tools/catalog_test.go`
- `web/content/reference/tools.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Design deterministic contracts

- [ ] Define request schemas for date windows, sport filters, zone metric, and baseline window.
- [ ] Ensure all tools use analyzer meta and formula refs where relevant.
- [ ] Decide source-tool priority before any stream fallback.

### Step 2: Implement aggregation logic

- [ ] Compute zone time and load balance from precomputed zone fields when present.
- [ ] Compute baseline mean/std/z-score with minimum sample rules.
- [ ] Compute compliance from scheduled/completed event pairings and propagate auto-lap caution where relevant.

### Step 3: Register and document activation hints

- [ ] Register the tools in `full` by default.
- [ ] Descriptions must say not to fetch rows/streams and reduce manually.
- [ ] Ensure `analyze_trend`, `compute_zone_time`, and `compute_baseline` are identifiable for later core promotion.

### Step 4: Tests and verification

- [ ] Add golden tests for precomputed data, missing days, insufficient samples, auto-lap caution, and no-precomputed fallback behavior.
- [ ] Run full quality gate and update docs/CHANGELOG.

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

- Four compute tools are registered and deterministic.
- Precomputed zone times are preferred over stream math.
- All tools include mandatory meta and missing/insufficient signals.
- `make test`, `make build`, and `make lint` pass.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-093` for traceability. Examples:

- `feat(TP-093): complete step 1 — scope current behavior`
- `fix(TP-093): repair regression found during analyzer tests`
- `test(TP-093): add golden coverage for roadmap behavior`
- `hydrate: TP-093 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not silently forward-fill or impute missing days.
- Do not default to stream math when upstream/precomputed zone times are present.

---

## Amendments

_Add amendments below this line only._
