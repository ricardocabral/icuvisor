# TP-088 — MCP Resource `icuvisor://analysis-formulas`

**Created:** 2026-05-20
**Size:** M

## Review Level: 2

**Assessment:** New MCP resource and canonical definitions that analyzers will depend on.
**Score:** 4/8 — Blast radius: 1, Pattern novelty: 2, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-088-analysis-formulas-resource/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Add an MCP Resource containing one-paragraph canonical analyzer formula definitions for HR drift, Pw:HR decoupling, polarization index, EF, VI, and z-score, with citations. Analyzer responses will link to these entries via `_meta.formula_ref` instead of restating long math inline.

## Dependencies

- **Task:** TP-031 (MCP Resources registry pattern exists)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- internal/resources/* — resource registration patterns.
- internal/mcp/registrar_resources.go — MCP resource wiring.
- docs/prd/PRD-icuvisor.md — analyzer formula registry requirement.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `internal/resources/analysis_formulas*.go`
- `internal/resources/registry.go`
- `internal/mcp/registrar_resources.go`
- `internal/resources/*_test.go`
- `web/content/reference/resources-prompts.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Draft canonical formulas

- [ ] Write concise definitions for HR drift, Pw:HR decoupling, polarization index, EF, VI, and z-score.
- [ ] Include citations to accepted public sources; avoid copying proprietary/copyrighted text.
- [ ] Assign stable formula refs/anchors.

### Step 2: Implement resource

- [ ] Add `icuvisor://analysis-formulas` to the resources registry.
- [ ] Return a compact markdown or JSON shape consistent with existing resources.
- [ ] Add tests for resource listing and read content.

### Step 3: Wire docs and catalog

- [ ] Update resource reference docs.
- [ ] Ensure analyzer tasks can reference stable formula refs.
- [ ] Run targeted resource/MCP tests.

### Step 4: Verify

- [ ] Run full suite/build/lint.
- [ ] Update CHANGELOG.md.
- [ ] Record formula-source decisions in STATUS.md.

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

- Resource is listed and readable through MCP.
- Formula refs are stable and documented.
- Definitions are concise and cited.
- `make test`, `make build`, and `make lint` pass.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-088` for traceability. Examples:

- `feat(TP-088): complete step 1 — scope current behavior`
- `fix(TP-088): repair regression found during analyzer tests`
- `test(TP-088): add golden coverage for roadmap behavior`
- `hydrate: TP-088 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not copy long copyrighted formula text; summarize in original words with citations.
- Do not silently change formulas without tests.

---

## Amendments

_Add amendments below this line only._
