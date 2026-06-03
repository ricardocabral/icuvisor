# TP-145: Read-only planning context tool — Status

**Current Step:** Step 3: Add tests and docs
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-03
**Review Level:** 1
**Review Counter:** 6
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
**Status:** ✅ Complete

- [x] Inventory existing tool/client patterns for get_today, get_training_plan, get_events, get_fitness, and prompt planning guidance
- [x] Define a terse default response with `_meta.source_tools`, timezone/as-of, week window, and no write behavior
- [x] R001 define deterministic week anchor/window, exact fitness and race windows, and event classification rules
- [x] R001 justify toolset tier placement and stable caveat/validation contract
- [x] R002 define bounded current fitness window and future week_start behavior
- [x] R002 define event/race fetch limits, truncation metadata, caveat conditions, and include_full scope
- [x] Plan-review checkpoint completed before implementation

---

### Step 2: Implement get_planning_context
**Status:** ✅ Complete

- [x] R004 choose catalog group, tier-test update scope, and deterministic clock injection pattern
- [x] Add the tool using existing intervals client methods and response shaping patterns
- [x] Return week events/workouts, active training-plan summary, current/recent fitness context, upcoming race context, and caveats without creating/updating/deleting calendar items
- [x] Register the tool in the catalog/toolcatalog with appropriate core/full tier placement
- [x] Add input and output schema descriptions that clearly distinguish planning context from ATP creation

---

### Step 3: Add tests and docs
**Status:** 🟨 In Progress

- [x] R006 expand handler matrix for full contract, clock/window/call, classification, caveat, and truncation cases
- [ ] Add table-driven handler tests for terse default, include_full behavior, source_tools metadata, timezone/week window handling, and empty-data caveats
- [ ] Add catalog/registration tests for full tier, workout-library group, toolcatalog known name, and athlete ACL scope
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
| R002 | Plan | Step 1 | REVISE | .reviews/R002-plan-step1.md |
| R003 | Plan | Step 1 | APPROVE | inline |
| R004 | Plan | Step 2 | REVISE | .reviews/R004-plan-step2.md |
| R005 | Plan | Step 2 | APPROVE | inline |
| R006 | Plan | Step 3 | REVISE | .reviews/R006-plan-step3.md |

---

## Discoveries

<!-- Workers log durable discoveries here. -->

