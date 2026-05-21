# TP-091 — `analyze_trend`, `analyze_distribution`, `analyze_correlation`, `analyze_efforts_delta`

**Created:** 2026-05-20
**Size:** L

## Review Level: 3

**Assessment:** Large new public tool family with deterministic math and golden tests; use full review.
**Score:** 6/8 — Blast radius: 2, Pattern novelty: 2, Security: 0, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-091-analyze-tools-cluster/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Implement the inferential analyzer tools for trend, distribution, correlation, and efforts-delta analysis using the shared analyzer metric and meta contracts. These tools should answer common training-analysis prompts without the LLM fetching rows and reducing them in chat.

## Dependencies

- **Task:** TP-087 (`analysis_metric` enum exists)
- **Task:** TP-088 (`analysis-formulas` resource exists)
- **Task:** TP-089 (analyzer meta skeleton exists)
- **Task:** TP-090 (auto-lap signal exists for efforts/compliance consumers)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- docs/prd/PRD-icuvisor.md — analyzer tool catalog and non-goals.
- internal/tools/get_fitness.go, get_training_summary.go, get_best_efforts.go, get_activities.go, get_wellness_data.go — source reads to reuse.
- internal/analysis/* — shared metric/meta helpers.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `internal/analysis/trend*.go`
- `internal/analysis/distribution*.go`
- `internal/analysis/correlation*.go`
- `internal/analysis/efforts*.go`
- `internal/tools/analyze_trend*.go`
- `internal/tools/analyze_distribution*.go`
- `internal/tools/analyze_correlation*.go`
- `internal/tools/analyze_efforts_delta*.go`
- `internal/tools/catalog.go`
- `internal/tools/catalog_test.go`
- `web/content/reference/tools.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Design request/response contracts

- [ ] Define schemas for windows, baseline windows, metrics, sport filters, lag days, and `include_full`.
- [ ] Ensure every tool uses the closed metric enum and analyzer meta helpers.
- [ ] Decide source read clients needed for each analyzer.

### Step 2: Implement computations

- [ ] Implement trend rolling mean/slope/delta, distribution histogram/quantiles, correlation Pearson/Spearman with lag, and efforts current-vs-baseline delta.
- [ ] Skip and count missing days; enforce minimum n with `insufficient_sample`.
- [ ] Keep raw rows out of terse responses.

### Step 3: Register tools and descriptions

- [ ] Add tool files and catalog registration in `full` by default.
- [ ] Descriptions must lead with activation hints and tell the LLM not to roll its own reductions.
- [ ] Link `_meta.formula_ref` where formulas apply.

### Step 4: Tests and verification

- [ ] Add deterministic fixtures/golden tests for each analyzer, including missing data and insufficient sample.
- [ ] Add auto-lap propagation test for `analyze_efforts_delta` if it consumes intervals.
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

- Four analyzer tools are registered and fixture-tested.
- All responses include mandatory analyzer meta fields.
- Unknown metrics are rejected through the shared enum path.
- Terse responses omit raw rows; `include_full` reveals series where intended.
- `make test`, `make build`, and `make lint` pass.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-091` for traceability. Examples:

- `feat(TP-091): complete step 1 — scope current behavior`
- `fix(TP-091): repair regression found during analyzer tests`
- `test(TP-091): add golden coverage for roadmap behavior`
- `hydrate: TP-091 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not fetch raw streams for these analyzers unless explicitly required by the specific tool contract.
- Do not silently impute missing days.

---

## Amendments

_Add amendments below this line only._
