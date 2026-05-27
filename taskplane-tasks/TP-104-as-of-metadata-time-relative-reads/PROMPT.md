# Task: TP-104 - As-of metadata for time-relative reads

**Created:** 2026-05-26
**Size:** M

## Review Level: 2

**Assessment:** Additive metadata across several read tools with date/time edge cases; existing shaping patterns apply, but timezone correctness needs code review.
**Score:** 4/8 — Blast radius: 2, Pattern novelty: 1, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-104-as-of-metadata-time-relative-reads/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Add athlete-local "as of" anchors to reads whose interpretation depends on the current day. Assistants should receive the server-computed local datetime, date, weekday, and timezone so they do not infer local calendar context from UTC or stale chat context.

Tracking issue: https://github.com/ricardocabral/icuvisor/issues/31

## Dependencies

- **Task:** TP-007 (response shaping primitives exist)
- **Task:** TP-009 (activity reads exist)
- **Task:** TP-011 (wellness reads exist)
- **Task:** TP-012 (event reads exist)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — repository rules and Go/MCP conventions.
- `ROADMAP.md` — roadmap item for `_meta.as_of` on time-relative reads.
- `docs/prd/PRD-icuvisor.md` — timezone and response-shaping product contract.
- `internal/tools/get_today.go` — existing athlete-local date helper and injectable clock.
- `internal/tools/get_activities.go` — activities response metadata and pagination.
- `internal/tools/get_events.go` — event response metadata.
- `internal/tools/get_wellness_data.go` — wellness row/response metadata.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None. Unit tests must not hit the network.

## File Scope

- `internal/tools/get_today.go`
- `internal/tools/get_today_test.go`
- `internal/tools/get_activities.go`
- `internal/tools/get_activities*_test.go`
- `internal/tools/get_events.go`
- `internal/tools/get_events*_test.go`
- `internal/tools/get_wellness_data.go`
- `internal/tools/get_wellness_data_test.go`
- `internal/response/*`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm current metadata conventions before adding new keys

### Step 1: Design and implement shared athlete-local as-of helper

- [ ] Add or identify a helper that returns athlete-local RFC3339 datetime, date, weekday, and timezone from a clock and timezone name.
- [ ] Cover timezone edge cases with deterministic tests, including local date differing from UTC.
- [ ] Ensure failures follow existing timezone error behavior and do not silently change dates.
- [ ] Run targeted helper tests.

**Artifacts:**
- `internal/tools/*` or `internal/response/*` helper (modified/new)
- Helper tests (modified/new)

### Step 2: Add metadata to `get_today`

- [ ] Add `_meta.as_of`, `_meta.as_of_date`, `_meta.as_of_weekday`, and `_meta.timezone` to `get_today` output.
- [ ] Use the existing injectable clock for deterministic tests.
- [ ] Preserve existing `date`, `activity_window`, and section count semantics.
- [ ] Run targeted `get_today` tests.

**Artifacts:**
- `internal/tools/get_today.go` (modified)
- `internal/tools/get_today_test.go` (modified)

### Step 3: Add metadata to current-day range reads

- [ ] Add the same anchor metadata to `get_activities` when the requested range includes the athlete-local current date.
- [ ] Add the same anchor metadata to `get_events` when the requested range includes the athlete-local current date.
- [ ] Add the same anchor metadata to `get_wellness_data` when the requested range includes the athlete-local current date.
- [ ] Preserve existing pagination tokens, null stripping, and terse/full behavior.
- [ ] Run targeted tests for each tool.

**Artifacts:**
- `internal/tools/get_activities.go` (modified)
- `internal/tools/get_events.go` (modified)
- `internal/tools/get_wellness_data.go` (modified)
- Tool tests (modified/new)

### Step 4: Regression tests and changelog

- [ ] Cover positive-offset and negative-offset timezone examples.
- [ ] Cover date ranges that include and exclude the local current day.
- [ ] Confirm past-only ranges do not gain confusing current-day metadata unless deliberately documented.
- [ ] Update `CHANGELOG.md` under `[Unreleased]`.

**Artifacts:**
- Tool tests/fixtures (modified/new)
- `CHANGELOG.md` (modified)

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
- `CHANGELOG.md` — record additive metadata behavior under `[Unreleased]`.
- `STATUS.md` — keep execution state current.

**Check If Affected:**
- `ROADMAP.md` — mark/update roadmap item only if this fully satisfies it.
- `docs/prd/PRD-icuvisor.md` — update only if product contract changes.
- Generated tool-reference data/docs — regenerate if response metadata docs require it.

## Completion Criteria

- `get_today` always returns athlete-local as-of metadata.
- `get_activities`, `get_events`, and `get_wellness_data` return as-of metadata when their range includes today in the athlete timezone.
- Weekday is computed server-side from the same localized instant as the date.
- Tests cover local date differing from UTC.
- `make test`, `make build`, and `make lint` pass or pre-existing failures are documented.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-104` for traceability. Examples:

- `feat(TP-104): complete step 1 — add as-of helper`
- `test(TP-104): cover timezone boundary metadata`
- `hydrate: TP-104 expand step checkboxes`

## Do NOT

- Do not rename or remove existing response fields.
- Do not alter activity pagination token semantics.
- Do not silently fall back to UTC if existing behavior would return a timezone error.
- Do not broaden into a calendar/date utility API unless needed for this metadata.
- Do not modify protected docs without explicit approval.

---

## Amendments

_Add amendments below this line only._
