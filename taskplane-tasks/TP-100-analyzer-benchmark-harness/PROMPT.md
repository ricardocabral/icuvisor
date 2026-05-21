# TP-100 — Extend KR5 benchmark harness for analyzer family

**Created:** 2026-05-20
**Size:** M

## Review Level: 2

**Assessment:** Benchmark harness extension with product gating impact.
**Score:** 4/8 — Blast radius: 1, Pattern novelty: 2, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-100-analyzer-benchmark-harness/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Extend the v0.4 KR5 benchmark harness to compare training-analysis prompts with and without the analyzer family, measuring token usage and raw-stream pull counts. This validates that analyzers reduce fetch-and-reduce behavior by at least the roadmap target.

## Dependencies

- **Task:** TP-034 (KR5 benchmark harness exists)
- **Task:** TP-091 (`analyze_*` tools exist)
- **Task:** TP-092 (`get_activity_histogram` exists)
- **Task:** TP-093 (`compute_*` tools exist)
- **Task:** TP-094 (`compute_activity_segment_stats` exists)
- **Task:** TP-095 (`get_fitness_projection` exists)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- docs/kr5-benchmark.md — existing benchmark report.
- scripts/* and testdata/* — current benchmark harness location.
- internal/diagnostics/recent_tool_calls.go — possible stream-pull counting precedent.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `scripts/**/*`
- `testdata/**/*`
- `docs/kr5-benchmark.md`
- `internal/diagnostics/*`
- `internal/tools/catalog*.go`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Find and extend benchmark harness

- [ ] Locate the v0.4 prompt set and token/response measurement code.
- [ ] Add analyzer-enabled and analyzer-disabled benchmark modes.
- [ ] Define stream-pull count instrumentation or log parsing.

### Step 2: Add analyzer prompt shapes

- [ ] Add trend, distribution, baseline, correlation/compliance, and single-activity histogram prompt cases.
- [ ] Ensure prompts are identical across with/without-analyzer runs except tool availability.
- [ ] Record expected source-tool usage.

### Step 3: Run and record results

- [ ] Run the benchmark against fixture/reference data.
- [ ] Record token deltas and raw-stream pull counts in `docs/kr5-benchmark.md`.
- [ ] Flag whether core-promotion candidates meet net-savings criteria for TP-098.

### Step 4: Verify

- [ ] Add tests or smoke checks so the harness does not silently break.
- [ ] Run full quality gate where applicable.
- [ ] Update CHANGELOG.md for benchmark/docs visibility.

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

- Benchmark can run with analyzers enabled/disabled.
- Report includes token deltas and stream-pull counts.
- Results explicitly inform TP-098 core-promotion decision.
- Harness is documented and reproducible.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-100` for traceability. Examples:

- `feat(TP-100): complete step 1 — scope current behavior`
- `fix(TP-100): repair regression found during analyzer tests`
- `test(TP-100): add golden coverage for roadmap behavior`
- `hydrate: TP-100 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not include live athlete payloads in benchmark fixtures.
- Do not promote tools to core in this task; record evidence for TP-098.

---

## Amendments

_Add amendments below this line only._
