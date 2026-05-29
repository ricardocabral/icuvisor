# TP-131: Workout change preview guidance — Status

**Current Step:** Step 1: Audit current pre-write guidance
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-29
**Review Level:** 2
**Review Counter:** 2
**Iteration:** 1
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no GPL/copyleft competitor source is opened or copied; use only public forum behavior signals and project docs.

---

### Step 1: Audit current pre-write guidance
**Status:** 🟨 In Progress

- [x] Inspect `validate_workout`, workout write tool descriptions/examples, and weekly-planning/build-workouts prompts.
- [x] Identify whether assistants are instructed to summarize proposed changes before writes.
- [x] Record current behavior and chosen changes in STATUS.md Discoveries.
- [x] Run targeted tests: `go test ./internal/tools ./internal/prompts`

---

### Step 2: Harden preview guidance
**Status:** ⬜ Not Started

- [ ] Update prompts/tool examples so proposed changes include total duration, key steps, target intensities, load/distance/time changes, and what is being preserved.
- [ ] Recommend `validate_workout` preflight for uncertain DSL or structured workout changes.
- [ ] Do not introduce a model-controlled `confirm` override or bypass safety modes.
- [ ] Run targeted tests: `go test ./internal/tools ./internal/prompts`

---

### Step 3: Update cookbook examples
**Status:** ⬜ Not Started

- [ ] Update build-workouts cookbook with before/after preview language and approval workflow.
- [ ] Ensure examples distinguish prose description from structured `workout_doc`.
- [ ] Run targeted tests/docs validation as available.

---

### Step 4: Testing & Verification
**Status:** ⬜ Not Started

- [ ] FULL test suite passing: `make test`
- [ ] Lint passes or pre-existing linter limitations are documented: `make lint`
- [ ] Build passes: `make build`
- [ ] All failures fixed or clearly documented as pre-existing

---

### Step 5: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | plan | 1 | APPROVE | `.reviews/R001-plan-step1.md` |
| R002 | code | 1 | REVISE | `.reviews/R002-code-step1.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Existing guidance requires a reviewed proposal/approval before writes in weekly planning and the cookbook, and write tools already warn that description-only workout updates can replace structured steps. | Harden preview content by explicitly requiring a human-readable change preview: total duration, key steps, target intensities, load/distance/time deltas, and preserved fields before create/update/schedule writes. | internal/prompts/testdata/weekly_planning.md; web/content/cookbook/build-workouts.md; internal/tools/add_or_update_event.go; internal/tools/create_workout.go; internal/tools/update_workout.go |
| `validate_workout` is read-only and reports canonical DSL plus estimated duration; current guidance recommends it only when syntax is uncertain. | Keep validate as preflight guidance for uncertain DSL/structured changes and use its output to support preview summaries, without making validation a write precondition or adding confirmation arguments. | internal/tools/validate_workout.go |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-29 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-29 14:18 | Task started | Runtime V2 lane-runner execution |
| 2026-05-29 14:18 | Step 0 started | Preflight |
| 2026-05-29 14:19 | Step 1 plan review | APPROVE; see `.reviews/R001-plan-step1.md` |
| 2026-05-29 14:20 | Step 1 targeted tests | `go test ./internal/tools ./internal/prompts` passed (`ok` for both packages, cached) |

---

## Blockers

*None*

---

## Notes

*Reserved for execution notes*
| 2026-05-29 14:23 | Review R001 | plan Step 1: APPROVE |
| 2026-05-29 14:26 | Review R002 | code Step 1: UNKNOWN |
