# TP-094 — `compute_activity_segment_stats` raw-stream analyzer

**Created:** 2026-05-20
**Size:** M

## Review Level: 2

**Assessment:** New raw-stream analyzer with careful validation and formula behavior.
**Score:** 5/8 — Blast radius: 1, Pattern novelty: 2, Security: 0, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-094-activity-segment-stats/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Implement `compute_activity_segment_stats`, the only analyzer that intentionally touches raw streams, for mean/median/p90/decoupling/drift/NP/IF over a specified time or distance segment inside one activity. It must stay behind canonical stream-key tests and emit all analyzer meta.

## Dependencies

- **Task:** TP-008 (stream-key canonicalization exists)
- **Task:** TP-088 (`analysis-formulas` resource exists)
- **Task:** TP-089 (analyzer meta skeleton exists)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- internal/tools/get_activity_streams.go — stream retrieval and include_full conventions.
- internal/streams/canonicalizer.go — canonical keys.
- docs/prd/PRD-icuvisor.md — segment stats analyzer non-goal/exception.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `internal/analysis/segment_stats*.go`
- `internal/tools/compute_activity_segment_stats*.go`
- `internal/tools/get_activity_streams*.go`
- `internal/streams/*`
- `internal/tools/catalog.go`
- `internal/tools/catalog_test.go`
- `web/content/reference/tools.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Define segment selection and metrics

- [ ] Define request fields for activity ID, metric, time range and/or distance range.
- [ ] Specify validation for ambiguous or out-of-range segments.
- [ ] Define supported stats and formula refs for decoupling/drift/NP/IF.

### Step 2: Implement stream slicing and stats

- [ ] Fetch canonical stream keys only for requested metrics.
- [ ] Slice by time or distance using normalized units.
- [ ] Compute mean/median/p90 and formula-based derived stats with minimum samples.

### Step 3: Register tool and tests

- [ ] Register in `full` and describe as the raw-stream exception.
- [ ] Add fixtures for time segment, distance segment, missing stream, and insufficient sample.
- [ ] Assert mandatory meta and source_tools.

### Step 4: Verify

- [ ] Run stream canonicalization tests plus full suite/build/lint.
- [ ] Update docs/reference and CHANGELOG.md.
- [ ] Record performance/token considerations in STATUS.md.

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

- Tool computes segment stats from canonical streams only.
- Invalid ranges and missing streams produce short actionable errors/meta.
- Terse response does not include raw samples.
- `make test`, `make build`, and `make lint` pass.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-094` for traceability. Examples:

- `feat(TP-094): complete step 1 — scope current behavior`
- `fix(TP-094): repair regression found during analyzer tests`
- `test(TP-094): add golden coverage for roadmap behavior`
- `hydrate: TP-094 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not return raw stream samples by default.
- Do not bypass stream-key canonicalization.

---

## Amendments

_Add amendments below this line only._
