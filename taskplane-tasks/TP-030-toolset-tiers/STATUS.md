# TP-030-toolset-tiers: TP-030-toolset-tiers — Status

**Current Step:** Step 2: Per-tool tier membership
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-14
**Review Level:** 2
**Review Counter:** 5
**Iteration:** 1
**Size:** M

---

### Step 1: Tier enum and parsing

**Status:** ✅ Complete

- [x] Enum: `core` (default) and `full`; case-insensitive parsing; unknown/empty → `core`
- [x] Wire `ICUVISOR_TOOLSET` through config loading (`Config.Toolset`, `.env`/env precedence, defensive string rendering)
- [x] Propagate parsed toolset into app startup (`ServerInfo`) and log the resolved tier exactly once without tool names
- [x] Pin Step 1 behavior with tests for parsing, config loading, startup propagation, and minimal logging
- [x] Decide and document the package boundary: extend `internal/safety` vs new `internal/toolset`. Record the choice and rationale in `STATUS.md`

### Step 2: Per-tool tier membership

**Status:** 🟨 In Progress

- [x] Add `tools.Tool.Toolset safety.Toolset` plus an effective-tier helper that defaults empty to `full` and rejects unknown non-empty in-code tier values during validation
- [x] Each tool self-declares its tier (`core` or `full`) in its constructor; no production name→tier map is introduced
- [x] Curate the `core` set to the §7.2.E daily-use path: read activities/fitness/wellness/events, write events/wellness/messages, plus `icuvisor_list_advanced_capabilities`. Target ~17 tools; record the exact list in `STATUS.md`
- [x] Test matrix: every registered current tool's tier membership is asserted in a table-driven test that fails on missing expected tools and unexpected newly registered tools
- [x] Preserve Step 2 boundaries: no `tools/list` filtering, no startup skip counts, and no implementation of `icuvisor_list_advanced_capabilities` yet

### Step 3: Registry filtering composition

**Status:** ⏳ Not started

### Step 4: `icuvisor_list_advanced_capabilities`

**Status:** ⏳ Not started

### Step 5: `_meta` surfacing + docs

**Status:** ⏳ Not started

### Step 6: Verify

**Status:** ⏳ Not started

---

## Reviews

| #   | Type | Step | Verdict | File |
| --- | ---- | ---- | ------- | ---- |
| R001 | plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |
| R002 | plan | 1 | APPROVE | `.reviews/R002-plan-step1.md` |
| R003 | code | 1 | APPROVE | `.reviews/R003-code-step1.md` |
| R004 | plan | 2 | REVISE | `.reviews/R004-plan-step2.md` |
| R005 | plan | 2 | APPROVE | `.reviews/R005-plan-step2.md` |

---

## Discoveries

| Discovery | Disposition | Location |
| --------- | ----------- | -------- |
| Package boundary for toolset tiers | Extend `internal/safety` rather than create `internal/toolset`, because TP-018 already centralizes registration-time environment gates and capability decisions there; toolset tiering is an orthogonal registration gate using the same pattern. | Step 1 / `internal/safety` |

---

## Execution Log

| Timestamp  | Action      | Outcome                          |
| ---------- | ----------- | -------------------------------- |
| 2026-05-14 | Task staged | Scaffolded from ROADMAP.md v0.4   |
| 2026-05-14 12:05 | Task started | Runtime V2 lane-runner execution |
| 2026-05-14 12:05 | Step 1 started | Tier enum and parsing |

---

## Blockers

_None_

---

## Notes

- R001 required the Step 1 plan to explicitly include a separate `safety.Toolset` API, config loader plumbing, app startup propagation/logging, and tests before implementation.
- R002 approved the revised Step 1 plan for implementation.
- R003 approved the Step 1 implementation; reviewer verified `go test ./...` and diff checks.
- R004 required Step 2 to pin a `Tool` metadata/helper API, keep membership self-declared, record an exact core/full tier table before coding, define the drift-catching catalog test, and avoid Step 3/4 scope.
- R005 approved the revised Step 2 plan; implementation kept invalid in-code tier validation on the MCP registration path and used a full-capability collecting registrar for the membership matrix.

### Step 2 tier plan

Current core tools (16 existing + `icuvisor_list_advanced_capabilities` planned in Step 4 = 17): `get_athlete_profile`, `get_activities`, `get_activity_details`, `get_activity_intervals`, `get_activity_splits`, `get_activity_messages`, `get_fitness`, `get_training_summary`, `get_best_efforts`, `get_wellness_data`, `get_events`, `get_event_by_id`, `add_or_update_event`, `update_wellness`, `add_activity_message`, `link_activity_to_event`, `icuvisor_list_advanced_capabilities`.

Current full-only tools: `get_power_curves`, `get_extended_metrics`, `get_activity_streams`, `get_training_plan`, `apply_training_plan`, `get_workout_library`, `get_workouts_in_folder`, `create_workout`, `update_workout`, `delete_workout`, `update_sport_settings`, `delete_sport_settings`, `get_custom_items`, `get_custom_item_by_id`, `create_custom_item`, `update_custom_item`, `delete_custom_item`, `delete_event`, `delete_events_by_date_range`, `delete_activity`, `delete_gear`.

Rationale: core covers profile context plus daily activity/fitness/wellness/event reads and non-destructive event/wellness/message writes; full holds raw/heavy reads, specialist workout-library/training-plan/custom-item/sport-settings surfaces, and all destructive delete tools.
| 2026-05-14 12:40 | Review R003 | code Step 1: APPROVE |
| 2026-05-14 12:45 | Review R004 | plan Step 2: REVISE |
| 2026-05-14 12:50 | Review R005 | plan Step 2: APPROVE |
