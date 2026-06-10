# Task: TP-156 - WorkoutDoc repeat plus trailing cooldown regression

**Created:** 2026-06-09
**Size:** S

## Review Level: 1 (Plan Only)

**Assessment:** Adds golden coverage around an existing parser/serializer edge case. The change is low risk but touches a compact DSL parser with user-visible workout behavior.
**Score:** 2/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-156-workoutdoc-repeat-trailing-cooldown-regression/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Add a WorkoutDoc regression proving that a repeat main set followed by a single trailing cooldown remains outside the repeat after parse/serialize round trips. A public Montis report described cooldown text being interpreted inside every repeat when workout descriptions were flattened into one string; icuvisor should lock in blank-line/group-boundary behavior.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — repo rules and clean-room constraints.
- `internal/workoutdoc/testdata/README.md` — fixture naming/meaning.

## Environment

- **Workspace:** Go module root
- **Services required:** None

## File Scope

- `internal/workoutdoc/workoutdoc_test.go`
- `internal/workoutdoc/parse.go`
- `internal/workoutdoc/serialize.go`
- `internal/workoutdoc/testdata/*repeat*cooldown*`
- `internal/workoutdoc/testdata/README.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied

### Step 1: Add golden fixture for repeat main set plus final cooldown

- [ ] Add a DSL fixture with a named warmup, a `Main Set 3x ...` repeat block, a blank/group boundary, and one final `Cooldown` after the repeat.
- [ ] Add the matching structured JSON fixture asserting the cooldown is a sibling after the repeat, not a child repeated three times.
- [ ] Document the fixture in `internal/workoutdoc/testdata/README.md`.
- [ ] Run targeted tests: `go test ./internal/workoutdoc`

**Artifacts:**
- `internal/workoutdoc/testdata/*repeat*cooldown*-dsl.txt` (new)
- `internal/workoutdoc/testdata/*repeat*cooldown*-structured.json` (new)
- `internal/workoutdoc/testdata/README.md` (modified)

### Step 2: Fix parser/serializer only if the new fixture fails

- [ ] If the regression fails, update `internal/workoutdoc/parse.go` so blank-line/group boundaries keep trailing sections outside repeat blocks.
- [ ] If serialization loses the boundary, update `internal/workoutdoc/serialize.go` to preserve enough grouping for the round trip.
- [ ] Add focused assertions to `internal/workoutdoc/workoutdoc_test.go` if the generic golden harness does not make the cooldown nesting failure obvious.
- [ ] Run targeted tests: `go test ./internal/workoutdoc`

**Artifacts:**
- `internal/workoutdoc/parse.go` (modified if needed)
- `internal/workoutdoc/serialize.go` (modified if needed)
- `internal/workoutdoc/workoutdoc_test.go` (modified if needed)

### Step 3: Testing & Verification

- [ ] Run FULL test suite: `make test`
- [ ] Run integration tests (if applicable)
- [ ] Fix all failures
- [ ] Build passes: `make build`

### Step 4: Documentation & Delivery

- [ ] `CHANGELOG.md` notes the WorkoutDoc regression coverage/fix.
- [ ] `internal/workoutdoc/testdata/README.md` remains accurate.
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**
- `CHANGELOG.md` — note added regression coverage or parser fix.
- `internal/workoutdoc/testdata/README.md` — describe the new fixture.

**Check If Affected:**
- `docs/prd/PRD-icuvisor.md` — only if behavior text changes.

## Completion Criteria

- [ ] A golden fixture proves trailing cooldown is not nested inside repeats.
- [ ] WorkoutDoc targeted tests, full tests, and build pass.

## Git Commit Convention

Commits happen at step boundaries. All commits for this task MUST include the task ID:

- **Step completion:** `feat(TP-156): complete Step N — description`
- **Bug fixes:** `fix(TP-156): description`
- **Tests:** `test(TP-156): description`
- **Hydration:** `hydrate: TP-156 expand Step N checkboxes`

## Do NOT

- Copy competitor implementation or test code; use the public behavior report only.
- Change unrelated WorkoutDoc syntax.
- Skip the full test suite.
- Load docs not listed above.
- Commit without the task ID prefix.

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if issues discovered during execution. -->
