# Task: TP-159 - Gear name resolution regression

**Created:** 2026-06-09
**Size:** S

## Review Level: 1 (Plan Only)

**Assessment:** Adds/strengthens regression coverage around existing gear resolution behavior. The implementation should be test-first and narrow.
**Score:** 2/8 — Blast radius: 1, Pattern novelty: 0, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-159-gear-name-resolution-regression/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Reconfirm that activity gear IDs are resolved to human-readable names by fetching the full gear list when an activity exposes only a numeric `gear_id`. Recent public IcuSync feedback showed this is an important upstream behavior; icuvisor already claims gear resolution and has tests, but should have a fixture that directly covers “activity has gear ID, no name, gear list resolves it.”

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — repo rules and clean-room constraints.
- `README.md` — existing gear-resolution claim.

## Environment

- **Workspace:** Go module root
- **Services required:** None

## File Scope

- `internal/tools/activity_gear_resolution.go`
- `internal/tools/activity_gear_resolution_test.go`
- `internal/tools/get_activity_details_test.go`
- `internal/tools/get_activities_test.go`
- `internal/intervals/activity_gear_test.go`
- `internal/intervals/testdata/activity_detail_with_gear.json`
- `internal/intervals/testdata/activity_list_with_gear.json`
- `internal/intervals/testdata/gear_list.json`
- `README.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied

### Step 1: Add direct numeric-gear-id regression coverage

- [ ] Add or update fixtures so an activity has `gear_id` but no embedded gear name.
- [ ] Add tests proving the resolver fetches/list-uses `gear_list.json` and emits the resolved gear name in activity details/list output.
- [ ] Add a fallback assertion for unknown gear IDs so unresolved names stay absent/marked without misleading the LLM.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/intervals -run 'Gear|Activity.*Gear|GetActivit'`

**Artifacts:**
- `internal/tools/activity_gear_resolution_test.go` (modified)
- `internal/tools/get_activity_details_test.go` (modified if needed)
- `internal/tools/get_activities_test.go` (modified if needed)
- `internal/intervals/testdata/*gear*.json` (modified if needed)

### Step 2: Fix resolver behavior only if the regression fails

- [ ] If the test fails, update `internal/tools/activity_gear_resolution.go` to resolve by numeric/string ID consistently using the gear list.
- [ ] Preserve existing cache behavior and error handling so gear lookup failures do not break activity reads unnecessarily.
- [ ] Ensure README wording matches actual output fields.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/intervals -run 'Gear|Activity.*Gear|GetActivit'`

**Artifacts:**
- `internal/tools/activity_gear_resolution.go` (modified if needed)
- `README.md` (modified if needed)

### Step 3: Testing & Verification

- [ ] Run FULL test suite: `make test`
- [ ] Run integration tests (if applicable)
- [ ] Fix all failures
- [ ] Build passes: `make build`

### Step 4: Documentation & Delivery

- [ ] `CHANGELOG.md` notes the gear resolution regression coverage/fix.
- [ ] `README.md` reviewed for gear-resolution claim accuracy.
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**
- `CHANGELOG.md` — note regression coverage/fix.

**Check If Affected:**
- `README.md` — ensure gear resolution claim remains accurate.

## Completion Criteria

- [ ] Test fixture covers ID-only activity gear resolved via full gear list.
- [ ] Unknown gear fallback remains safe and truthful.
- [ ] Full tests and build pass.

## Git Commit Convention

Commits happen at step boundaries. All commits for this task MUST include the task ID:

- **Step completion:** `feat(TP-159): complete Step N — description`
- **Bug fixes:** `fix(TP-159): description`
- **Tests:** `test(TP-159): description`
- **Hydration:** `hydrate: TP-159 expand Step N checkboxes`

## Do NOT

- Copy competitor code; use only the public feedback signal and upstream fixtures.
- Introduce network calls in tests.
- Make activity reads fail solely because gear lookup is unavailable.
- Commit without the task ID prefix.

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if issues discovered during execution. -->