| 2026-06-03 16:11 | Task started | Runtime V2 lane-runner execution |
| 2026-06-03 16:11 | Step 0 started | Preflight |
| 2026-06-03 16:11 | Preflight findings | Required task files, repo context, go.mod, and package dependencies are present; clean-room constraint confirmed from public behavior signal only. |
| 2026-06-03 16:12 | Step 1 design inventory | get_today composes profile timezone/as-of plus fitness/wellness/activities/events and records `_meta.source_tools`; get_training_plan is full-tier and returns assignment + lightweight plan summary with raw nested payloads only under include_full; get_events uses athlete-local bounded windows, category/limit/resolve, eventRow, sorting, and as-of metadata only when the requested range includes today; get_fitness fetches athlete summary rows and shapes CTL/ATL/TSB; weekly_planning/plan_health/race_taper prompts require reading events, training plan, fitness/summary/activity context, ask/confirm anchors, and explicitly avoid auto-filling calendars or ATP creation. |
| 2026-06-03 16:13 | Step 1 response contract | get_planning_context should be core-tier, read-only, no required args, optional `week_start` (athlete-local Monday or any date normalized to that week's Monday), optional `include_full`. Default response: `week` with start/end/as_of/timezone; `week_events` terse event rows split into planned workouts vs races/other events; `training_plan` as the existing assignment summary/unavailable shape; `fitness_context` with current row plus recent 7-day rows or a compact summary; `upcoming_races` from the week plus near-future window; `caveats` for no active plan, no week events/workouts, no fitness rows, and read-only/no ATP behavior. `_meta` must include `source_tools: [get_events,get_training_plan,get_fitness]`, `include_full`, section counts, week window, as_of metadata, and `read_only: true` / `writes_performed: false`. |
| 2026-06-03 16:14 | R001 plan review | Reviewer requested exact default week anchor, tier rationale, exact fitness/race windows, event classification rules, caveat codes, and week_start validation before implementation. |
| 2026-06-03 16:15 | R001 deterministic windows/classification | Default window: upcoming athlete-local Monday-Sunday week computed from profile timezone and current clock; if today is Monday, use today's Monday as the start (not the following week). Supplied `week_start` is trimmed, must be YYYY-MM-DD, and is normalized backward to that date's Monday. `week.end_date = week.start_date + 6 days`. Fitness context fetches exactly `week_start - 7 days` through `as_of_date` inclusive and returns terse sorted rows plus `current` as the latest row on/before as_of_date. Upcoming races fetch from `as_of_date` through `as_of_date + 84 days` inclusive and include categories `RACE` and any category with `RACE_` prefix. Week event classification: category `WORKOUT` goes to `planned_workouts`; `RACE`/`RACE_*` goes to `races`; `NOTE` goes to `notes`; all other known/unknown categories go to `other_events` preserving upstream category. |
| 2026-06-03 16:16 | R001 tier/caveat contract | Tier decision supersedes the initial core-tier note: implement as full-tier because it composes `get_training_plan`, a current full-tier source, and should not widen core visibility of training-plan assignment summaries without a separate toolset policy decision. Stable caveat codes: `no_active_training_plan`, `no_week_workouts`, `no_week_events`, `no_upcoming_races`, `no_fitness_rows`, `read_only_no_atp`, and `partial_training_plan_summary` when nested plan summary is absent. `_meta.source_tools` includes `get_athlete_profile` for timezone/as-of context plus `get_events`, `get_training_plan`, and `get_fitness`; `_meta` also exposes `read_only: true`, `writes_performed: false`, `planning_scope: context_only`, `week_window`, `fitness_window`, `race_window`, counts, and `caveat_codes`. Invalid `week_start` returns a short user-facing error. |
| 2026-06-03 16:17 | R002 plan review | Reviewer requested a bounded current-fitness window independent of future/historical week_start drift, explicit week-event/race fetch limits, truncation metadata/caveats, exact caveat-code conditions, and include_full section scope. |
| 2026-06-03 16:18 | R002 plan review | Reviewer requested a bounded current-fitness window independent of future/historical week_start drift, explicit week-event/race fetch limits, truncation metadata/caveats, exact caveat-code conditions, and include_full section scope. |
| 2026-06-03 16:18 | R002 fitness window | Fitness context is current-load context, not plan-window history: always fetch exactly the 7 athlete-local dates from `as_of_date - 6 days` through `as_of_date` inclusive, independent of supplied or default `week_start`. `fitness_context.rows` is sorted by date; `fitness_context.current` is the latest fetched row on or before `as_of_date`; future and historical `week_start` values only affect `week` and week events, not the fitness window. Tests should include a future `week_start` case asserting the fitness client still receives the as-of current window. |
| 2026-06-03 16:19 | R002 event limits/caveats/full scope | Week events fetch `week_start..week_end` with `Limit: 500`; race scan fetches `as_of_date..as_of_date+84 days` with `Limit: 500` and filters `RACE`/`RACE_*`. Because upstream applies the limit, `len(events) >= 500` is treated as `*_may_be_truncated`, surfaced in `_meta.truncation` and caveat codes `week_events_may_be_truncated` / `upcoming_races_may_be_truncated`. Empty week emits `no_week_events` and, because workouts are empty, `no_week_workouts`; a non-empty week without `WORKOUT` emits only `no_week_workouts`. No race rows after filtering emits `no_upcoming_races` unless race scan may be truncated, in which case include both caveats. `include_full:true` only adds per-event `full`, fitness row `full`, and raw training-plan assignment/nested payloads; default omits raw upstream payloads in all sections. Implementation must compose read-only client methods directly: `GetAthleteProfile`, `ListEvents`, `GetTrainingPlan`, `ListAthleteSummary`; no create/update/delete methods. |
| 2026-06-03 16:14 | Review R001 | plan Step 1: UNKNOWN |
| 2026-06-03 16:17 | Review R002 | plan Step 1: REVISE |
| 2026-06-03 16:20 | R004 plan review | Reviewer requested Step 2-specific integration choices: catalog group, catalog_tiers_test update scope, and deterministic clock injection before implementation. |
| 2026-06-03 16:21 | R004 Step 2 integration plan | Catalog group: `workout-library`, because the tool is full-tier specifically due to active training-plan context and belongs next to `get_training_plan`/`apply_training_plan` rather than widening the existing events group. Update `internal/tools/catalog_tiers_test.go` to assert `get_planning_context` is `safety.ToolsetFull`; update `internal/toolcatalog/catalog.go` constants and athlete-scoped list. Implement `newGetPlanningContextToolWithClock(..., now func() time.Time, ...)` matching `get_today`/`get_events` so default week anchoring, as-of metadata, current fitness window, and race scan window are deterministic in tests. |
| 2026-06-03 16:21 | Review R005 | plan Step 2: APPROVE |
| 2026-06-03 16:23 | R006 plan review | Reviewer requested explicit Step 3 handler matrix, required catalog/registration assertions, and docs-generation decision before writing tests/docs. |
| 2026-06-03 16:24 | R006 Step 3 test/docs matrix | Handler tests will use a deterministic fake planning client and table/scenario cases for: terse default with no `full`, exact `_meta.source_tools`, read-only/write flags, planning scope, counts and windows; `include_full:true` widening only event/fitness/training-plan raw payloads; default upcoming Monday, supplied mid-week normalization, invalid week_start user error, and future week_start with current 7-day fitness calls; week/race ListEvents calls using Limit 500 and exact ranges; WORKOUT/RACE/RACE_*/NOTE/unknown classification and upcoming_races filtering; empty/partial/truncation caveat codes including read_only_no_atp, no_week_events, no_week_workouts, no_active_training_plan, partial_training_plan_summary, no_fitness_rows, no_upcoming_races, week_events_may_be_truncated, and upcoming_races_may_be_truncated. Catalog tests must assert full tier, workout-library group, toolcatalog known name, and athlete ACL scope. Docs: update CHANGELOG; inspect README for public tool list; run `make docs-tools` if README/web generated catalog data is expected to change. |
| 2026-06-03 16:20 | Review R003 | plan Step 1: APPROVE |
| 2026-06-03 16:22 | Review R004 | plan Step 2: UNKNOWN |
| 2026-06-03 16:25 | Review R005 | plan Step 2: APPROVE |
| 2026-06-03 16:30 | Review R006 | plan Step 3: UNKNOWN |
