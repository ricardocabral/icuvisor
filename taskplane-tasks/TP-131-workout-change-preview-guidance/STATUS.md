# TP-131: Workout change preview guidance — Status

**Current Step:** Step 4: Testing & Verification
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-29
**Review Level:** 2
**Review Counter:** 10
**Iteration:** 2
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
**Status:** ✅ Complete

- [x] Inspect `validate_workout`, workout write tool descriptions/examples, and weekly-planning/build-workouts prompts.
- [x] Identify whether assistants are instructed to summarize proposed changes before writes.
- [x] Record current behavior and chosen changes in STATUS.md Discoveries.
- [x] Run targeted tests: `go test ./internal/tools ./internal/prompts`

---

### Step 2: Harden preview guidance
**Status:** ✅ Complete

- [x] Update prompts/tool examples so proposed changes include total duration, key steps, target intensities, load/distance/time changes, and what is being preserved.
- [x] Recommend `validate_workout` preflight for uncertain DSL or structured workout changes.
- [x] Do not introduce a model-controlled `confirm` override or bypass safety modes.
- [x] Run targeted tests: `go test ./internal/tools ./internal/prompts`
- [x] Regenerate/update generated tool catalog artifacts after `validate_workout` summary changes (`web/data/tools.json`, `cmd/gendocs/testdata/tools.golden.json`).

---

### Step 3: Update cookbook examples
**Status:** ✅ Complete

- [x] Update build-workouts cookbook with before/after preview language and approval workflow.
- [x] Ensure examples distinguish prose description from structured `workout_doc`.
- [x] Run targeted tests/docs validation as available.
- [x] Add a concrete existing-template before/after update preview showing duration, key intervals, intensity targets, load/distance/time deltas, and preserved fields before approval.

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
| R003 | code | 1 | APPROVE | inline review_step response |
| R004 | plan | 2 | APPROVE | `.reviews/R004-plan-step2.md` |
| R005 | code | 2 | REVISE | `.reviews/R005-code-step2.md` |
| R006 | code | 2 | APPROVE | inline review_step response |
| R007 | plan | 3 | APPROVE | `.reviews/R007-plan-step3.md` |
| R008 | code | 3 | REVISE | `.reviews/R008-code-step3.md` |
| R009 | code | 3 | UNAVAILABLE | inline review_step response |
| R010 | code | 3 | APPROVE | `.reviews/R010-code-step3.md` |

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
| 2026-05-29 14:40 | Worker iter 1 | done in 1312s, tools: 52 |

---

## Blockers

*None*

---

## Notes

*Reserved for execution notes*
| 2026-05-29 14:23 | Review R001 | plan Step 1: APPROVE |
| 2026-05-29 14:26 | Review R002 | code Step 1: UNKNOWN |
| 2026-05-29 14:28 | Review R003 | code Step 1: APPROVE |
| 2026-05-29 14:31 | Review R004 | plan Step 2: APPROVE |
| 2026-05-29 14:47 | Review R005 | code Step 2: REVISE |
| 2026-05-29 14:51 | Review R006 | code Step 2: APPROVE |
| 2026-05-29 14:53 | Review R007 | plan Step 3: APPROVE |
| 2026-05-29 14:56 | Review R008 | code Step 3: REVISE |
| 2026-05-29 15:15 | Review R010 | code Step 3: APPROVE |
