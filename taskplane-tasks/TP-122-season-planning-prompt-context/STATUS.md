# TP-122: Season planning prompt and context hardening — Status

**Current Step:** Step 3: Strengthen race-event write examples if needed
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-29
**Review Level:** 2
**Review Counter:** 5
**Iteration:** 1
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it — aim for 2-5 outcome-level items per step, not exhaustive implementation scripts.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Clean-room guardrail confirmed

---

### Step 1: Design the safe season-planning guidance surface
**Status:** ✅ Complete

- [x] Existing planning/taper/review prompts inspected
- [x] Enhance-existing vs new-prompt approach decided
- [x] Deterministic existing tools selected; ATP writer excluded
- [x] Approach and non-goals recorded in Discoveries
- [x] Prompt tests run

---

### Step 2: Implement prompt and golden-test updates
**Status:** ✅ Complete

- [x] Prompt text updated or new prompt added
- [x] Guardrails added against automatic calendar filling/ATP creation
- [x] Golden fixtures/tests updated
- [x] Prompt tests run

---

### Step 3: Strengthen race-event write examples if needed
**Status:** 🟨 In Progress

- [ ] Race-event input examples reviewed
- [ ] Example/test strengthened if needed
- [ ] Write behavior preserved unless bug found
- [ ] Targeted tool tests run

---

### Step 4: Testing & Verification
**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Lint passing
- [ ] All failures fixed
- [ ] Build passes

---

### Step 5: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] `CHANGELOG.md` updated
- [ ] `ROADMAP.md` checked if affected
- [ ] PRD checked if affected
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Step 1 design: enhance existing prompts instead of adding a seventh `season_planning` prompt; strengthen `weekly_planning` as the season/race planning entry point, with lighter race-priority/compliance context in `race_week_taper` and `weekly_review`. | Avoids PRD prompt-catalog churn and keeps the guidance surface within the existing registry/golden-test pattern. | internal/prompts/catalog.go |
| Planning guidance will cite existing deterministic reads/analyzers only: `get_events`, `get_training_plan`, `get_fitness`, `get_training_summary`, `compute_compliance_rate`, and `icuvisor_list_advanced_capabilities`; `get_training_plan` and `compute_compliance_rate` may be unavailable in core toolsets, so prompts must tell assistants to list capabilities and proceed from events/fitness/summary when absent. | Do not add or imply an ATP/season calendar writer; do not automatically fill the calendar, create ATP notes, or call write/delete tools before the user approves exact changes. | internal/prompts/catalog.go |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-29 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-29 13:19 | Task started | Runtime V2 lane-runner execution |
| 2026-05-29 13:19 | Step 0 started | Preflight |
| 2026-05-29 13:20 | Step 0 completed | Required files, dependencies, and clean-room guardrail confirmed |
| 2026-05-29 13:20 | Step 1 started | Design prompt guidance surface |
| 2026-05-29 13:27 | Step 1 completed | Plan and code reviews approved |
| 2026-05-29 13:27 | Step 2 started | Prompt and golden-test updates |
| 2026-05-29 13:33 | Step 2 completed | Code review approved |
| 2026-05-29 13:33 | Step 3 started | Race-event write example hardening |

---

## Blockers

*None*

---

## Notes

*Reserved for execution notes*
| 2026-05-29 13:22 | Review R001 | plan Step 1: REVISE |
| 2026-05-29 13:25 | Review R002 | plan Step 1: APPROVE |
| 2026-05-29 13:27 | Review R003 | code Step 1: APPROVE |
| 2026-05-29 13:29 | Review R004 | plan Step 2: APPROVE |
| 2026-05-29 13:34 | Review R005 | code Step 2: APPROVE |
