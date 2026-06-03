# Task: TP-150 - Workout DSL metric suffix from sport priority

**Created:** 2026-06-03
**Size:** M

## Review Level: 2 (Plan and Code)

**Assessment:** This changes workout serialization semantics for planned workouts and can affect calendar write behavior. It adapts existing WorkoutDoc serialization patterns and needs code review for compatibility.
**Score:** 4/8 — Blast radius: 1, Pattern novelty: 2, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-150-workout-dsl-metric-suffix/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Fix the latent planned-workout bug where zone DSL can target the wrong metric for athletes whose sport priority order is not the Intervals default. Public WorkoutContext behavior showed a recent fix for Run athletes configured with running power: planned run DSL must use suffixes like `Z2 Power`, `Z2 HR`, or `Z2 Pace` based on sport settings, otherwise Intervals can interpret bare `Z2` as the wrong metric and render/load planned workouts incorrectly. Implement this clean-room from icuvisor's own code and upstream behavior only.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — project clean-room and MCP/tool conventions.
- `docs/prd/PRD-icuvisor.md` — planned-workout/write behavior requirements.
- `internal/workoutdoc/syntax.go` — supported DSL target units.
- `internal/workoutdoc/workoutdoc_test.go` — existing serialization golden coverage.

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `internal/workoutdoc/*`
- `internal/tools/add_or_update_event.go`
- `internal/tools/create_workout.go`
- `internal/tools/update_workout.go`
- `internal/tools/*workout*test.go`
- `internal/tools/schema_snapshot/add_or_update_event.json`
- `internal/tools/schema_snapshot/create_workout.json`
- `internal/tools/schema_snapshot/update_workout.json`
- `web/content/cookbook/build-workouts.md`
- `web/content/reference/resources-prompts.md`
- `docs/prd/PRD-icuvisor.md`
- `CHANGELOG.md`

## Steps

> **Hydration:** STATUS.md tracks outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm current tests lock bare power-zone serialization (for example `POWER_ZONE` → `Z2`) so the behavioral gap is explicit

### Step 1: Design the sport-aware suffix boundary

**Plan-review checkpoint** — Review the chosen boundary before implementation.

- [ ] Decide whether suffix selection belongs in `workoutdoc.Serialize`, a new options-aware wrapper, or event/workout write call sites
- [ ] Preserve existing serializer behavior for contexts that do not know sport settings
- [ ] Define expected suffix behavior for `POWER_HR_PACE`, `HR_POWER_PACE`, and `PACE_HR_POWER`
- [ ] Document any upstream ambiguity discovered in STATUS.md rather than guessing

**Artifacts:**
- `internal/workoutdoc/*` (modified if options live there)
- `internal/tools/add_or_update_event.go` (modified if call-site wrapper lives there)
- `STATUS.md` (discoveries)

### Step 2: Implement and test metric suffix behavior

- [ ] Add regression tests for a Run planned workout with `workout_order=POWER_HR_PACE` producing power-targeted DSL (`Z2 Power` / equivalent)
- [ ] Add regression coverage for HR-primary and pace-primary order where applicable
- [ ] Implement the smallest code change that passes the tests without changing unrelated DSL output
- [ ] Run targeted tests: `go test ./internal/workoutdoc ./internal/tools -run 'Workout|AddOrUpdateEvent|CreateWorkout|UpdateWorkout'`

**Artifacts:**
- `internal/workoutdoc/*` (modified as needed)
- `internal/tools/*workout*test.go` (modified/new tests)
- `internal/tools/add_or_update_event_test.go` (modified/new tests if event path changes)

### Step 3: Refresh schemas and user guidance

- [ ] Update tool descriptions/input schema wording if callers need to supply or understand sport-aware suffix behavior
- [ ] Regenerate affected schema snapshots if input schema descriptions change: `go run ./scripts/snapshot_tool_schemas.go`
- [ ] Update end-user workout docs if behavior or recommendations changed
- [ ] Add a CHANGELOG `[Unreleased]` entry if user-visible

**Artifacts:**
- `internal/tools/schema_snapshot/*.json` (modified if schemas change)
- `web/content/cookbook/build-workouts.md` (modified if affected)
- `web/content/reference/resources-prompts.md` (modified if affected)
- `CHANGELOG.md` (modified if user-visible)

### Step 4: Testing & Verification

> ZERO test failures allowed. This step runs the FULL test suite as a quality gate.

- [ ] Run FULL test suite: `make test`
- [ ] Run lint: `make lint`
- [ ] Fix all failures
- [ ] Build passes: `make build`

### Step 5: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md
- [ ] Summarize the clean-room behavior source as public behavior, not copied implementation

## Documentation Requirements

**Must Update:**
- `CHANGELOG.md` — add an `[Unreleased]` note if DSL output changes for users.

**Check If Affected:**
- `web/content/cookbook/build-workouts.md` — update planned-workout guidance if behavior changes.
- `web/content/reference/resources-prompts.md` — update MCP resource/prompt guidance if needed.
- `docs/prd/PRD-icuvisor.md` — only update if product contract changes; otherwise leave untouched.

## Completion Criteria

- [ ] Planned-workout DSL suffix selection respects athlete sport metric priority where needed
- [ ] Existing bare serializer behavior remains stable where no sport context is available
- [ ] Regression tests cover the running-power case
- [ ] All tests passing
- [ ] Documentation updated or explicitly marked unaffected

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-150): complete Step N — description`
- **Bug fixes:** `fix(TP-150): description`
- **Tests:** `test(TP-150): description`
- **Hydration:** `hydrate: TP-150 expand Step N checkboxes`

## Do NOT

- Copy or transliterate competitor source. Use only the public behavior signal summarized in this prompt and upstream Intervals behavior.
- Break existing WorkoutDoc golden fixtures unless the change is intentional and explained.
- Introduce a model-controlled `confirm` or safety override.
- Skip tests.
- Modify framework/standards docs without explicit user approval.
- Load docs not listed in "Context to Read First".
- Commit without the task ID prefix in the commit message.

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if issues discovered during execution.
     Format:
     ### Amendment N — YYYY-MM-DD HH:MM
     **Issue:** [what was wrong]
     **Resolution:** [what was changed] -->
