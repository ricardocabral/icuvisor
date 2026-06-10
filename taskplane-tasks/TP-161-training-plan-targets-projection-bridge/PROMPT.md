# Task: TP-161 - Bridge training-plan targets into fitness projection

**Created:** 2026-06-09
**Size:** M

## Review Level: 2 (Plan and Code)

**Assessment:** Extends analyzer input semantics and metadata so weekly plan targets can drive projection loads. It is deterministic/read-only, but touches analysis contracts, schemas, and docs.
**Score:** 4/8 — Blast radius: 2, Pattern novelty: 1, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-161-training-plan-targets-projection-bridge/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Make it obvious how weekly training-plan targets feed `get_fitness_projection` instead of requiring the LLM to invent `planned_daily_loads` in chat. A public IntervalCoach report noted a goal-progress projection falling off after 7 days when future workouts were absent, even though weekly TSS targets existed. icuvisor should provide a deterministic bridge from weekly plan targets to projection load assumptions.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — analyzer/tool rules.
- `docs/prd/PRD-icuvisor.md` — analyzer and training-plan tool expectations.
- `ROADMAP.md` — analyzer phasing context.

## Environment

- **Workspace:** Go module root
- **Services required:** None

## File Scope

- `internal/analysis/fitness_projection.go`
- `internal/analysis/fitness_projection_test.go`
- `internal/tools/get_fitness_projection.go`
- `internal/tools/get_fitness_projection_test.go`
- `internal/tools/get_training_plan.go`
- `internal/tools/get_events_training_plan_test.go`
- `internal/tools/schema_snapshot/get_fitness_projection.json`
- `README.md`
- `CHANGELOG.md`
- `docs/prd/PRD-icuvisor.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied

### Step 1: Design deterministic weekly-target distribution

- [ ] Add failing tests for a weekly target bridge that converts weekly plan load/TSS targets into `planned_daily_loads` over the projection horizon.
- [ ] Define deterministic distribution rules (for example even daily distribution across days in the target week unless explicit daily loads override) and document assumptions in analyzer `_meta.assumptions`.
- [ ] Define conflict precedence: explicit `planned_daily_loads` should win for exact dates; weekly targets fill only missing future dates.
- [ ] Run targeted tests after adding failing cases: `go test ./internal/analysis ./internal/tools -run 'FitnessProjection|TrainingPlan'`

**Artifacts:**
- `internal/analysis/fitness_projection_test.go` (modified/new if missing)
- `internal/tools/get_fitness_projection_test.go` (modified)

### Step 2: Implement bridge in projection input and schema

> ⚠️ Hydrate: Expand after inspecting current `get_training_plan` output fields and upstream plan target shapes.

- [ ] Add a typed optional request field to `get_fitness_projection` for weekly plan targets (or a clearly named helper argument) with JSON Schema descriptions an LLM can use.
- [ ] Implement target-to-daily-load conversion in `internal/analysis/fitness_projection.go` or a small helper, keeping the math deterministic and testable.
- [ ] Include `_meta.source_tools`/assumptions/boundaries that tell the caller when projection load came from weekly plan targets versus explicit daily loads or modeled ramp.
- [ ] Refresh `internal/tools/schema_snapshot/get_fitness_projection.json`.
- [ ] Run targeted tests: `go test ./internal/analysis ./internal/tools -run 'FitnessProjection|TrainingPlan'`

**Artifacts:**
- `internal/analysis/fitness_projection.go` (modified)
- `internal/tools/get_fitness_projection.go` (modified)
- `internal/tools/schema_snapshot/get_fitness_projection.json` (modified)

### Step 3: Testing & Verification

- [ ] Run FULL test suite: `make test`
- [ ] Run integration tests (if applicable)
- [ ] Fix all failures
- [ ] Build passes: `make build`

### Step 4: Documentation & Delivery

- [ ] `README.md` documents the new projection bridge and precedence with an example.
- [ ] `CHANGELOG.md` notes training-plan target projection support.
- [ ] PRD reviewed/updated if analyzer contract changes.
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**
- `CHANGELOG.md` — note user-visible analyzer enhancement.
- `README.md` — update get_fitness_projection usage.

**Check If Affected:**
- `docs/prd/PRD-icuvisor.md` — update if public contract changes.

## Completion Criteria

- [ ] Weekly targets can deterministically fill future projection loads.
- [ ] Explicit planned daily loads override weekly target fill for the same date.
- [ ] Projection metadata tells users which assumptions were used.
- [ ] Full tests and build pass.

## Git Commit Convention

Commits happen at step boundaries. All commits for this task MUST include the task ID:

- **Step completion:** `feat(TP-161): complete Step N — description`
- **Bug fixes:** `fix(TP-161): description`
- **Tests:** `test(TP-161): description`
- **Hydration:** `hydrate: TP-161 expand Step N checkboxes`

## Do NOT

- Fetch training plans implicitly from inside `get_fitness_projection` unless existing tool architecture already supports that safely; prefer explicit typed inputs.
- Invent undocumented physiology models or stochastic projections.
- Copy competitor implementation; use only public behavior signal.
- Commit without the task ID prefix.

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if issues discovered during execution. -->
