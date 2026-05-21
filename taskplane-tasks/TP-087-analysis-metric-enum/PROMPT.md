# TP-087 — `analysis_metric` closed enum and unknown-metric hints

**Created:** 2026-05-20
**Size:** M

## Review Level: 2

**Assessment:** Creates a shared analyzer contract that future tools depend on.
**Score:** 4/8 — Blast radius: 1, Pattern novelty: 2, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-087-analysis-metric-enum/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Introduce the closed `analysis_metric` enum shared by analyzer tools, with validation that rejects unknown or free-form arithmetic inputs and returns concise hints toward the right analyzer or supported metric.

## Dependencies

- **Task:** TP-030 (`core`/`full` toolset tiers exist)
- **Task:** TP-007 (response shaping/meta conventions exist)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- docs/prd/PRD-icuvisor.md — Planned analyzers design rules.
- internal/tools/decode.go — strict input decoding patterns.
- internal/tools/catalog.go — schema registration patterns.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `internal/analysis/metrics*.go`
- `internal/tools/analyzer_metric*.go`
- `internal/tools/decode*.go`
- `internal/tools/catalog*.go`
- `web/content/reference/tools.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Design the enum and aliases

- [ ] List supported metric identifiers from existing read-tool fields.
- [ ] Decide canonical names and any safe aliases; reject derived/free-form expressions.
- [ ] Document hint strategy for unknown metrics.

### Step 2: Implement validation helpers

- [ ] Add shared enum parsing/JSON schema helpers under `internal/analysis` or a small tools helper.
- [ ] Return short actionable invalid-argument errors with hints.
- [ ] Keep validation reusable by all analyzer tools.

### Step 3: Tests

- [ ] Add table-driven parsing/schema tests for valid metrics, aliases, unknown names, and arithmetic expressions.
- [ ] Assert error text is concise and does not expose internals.
- [ ] Run targeted tests.

### Step 4: Docs and verification

- [ ] Update analyzer docs/reference stubs if present.
- [ ] Run full quality gate.
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

- Closed enum exists and is reusable by analyzer tools.
- Unknown metrics and free-form arithmetic are rejected with hints.
- Schema descriptions enumerate supported metrics.
- `make test`, `make build`, and `make lint` pass.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-087` for traceability. Examples:

- `feat(TP-087): complete step 1 — scope current behavior`
- `fix(TP-087): repair regression found during analyzer tests`
- `test(TP-087): add golden coverage for roadmap behavior`
- `hydrate: TP-087 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not accept arbitrary mathematical expressions from the LLM.

---

## Amendments

_Add amendments below this line only._
