# Task: TP-122 - Season planning prompt and context hardening

**Created:** 2026-05-29
**Size:** M

## Review Level: 2 (Plan and Code)

**Assessment:** This changes user-facing MCP prompts and potentially prompt registry tests, with moderate product impact but no destructive API writes. The task must avoid introducing an opinionated auto-ATP scheduler or unreviewed physiology model.
**Score:** 4/8 — Blast radius: 1, Pattern novelty: 2, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-122-season-planning-prompt-context/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Respond to demand for season-planning workflows by strengthening icuvisor's safe, deterministic planning guidance. Improve prompts/tests so assistants gather race events, active training plan, current load, recent compliance, and available write tools before proposing season or week plans. Do not add an automated ATP calendar writer in this task.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `docs/prd/PRD-icuvisor.md` — prompts, training-plan, and future planning scope.
- `ROADMAP.md` — v2.x planning/fill-calendar direction and out-of-scope boundaries.
- `CONTRIBUTING.md` — test and Go workflow expectations.

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `internal/prompts/catalog.go`
- `internal/prompts/catalog_test.go`
- `internal/prompts/testdata/*.md`
- `internal/tools/add_or_update_event.go`
- `internal/tools/add_or_update_event_test.go`
- `docs/prd/PRD-icuvisor.md`
- `ROADMAP.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no GPL/copyleft competitor source is opened or copied; use only the public product signal summarized in this task.

### Step 1: Design the safe season-planning guidance surface

- [ ] Inspect existing `weekly_planning`, `weekly_review`, and `race_week_taper` prompts and tests.
- [ ] Decide whether to enhance existing prompts only or add a new prompt such as `season_planning` using the existing prompt registry pattern.
- [ ] Ensure the plan uses existing deterministic tools (`get_events`, `get_training_plan`, `get_fitness`, `get_training_summary`, `compute_compliance_rate`, `icuvisor_list_advanced_capabilities`) rather than inventing an ATP writer.
- [ ] Record the chosen approach and non-goals in STATUS.md Discoveries.
- [ ] Run targeted tests: `go test ./internal/prompts`

**Artifacts:**
- `internal/prompts/catalog.go` (inspected/modified)
- `internal/prompts/catalog_test.go` (inspected/modified)
- STATUS.md Discoveries (updated by worker)

### Step 2: Implement prompt and golden-test updates

- [ ] Update existing prompt text or add a new prompt so season/race planning asks for race priority/date context, current load, active plan, planned events, recent completion, and user approval before writes.
- [ ] Add guardrails that explicitly avoid automatic calendar filling or ATP-note creation without a reviewed proposal and user approval.
- [ ] Update golden prompt fixtures and tests for the prompt registry.
- [ ] Run targeted tests: `go test ./internal/prompts`

**Artifacts:**
- `internal/prompts/catalog.go` (modified)
- `internal/prompts/catalog_test.go` (modified)
- `internal/prompts/testdata/*.md` (modified/added)

### Step 3: Strengthen race-event write examples if needed

- [ ] Check whether `add_or_update_event` examples clearly show race categories (`RACE_A`, `RACE_B`, `RACE_C`), sport type, date, distance, expected duration, and target load.
- [ ] Add or adjust one input example/test if current race-event coverage is insufficient for planning prompts.
- [ ] Keep write behavior unchanged; this step is examples/schema-test hardening only unless tests reveal a bug.
- [ ] Run targeted tests: `go test ./internal/tools -run 'AddOrUpdateEvent|InputExamples|EventCategory'`

**Artifacts:**
- `internal/tools/add_or_update_event.go` (modified if needed)
- `internal/tools/add_or_update_event_test.go` (modified if needed)

### Step 4: Testing & Verification

- [ ] Run FULL test suite: `make test`
- [ ] Run lint: `make lint`
- [ ] Fix all failures
- [ ] Build passes: `make build`

### Step 5: Documentation & Delivery

- [ ] `CHANGELOG.md` updated for user-visible prompt/example changes.
- [ ] `ROADMAP.md` checked to ensure future ATP/fill-calendar scope remains accurate; update only if the task clarifies the path.
- [ ] `docs/prd/PRD-icuvisor.md` checked for prompt catalog drift; update only if a new prompt is added or prompt scope changes materially.
- [ ] Discoveries logged in STATUS.md.

## Documentation Requirements

**Must Update:**
- `CHANGELOG.md` — user-visible prompt/example changes.

**Check If Affected:**
- `ROADMAP.md` — season-planning/fill-calendar future scope.
- `docs/prd/PRD-icuvisor.md` — prompt catalog and planning behavior.

## Completion Criteria

- [ ] Prompt guidance supports season-planning context gathering without adding an automated ATP writer.
- [ ] Golden prompt tests pass and lock the new behavior.
- [ ] Race-event write examples are sufficient for an assistant to use existing tools safely.
- [ ] Full verification passes.

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-122): complete Step N — description`
- **Bug fixes:** `fix(TP-122): description`
- **Tests:** `test(TP-122): description`
- **Hydration:** `hydrate: TP-122 expand Step N checkboxes`

## Do NOT

- Do not add a deterministic ATP/season-plan calendar writer in this task.
- Do not introduce new physiology formulas or coaching models.
- Do not read, copy, paraphrase, or port GPL/copyleft competitor source code.
- Do not call write/delete tools without explicit prompt guardrails requiring user approval.
- Do not skip golden prompt tests.

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if issues discovered during execution.
     Format:
     ### Amendment N — YYYY-MM-DD HH:MM
     **Issue:** [what was wrong]
     **Resolution:** [what was changed] -->
