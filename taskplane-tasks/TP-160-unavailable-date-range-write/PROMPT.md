# Task: TP-160 - Unavailable date-range write convenience

**Created:** 2026-06-09
**Size:** M

## Review Level: 2 (Plan and Code)

**Assessment:** Adds a new write convenience around calendar time-off categories. It is not destructive, but it changes the MCP catalog, schemas, write behavior, docs, and tests.
**Score:** 5/8 — Blast radius: 2, Pattern novelty: 1, Security: 1, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-160-unavailable-date-range-write/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Add an ergonomic date-range write for unavailable/time-off calendar entries such as Sick, Injured, and Holiday. IntervalCoach feedback shows users want to mark ranges, while icuvisor currently supports these categories and protects them during `apply_training_plan`; a range convenience should reduce repetitive single-event writes without weakening safety.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — write-tool, safety, catalog, and clean-room rules.
- `docs/prd/PRD-icuvisor.md` — write tool catalog and safety expectations.
- `ROADMAP.md` — write-path phasing context.

## Environment

- **Workspace:** Go module root
- **Services required:** None

## File Scope

- `internal/tools/add_or_update_event.go`
- `internal/tools/add_or_update_event_test.go`
- `internal/tools/delete_events_by_date_range.go`
- `internal/tools/event_category_schema_test.go`
- `internal/tools/registry.go`
- `internal/tools/schema_snapshot/*.json`
- `internal/mcp/catalog_hash.go`
- `internal/mcp/catalog_hash_test.go`
- `internal/toolcatalog/*`
- `internal/toolchecks/*`
- `README.md`
- `CHANGELOG.md`
- `docs/prd/PRD-icuvisor.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied

### Step 1: Design the range-write contract

- [ ] Choose the minimal public surface: either a dedicated tool (preferred if catalog patterns support it) such as `add_unavailable_date_range`, or an explicit range mode on `add_or_update_event` if that is safer and less duplicative.
- [ ] Define allowed categories as a closed set for unavailability only: Sick, Injured, Holiday/time-off equivalents already accepted by event category normalization.
- [ ] Define idempotency semantics for repeated calls over the same athlete/date/category range, including whether one multi-day event or per-day events are created based on upstream API capabilities.
- [ ] Run targeted tests after adding failing tests: `go test ./internal/tools -run 'Unavailable|DateRange|Event'`

**Artifacts:**
- `internal/tools/add_or_update_event_test.go` (modified)
- New/modified tool test file if a dedicated tool is added

### Step 2: Implement the write convenience and catalog integration

> ⚠️ Hydrate: Expand implementation checkboxes after inspecting current event write helpers, category normalization, and registry conventions.

- [ ] Implement the range write using existing event client interfaces and date validation; reject reversed ranges and excessive ranges with short user errors.
- [ ] Register the tool or extended schema in `internal/tools/registry.go` and update schema snapshots/catalog hash surfaces.
- [ ] Preserve existing safety: do not add delete-mode bypasses, do not overwrite workouts by default, and ensure `apply_training_plan` protection for unavailable categories remains intact.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/mcp ./internal/toolchecks`

**Artifacts:**
- `internal/tools/add_or_update_event.go` (modified or shared helper)
- New `internal/tools/add_unavailable_date_range.go` and `_test.go` if dedicated tool is chosen
- `internal/tools/registry.go` (modified)
- `internal/tools/schema_snapshot/*.json` (modified/new)
- `internal/mcp/catalog_hash.go` / tests (modified if hash fixture changes)
- `internal/toolcatalog/*` / `internal/toolchecks/*` (modified if affected)

### Step 3: Testing & Verification

- [ ] Run FULL test suite: `make test`
- [ ] Run integration tests (if applicable)
- [ ] Fix all failures
- [ ] Build passes: `make build`

### Step 4: Documentation & Delivery

- [ ] `README.md` documents the new convenience or extended event write args with examples.
- [ ] `docs/prd/PRD-icuvisor.md` updated only if this formally adds a catalog tool/contract.
- [ ] `CHANGELOG.md` notes unavailable date-range write support.
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**
- `CHANGELOG.md` — note user-visible write convenience.
- `README.md` — update tool catalog/examples.

**Check If Affected:**
- `docs/prd/PRD-icuvisor.md` — update if adding a new MCP tool.
- `docs/safety/adversarial-prompts.md` — update only if write safety prompts change.

## Completion Criteria

- [ ] Users can add unavailable/time-off ranges without creating each day manually.
- [ ] Invalid categories/ranges fail with short actionable errors.
- [ ] Catalog/schema/tests/docs are consistent.
- [ ] Full tests and build pass.

## Git Commit Convention

Commits happen at step boundaries. All commits for this task MUST include the task ID:

- **Step completion:** `feat(TP-160): complete Step N — description`
- **Bug fixes:** `fix(TP-160): description`
- **Tests:** `test(TP-160): description`
- **Hydration:** `hydrate: TP-160 expand Step N checkboxes`

## Do NOT

- Add delete behavior or a model-controlled confirmation override.
- Accept API keys or athlete credentials as tool args.
- Overwrite workouts/plans by default.
- Copy competitor implementation; use only the public feature signal.
- Commit without the task ID prefix.

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if issues discovered during execution. -->
