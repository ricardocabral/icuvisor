# TP-145: Read-only planning context tool — Status

**Current Step:** Step 1: Design read-only planning context contract
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-03
**Review Level:** 1
**Review Counter:** 1
**Iteration:** 1
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Clean-room constraint confirmed

---

### Step 1: Design read-only planning context contract
**Status:** 🟨 In Progress

- [x] Inventory existing tool/client patterns for get_today, get_training_plan, get_events, get_fitness, and prompt planning guidance
- [x] Define a terse default response with `_meta.source_tools`, timezone/as-of, week window, and no write behavior
- [ ] R001 define deterministic week anchor/window, exact fitness and race windows, and event classification rules
- [ ] R001 justify toolset tier placement and stable caveat/validation contract
- [ ] Plan-review checkpoint completed before implementation

---

### Step 2: Implement get_planning_context
**Status:** ⬜ Not Started

- [ ] Add the tool using existing intervals client methods and response shaping patterns
- [ ] Return week events/workouts, active training-plan summary, current/recent fitness context, upcoming race context, and caveats without creating/updating/deleting calendar items
- [ ] Register the tool in the catalog/toolcatalog with appropriate core/full tier placement
- [ ] Add input and output schema descriptions that clearly distinguish planning context from ATP creation

---

### Step 3: Add tests and docs
**Status:** ⬜ Not Started

- [ ] Add table-driven handler tests for terse default, include_full behavior, source_tools metadata, timezone/week window handling, and empty-data caveats
- [ ] Add catalog/registration tests if needed
- [ ] Update CHANGELOG and README/catalog docs if user-visible
- [ ] Run targeted tests: `go test ./internal/tools ./internal/toolcatalog`

---

### Step 99: Testing & Verification
**Status:** ⬜ Not Started

- [ ] Targeted tests passing
- [ ] FULL test suite passing
- [ ] Build passes if code changed
- [ ] All failures fixed

---

### Step 100: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] Must-update docs modified
- [ ] Check-if-affected docs reviewed
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | Step 1 | REVISE | .reviews/R001-plan-step1.md |

---

## Discoveries

<!-- Workers log durable discoveries here. -->

| 2026-06-03 16:11 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 16:11 | Step 0 started | Preflight |
| 2026-06-03 16:11 | Preflight findings | Required task files, repo context, go.mod, and package dependencies are present; clean-room constraint confirmed from public behavior signal only. |
| 2026-06-03 16:12 | Step 1 design inventory | get_today composes profile timezone/as-of plus fitness/wellness/activities/events and records `_meta.source_tools`; get_training_plan is full-tier and returns assignment + lightweight plan summary with raw nested payloads only under include_full; get_events uses athlete-local bounded windows, category/limit/resolve, eventRow, sorting, and as-of metadata only when the requested range includes today; get_fitness fetches athlete summary rows and shapes CTL/ATL/TSB; weekly_planning/plan_health/race_taper prompts require reading events, training plan, fitness/summary/activity context, ask/confirm anchors, and explicitly avoid auto-filling calendars or ATP creation. |
| 2026-06-03 16:13 | Step 1 response contract | get_planning_context should be core-tier, read-only, no required args, optional `week_start` (athlete-local Monday or any date normalized to that week's Monday), optional `include_full`. Default response: `week` with start/end/as_of/timezone; `week_events` terse event rows split into planned workouts vs races/other events; `training_plan` as the existing assignment summary/unavailable shape; `fitness_context` with current row plus recent 7-day rows or a compact summary; `upcoming_races` from the week plus near-future window; `caveats` for no active plan, no week events/workouts, no fitness rows, and read-only/no ATP behavior. `_meta` must include `source_tools: [get_events,get_training_plan,get_fitness]`, `include_full`, section counts, week window, as_of metadata, and `read_only: true` / `writes_performed: false`. |
| 2026-06-03 16:14 | R001 plan review | Reviewer requested exact default week anchor, tier rationale, exact fitness/race windows, event classification rules, caveat codes, and week_start validation before implementation. |
| 2026-06-03 16:14 | Review R001 | plan Step 1: UNKNOWN |
