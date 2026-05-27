# TP-109: Description-only workout safety warning — Status

**Current Step:** Step 3: Testing & Verification
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-27
**Review Level:** 1
**Review Counter:** 2
**Iteration:** 1
**Size:** S

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Current add/update event and workout update behavior reviewed

---

### Step 1: Design the warning/guard contract
**Status:** ✅ Complete

> **Plan-review checkpoint**

- [x] Minimal additive response metadata field(s) chosen
- [x] Warning text is terse, actionable, and non-blocking for legitimate strength prose updates
- [x] Trigger condition documented

---

### Step 2: Implement warning behavior and tests
**Status:** ✅ Complete

- [x] Warning behavior implemented for description-only `WORKOUT` event updates
- [x] `update_workout` checked and covered if affected
- [x] Warning-present and warning-absent tests added
- [x] Targeted tests passing: `go test ./internal/tools`

---

### Step 3: Testing & Verification
**Status:** 🟨 In Progress

- [ ] FULL test suite passing: `make test`
- [ ] Lint passing or documented: `make lint`
- [ ] Build passes: `make build`
- [ ] All failures fixed or documented as pre-existing unrelated failures

---

### Step 4: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] `CHANGELOG.md` updated
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged
- [ ] Step-boundary commit includes `TP-109`

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | plan | 1 | APPROVE | `.reviews/R001-plan-step1.md` |
| R002 | plan | 2 | APPROVE | `.reviews/R002-plan-step2.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-27 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-27 20:40 | Task started | Runtime V2 lane-runner execution |
| 2026-05-27 20:40 | Step 0 started | Preflight |

---

## Blockers

*None*

---

## Notes

- Step 1 design: add a new optional `_meta.description_only_workout_warning` string to write responses instead of overloading existing `_meta.workout_doc_warning`, which is reserved for upstream render/parse failures after uploading WorkoutDoc.
- Step 1 warning text: use terse non-blocking copy: `Description was written without workout_doc; if this item previously had structured steps, they may have been replaced. Include workout_doc when preserving or merging workout structure.`
- Step 1 trigger: `add_or_update_event` sets the warning only on update-shaped writes with `event_id` present, `description` supplied, no `workout_doc`, and `category` equals `WORKOUT` case-insensitively. `update_workout` is affected because it sparsely updates an existing template by ID and a supplied `description` replaces the upstream description/DSL, so it sets the same warning when `description` is supplied and no `workout_doc` is supplied.
| 2026-05-27 20:44 | Review R001 | plan Step 1: APPROVE |
| 2026-05-27 20:46 | Review R002 | plan Step 2: APPROVE |
