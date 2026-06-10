# Task: TP-157 - get_today current-day freshness regression

**Created:** 2026-06-09
**Size:** M

## Review Level: 1 (Plan Only)

**Assessment:** Adds focused freshness invariants to an existing read-only digest. The behavior is user-visible but the work should remain within `get_today` tests and date filtering.
**Score:** 3/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-157-get-today-freshness-regression/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Harden `get_today` so it never composes previous-day workout/date rows with today’s `_meta` when morning wellness is partial or absent. Public IntervalCoach feedback described a “half-old, half-new” daily briefing; icuvisor already has `_meta.as_of` and wellness stale metadata, but needs an explicit regression that all sections are bounded to the athlete-local current day.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — repo rules.
- `docs/prd/PRD-icuvisor.md` — daily digest/read-tool behavior if contract docs need updates.

## Environment

- **Workspace:** Go module root
- **Services required:** None

## File Scope

- `internal/tools/get_today.go`
- `internal/tools/get_today_test.go`
- `internal/tools/get_wellness_data.go`
- `internal/tools/get_events.go`
- `internal/tools/get_activities.go`
- `internal/response/*`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied

### Step 1: Add explicit freshness regression tests

- [ ] Add a `get_today` test with a fixed athlete timezone clock where `_meta.date`, `_meta.as_of_date`, fitness, events, and activities are all for today.
- [ ] Include previous-day fitness/event/activity rows in the fake clients and assert they are excluded, not silently mixed into today’s digest.
- [ ] Add cases where today’s wellness is absent or partial while previous-day wellness exists; assert the response does not backfill yesterday as today.
- [ ] Run targeted tests: `go test ./internal/tools -run TestGetToday`

**Artifacts:**
- `internal/tools/get_today_test.go` (modified)

### Step 2: Fix filtering or shaping if tests expose stale composition

- [ ] If tests fail, update `internal/tools/get_today.go` to filter shaped sections by the same athlete-local `today` used for `_meta`.
- [ ] Preserve existing wellness stale metadata semantics; do not hide stale warnings if they are intentionally emitted by wellness shaping.
- [ ] Ensure activities/events/fetch params remain server-side date-bounded and that extra defensive filtering cannot leak previous-day rows.
- [ ] Run targeted tests: `go test ./internal/tools -run TestGetToday`

**Artifacts:**
- `internal/tools/get_today.go` (modified if needed)
- `internal/tools/get_wellness_data.go` (modified only if stale metadata requires a shared helper)
- `internal/tools/get_events.go` (modified only if event date helpers are reused)
- `internal/tools/get_activities.go` (modified only if activity date helpers are reused)

### Step 3: Testing & Verification

- [ ] Run FULL test suite: `make test`
- [ ] Run integration tests (if applicable)
- [ ] Fix all failures
- [ ] Build passes: `make build`

### Step 4: Documentation & Delivery

- [ ] `CHANGELOG.md` notes the get_today freshness regression/fix.
- [ ] README/PRD reviewed if any user-visible stale metadata wording changes.
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**
- `CHANGELOG.md` — note freshness hardening.

**Check If Affected:**
- `README.md` — update only if get_today output semantics are documented there.
- `docs/prd/PRD-icuvisor.md` — update only if the public daily digest contract changes.

## Completion Criteria

- [ ] Tests prove previous-day rows are never paired with today’s `_meta`.
- [ ] Partial/absent morning wellness does not trigger stale row backfill.
- [ ] Full tests and build pass.

## Git Commit Convention

Commits happen at step boundaries. All commits for this task MUST include the task ID:

- **Step completion:** `feat(TP-157): complete Step N — description`
- **Bug fixes:** `fix(TP-157): description`
- **Tests:** `test(TP-157): description`
- **Hydration:** `hydrate: TP-157 expand Step N checkboxes`

## Do NOT

- Copy competitor source; use only the public bug signal.
- Remove existing `_meta.as_of` or wellness stale metadata.
- Backfill yesterday’s rows into today’s digest.
- Skip tests.
- Commit without the task ID prefix.

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if issues discovered during execution. -->
