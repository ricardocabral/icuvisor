# Task: TP-155 - Activity interval manual and mixed classification

**Created:** 2026-06-09
**Size:** M

## Review Level: 1 (Plan Only)

**Assessment:** Extends an existing classifier and response metadata with two new source states. The blast radius is limited to interval-source analysis and callers, with no auth or persistence changes.
**Score:** 3/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-155-activity-interval-manual-mixed-classification/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Teach `get_activity_intervals` interval-source evidence to distinguish manually added intervals and mixed interval sets, not just `structured_workout`, `device_laps`, and `unknown`. Public forum behavior notes indicate intervals.icu auto-detected intervals carry `group_id` while manually added intervals do not; icuvisor should capture that useful distinction without copying any competitor source.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — clean-room and project coding rules.
- `docs/prd/PRD-icuvisor.md` — response/catalog behavior if public enum metadata changes.
- `ROADMAP.md` — phasing context if this affects analyzer milestones.

## Environment

- **Workspace:** Go module root
- **Services required:** None

## File Scope

- `internal/analysis/interval_source.go`
- `internal/analysis/interval_source_test.go`
- `internal/analysis/meta.go`
- `internal/analysis/meta_test.go`
- `internal/tools/analyzer_common_test.go`
- `internal/tools/get_activity_details.go`
- `internal/tools/get_activity_details_test.go`
- `internal/tools/schema_snapshot/get_activity_intervals.json`
- `README.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied

### Step 1: Add classifier states and fixture coverage

- [ ] Add `manual_added` and `mixed` interval-source states in `internal/analysis/interval_source.go` with comments and stable JSON string values.
- [ ] Add tests in `internal/analysis/interval_source_test.go` proving all intervals without upstream group markers classify as manual, all with group markers preserve existing structured/device behavior as appropriate, and a combination classifies as mixed.
- [ ] Preserve current structured-workout and device-lap precedence where explicit markers are stronger than the group-id heuristic.
- [ ] Run targeted tests: `go test ./internal/analysis`

**Artifacts:**
- `internal/analysis/interval_source.go` (modified)
- `internal/analysis/interval_source_test.go` (modified)

### Step 2: Propagate source evidence to tool/analyzer responses

- [ ] Update `internal/analysis/meta.go` and tests so analyzer `_meta.interval_source` can emit the new values without losing `auto_lap_suspected` behavior.
- [ ] Update get-activity-interval response shaping/tests so user-visible output distinguishes `manual_added` and `mixed` where interval rows expose the relevant raw fields.
- [ ] Refresh `internal/tools/schema_snapshot/get_activity_intervals.json` if schema descriptions/enums change.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/analysis`

**Artifacts:**
- `internal/analysis/meta.go` (modified)
- `internal/analysis/meta_test.go` (modified)
- `internal/tools/get_activity_details.go` (modified if needed)
- `internal/tools/get_activity_details_test.go` (modified)
- `internal/tools/schema_snapshot/get_activity_intervals.json` (modified if needed)

### Step 3: Testing & Verification

- [ ] Run FULL test suite: `make test`
- [ ] Run integration tests (if applicable)
- [ ] Fix all failures
- [ ] Build passes: `make build`

### Step 4: Documentation & Delivery

- [ ] `CHANGELOG.md` notes the clearer interval-source classification.
- [ ] `README.md` tool/output description reviewed and updated if it names the interval-source enum.
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**
- `CHANGELOG.md` — note the behavior change.

**Check If Affected:**
- `README.md` — update any interval-source enum/output docs.
- `docs/prd/PRD-icuvisor.md` — only update if public contract text needs to include the new states.

## Completion Criteria

- [ ] Manual, mixed, structured, device-lap, and unknown interval-source cases are tested.
- [ ] Public responses/meta can represent the new source states.
- [ ] Full tests and build pass.

## Git Commit Convention

Commits happen at step boundaries. All commits for this task MUST include the task ID:

- **Step completion:** `feat(TP-155): complete Step N — description`
- **Bug fixes:** `fix(TP-155): description`
- **Tests:** `test(TP-155): description`
- **Hydration:** `hydrate: TP-155 expand Step N checkboxes`

## Do NOT

- Read or copy GPL/copyleft competitor source; use only the public forum behavior signal and upstream API behavior.
- Rename existing interval-source values.
- Treat missing raw fields as manual if the upstream fixture does not expose enough evidence.
- Skip tests.
- Load docs not listed above.
- Commit without the task ID prefix.

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if issues discovered during execution. -->
