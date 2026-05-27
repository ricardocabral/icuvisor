# Task: TP-106 - Weekly review MCP prompt

**Created:** 2026-05-26
**Size:** S

## Review Level: 1

**Assessment:** Prompt registry addition using established golden-test patterns; user-visible but limited to prompt text and metadata.
**Score:** 2/8 — Blast radius: 1, Pattern novelty: 0, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-106-weekly-review-prompt/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Add a curated `weekly_review` MCP prompt that guides assistants through a structured review of the previous week and preview of the upcoming week using existing icuvisor tools. The prompt should produce consistent, safe reviews without adding a new analysis engine.

Tracking issue: https://github.com/ricardocabral/icuvisor/issues/33

## Dependencies

- **Task:** TP-032 (MCP prompts registry exists)
- **Task:** TP-091 (analyze tools exist)
- **Task:** TP-093 (compute tools exist)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — repository rules and prompt conventions.
- `docs/prd/PRD-icuvisor.md` — prompt/tool catalog product contract.
- `internal/prompts/catalog.go` — existing prompt registration patterns.
- `internal/prompts/catalog_test.go` — prompt golden-test patterns.
- `internal/prompts/testdata/*.md` — expected prompt text style.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None. Unit tests must not hit the network.

## File Scope

- `internal/prompts/catalog.go`
- `internal/prompts/catalog_test.go`
- `internal/prompts/testdata/*`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Existing prompt style and golden-test format reviewed

### Step 1: Add `weekly_review` prompt registration

- [ ] Register `weekly_review` with title, description, arguments, tool list, instructions, and return format.
- [ ] Include arguments for `week_start`, `lookback_days`, and `include_next_week` if they fit existing prompt argument conventions.
- [ ] Instruct the assistant to establish athlete-local dates/timezone before comparing days.
- [ ] Instruct the assistant not to use write/delete tools without explicit user approval.
- [ ] Run targeted prompt tests.

**Artifacts:**
- `internal/prompts/catalog.go` (modified)

### Step 2: Add golden tests

- [ ] Add a golden file for default weekly-review rendering.
- [ ] Add a test for explicit arguments, or document why the existing prompt test style only needs one golden.
- [ ] Ensure the prompt degrades gracefully when full-toolset helpers are unavailable by mentioning `icuvisor_list_advanced_capabilities` where appropriate.
- [ ] Run targeted prompt tests.

**Artifacts:**
- `internal/prompts/catalog_test.go` (modified)
- `internal/prompts/testdata/weekly_review.md` (new)

### Step 3: Changelog and full verification

- [ ] Update `CHANGELOG.md` under `[Unreleased]` for the new prompt.
- [ ] Run targeted prompt tests.
- [ ] Review generated prompt/tool docs if applicable.

**Artifacts:**
- `CHANGELOG.md` (modified)

### Step 4: Testing & Verification

- [ ] Run targeted tests added/affected by this task
- [ ] Run FULL test suite: `make test`
- [ ] Build passes: `make build`
- [ ] Lint passes: `make lint`
- [ ] Fix all failures or document pre-existing unrelated failures in STATUS.md

### Step 5: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] `STATUS.md` discoveries and final notes updated
- [ ] Commit at step boundary with the task ID in the message

## Documentation Requirements

**Must Update:**
- `CHANGELOG.md` — record the new curated prompt under `[Unreleased]`.
- `STATUS.md` — keep execution state current.

**Check If Affected:**
- Generated prompt reference/docs — regenerate if prompt docs are generated.
- `README.md` — update only if curated prompt list is manually maintained there.
- `docs/prd/PRD-icuvisor.md` — update only if prompt scope changes product contract.

## Completion Criteria

- `weekly_review` is registered in the MCP prompt catalog.
- Prompt has deterministic rendering and golden-test coverage.
- Prompt guidance covers athlete-local dates/timezone, planned-vs-completed comparison, wellness staleness, and no writes without approval.
- `make test`, `make build`, and `make lint` pass or pre-existing failures are documented.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-106` for traceability. Examples:

- `feat(TP-106): complete step 1 — add weekly review prompt`
- `test(TP-106): add weekly review golden prompt`
- `hydrate: TP-106 expand step checkboxes`

## Do NOT

- Do not add a new analysis tool for this task; this is prompt-only unless tests require small registry support.
- Do not instruct assistants to write/delete calendar items without explicit user approval.
- Do not include private athlete examples in golden files.
- Do not modify protected docs without explicit approval.

---

## Amendments

_Add amendments below this line only._
