# Task: TP-107 - Unit semantics audit

**Created:** 2026-05-26
**Size:** M

## Review Level: 2

**Assessment:** Cross-cutting unit and response-semantics regression coverage across workout docs, read tools, and response metadata; most changes should be additive but mistakes can mislead users.
**Score:** 4/8 — Blast radius: 2, Pattern novelty: 1, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-107-unit-semantics-audit/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Audit and harden unit semantics around workout targets, work/energy, calories, and hydration. The goal is regression coverage first: prove existing behavior is correct where it is correct, and make additive fixes where field names or metadata could cause assistants to misread units.

Tracking issue: https://github.com/ricardocabral/icuvisor/issues/34

## Dependencies

- **Task:** TP-007 (response shaping primitives exist)
- **Task:** TP-019 (workout_doc serializer exists)
- **Task:** TP-081 (nutrition and calorie labels exist)
- **Task:** TP-092 (activity histogram tool exists)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — repository rules and no-network test constraints.
- `docs/prd/PRD-icuvisor.md` — unit and response-shaping product rules.
- `internal/workoutdoc/syntax.go` — supported workout target units.
- `internal/workoutdoc/serialize.go` — workout target serialization.
- `internal/units/unit.go` — unit enum parsing.
- `internal/response/units.go` — preferred-unit metadata and conversion.
- `internal/tools/get_wellness_data.go` — wellness nutrition/hydration semantics.
- `internal/tools/get_activity_details.go` — activity calories and interval units.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None. Unit tests must not hit the network.

## File Scope

- `internal/workoutdoc/*`
- `internal/units/*`
- `internal/response/*`
- `internal/tools/get_activity_details.go`
- `internal/tools/get_activity_histogram.go`
- `internal/tools/get_wellness_data.go`
- `internal/tools/get_workout_library*.go`
- `internal/tools/*unit*_test.go`
- `internal/tools/*wellness*_test.go`
- `internal/tools/*activity*_test.go`
- `internal/intervals/testdata/**/*`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Current unit/metadata behavior scoped before changing code

### Step 1: Add workout target unit regression coverage

- [ ] Add tests for percent FTP / power target serialization.
- [ ] Add tests for pace target ranges and supported pace units.
- [ ] Add tests for heart-rate percent target variants where the DSL distinguishes them.
- [ ] Fix serializer behavior only if tests reveal a mismatch with icuvisor's documented contract.
- [ ] Run targeted workoutdoc tests.

**Artifacts:**
- `internal/workoutdoc/*_test.go` (modified/new)
- `internal/workoutdoc/serialize.go` or `syntax.go` (modified only if needed)

### Step 2: Add work/energy and unknown-unit regression coverage

- [ ] Audit joules/kilojoules handling in activity, interval, workout-library, histogram, and custom-item surfaces where relevant.
- [ ] Add tests proving raw joules are not mislabeled as kilojoules.
- [ ] Add tests proving unknown upstream units remain preserved as `unknown_unit` or equivalent.
- [ ] Apply additive metadata/label fixes if current behavior is ambiguous.
- [ ] Run targeted unit/response tests.

**Artifacts:**
- `internal/units/*` (modified/new tests)
- `internal/response/*` (modified/new tests)
- Relevant tool tests/fixtures (modified/new)

### Step 3: Add calories and hydration semantics coverage

- [ ] Add or extend tests that keep activity `calories_burned` distinct from wellness `calories_intake`.
- [ ] Clarify `hydration` versus `hydrationVolume` in wellness output via tests and additive metadata if needed.
- [ ] Ensure terse responses stay terse and use `_meta.field_semantics` or `_meta.units` for explanatory labels where possible.
- [ ] Run targeted wellness/activity tests.

**Artifacts:**
- `internal/tools/get_wellness_data.go` (modified only if needed)
- `internal/tools/get_activity_details.go` (modified only if needed)
- Tool tests/fixtures (modified/new)

### Step 4: Changelog and full verification

- [ ] Update `CHANGELOG.md` if behavior or metadata changes.
- [ ] Document discoveries for unit surfaces that are already correct and now covered by tests.
- [ ] Run targeted tests for all affected packages.

**Artifacts:**
- `CHANGELOG.md` (modified if needed)
- `STATUS.md` discoveries (updated)

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
- `CHANGELOG.md` — record user-visible behavior/metadata fixes if any.
- `STATUS.md` — keep execution state current and log unit-surface discoveries.

**Check If Affected:**
- `docs/prd/PRD-icuvisor.md` — update only if product unit semantics are clarified.
- Generated tool-reference data/docs — regenerate if tool metadata/schema descriptions change.
- `README.md` — update only if public examples change.

## Completion Criteria

- Regression tests cover workout target unit serialization, work/energy labeling, calories semantics, hydration semantics, and unknown units.
- Any discovered ambiguity is fixed additively where possible.
- No public field names are removed or renamed without explicit justification.
- `make test`, `make build`, and `make lint` pass or pre-existing failures are documented.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-107` for traceability. Examples:

- `test(TP-107): complete step 1 — cover workout target units`
- `fix(TP-107): clarify hydration volume metadata`
- `hydrate: TP-107 expand step checkboxes`

## Do NOT

- Do not change public field names if additive metadata can fix ambiguity.
- Do not invent units for upstream fields whose semantics are unknown.
- Do not broaden into unit conversion features outside the tracked scope.
- Do not hit the network in tests.
- Do not modify protected docs without explicit approval.

---

## Amendments

_Add amendments below this line only._
