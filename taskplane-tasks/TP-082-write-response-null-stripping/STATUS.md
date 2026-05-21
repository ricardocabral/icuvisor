# TP-082: Null stripping for write-tool responses — Status

**Current Step:** Step 6: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-05-20
**Review Level:** 2
**Review Counter:** 14
**Iteration:** 1
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no protected docs are changed without explicit approval

---

### Step 1: Audit write response shaping
**Status:** ✅ Complete

- [x] Audit registered `RequirementWrite` tools from `internal/tools/catalog.go` and classify upstream echo payloads, including prompt-scope omissions `link_activity_to_event` and `add_activity_message`.
- [x] Record a per-tool response-shaping matrix in STATUS.md covering builders, `encodeShaped` include-full mode, shaped collections, null stripping, raw/full preservation, and planned test locations.
- [x] Record intentional exceptions and high-risk divergent paths in STATUS.md before changing code.

---

### Step 2: Add failing golden tests
**Status:** ✅ Complete

- [x] Add or update write-tool fixture tests from the Step 2 matrix with sparse/null upstream fields for every audited write tool.
- [x] Assert terse defaults using key-presence checks so null keys are absent while meaningful zero/false/empty-string values remain.
- [x] Assert `include_full` raw-null preservation only for tools that already support it, and record expected red tests before Step 3.

---

### Step 3: Apply shared shaping consistently
**Status:** ✅ Complete

- [x] Route write responses through shared response shaping rather than bespoke map cleanup.
- [x] Preserve existing `_meta.server_version`, scale labels, and missing-field behavior.
- [x] Avoid changing request payloads.
- [x] Update custom-item write output schema descriptions to match default null-stripped detail shape.

---

### Step 4: Verify full write cluster
**Status:** ✅ Complete

- [x] Run targeted write-tool tests.
- [x] Run `make test`, `make build`, and `make lint`.
- [x] Update CHANGELOG.md.

---


### Step 5: Testing & Verification
**Status:** ✅ Complete

- [x] Targeted tests passing
- [x] FULL test suite passing: `make test`
- [x] Build passes: `make build`
- [x] Lint passes: `make lint`
- [x] All failures fixed or documented as pre-existing unrelated failures

---

### Step 6: Documentation & Delivery
**Status:** ✅ Complete

- [x] "Must Update" docs modified
- [x] "Check If Affected" docs reviewed
- [x] Discoveries logged
- [x] Final commit includes task ID

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |
| R002 | Plan | 1 | APPROVE | `.reviews/R002-plan-step1.md` |
| R003 | Code | 1 | APPROVE | `.reviews/R003-code-step1.md` |
| R004 | Plan | 2 | REVISE | `.reviews/R004-plan-step2.md` |
| R005 | Plan | 2 | REVISE | `.reviews/R005-plan-step2.md` |
| R006 | Plan | 2 | APPROVE | `.reviews/R006-plan-step2.md` |
| R007 | Code | 2 | APPROVE | `.reviews/R007-code-step2.md` |
| R008 | Plan | 3 | APPROVE | `.reviews/R008-plan-step3.md` |
| R009 | Code | 3 | REVISE | `.reviews/R009-code-step3.md` |
| R010 | Code | 3 | APPROVE | `.reviews/R010-code-step3.md` |
| R011 | Plan | 4 | APPROVE | `.reviews/R011-plan-step4.md` |
| R012 | Code | 4 | APPROVE | `.reviews/R012-code-step4.md` |
| R013 | Plan | 5 | APPROVE | `.reviews/R013-plan-step5.md` |
| R014 | Code | 5 | APPROVE | `.reviews/R014-code-step5.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Custom-item write responses hard-coded full shaping and preserved upstream null keys by default. | Fixed by routing create/update custom-item write responses through default shared null stripping and aligning output schema text. | `internal/tools/create_custom_item.go`, `internal/tools/update_custom_item.go`, `internal/tools/custom_item_write_validation.go` |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-20 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-20 12:41 | Task started | Runtime V2 lane-runner execution |
| 2026-05-20 12:41 | Step 0 started | Preflight |
| 2026-05-20 13:34 | Worker iter 1 | done in 3175s, tools: 213 |
| 2026-05-20 13:34 | Task complete | .DONE created |

---

## Blockers

*None*

---

## Notes

Step 1 concrete audit plan:
- Evidence commands/files: inspect `internal/tools/catalog.go` registrations, `grep -R "Requirement: RequirementWrite" internal/tools`, `grep -R "encodeShaped"` across all write tools, and targeted shared builders in `get_events.go`, `get_wellness_data.go`, `get_workout_library.go`, and `get_custom_item_by_id.go`.
- Scope write tools from registered catalog: `add_or_update_event`, `link_activity_to_event`, `add_activity_message`, `update_wellness`, `update_sport_settings`, `apply_training_plan`, `create_workout`, `update_workout`, `create_custom_item`, `update_custom_item`.
- Audit matrix fields to populate before code changes: upstream echo type, response builder, include-full mode, shaped row collections, terse null-key behavior, full/debug raw-null behavior, planned Step 2 test location.
- Intentional exceptions must be recorded for terse-only confirmation/write paths before Step 2.
- High-risk paths to inspect explicitly: custom item create/update hard-coded full shaping, wellness double shaping, event rows used by add/update/apply plan, and include-full raw confirmations for activity message/link activity.

Step 1 audit scope from registered `RequirementWrite` catalog (`internal/tools/catalog.go` plus `Requirement: RequirementWrite` grep):
| Tool | Upstream echo payload classification |
|------|--------------------------------------|
| `add_or_update_event` | Returns `intervals.Event` from `EventWriterClient.AddOrUpdateEvent`; upstream raw map can be exposed under `event.full` only when requested. |
| `link_activity_to_event` | Returns linked activity/update result with `Raw`; terse confirmation by default, raw echo under top-level `full` only with `include_full`. |
| `add_activity_message` | Returns activity message result with `Raw`; terse confirmation by default, raw echo under top-level `full` only with `include_full`. |
| `update_wellness` | Returns `intervals.Wellness` from `WellnessWriterClient.UpdateWellness`; raw wellness echo can be exposed under `wellness.full` with `include_full`. |
| `update_sport_settings` | Returns `intervals.SportSettings`, but response intentionally echoes selected writable settings/metadata rather than full upstream object. |
| `apply_training_plan` | Creates events via `AddOrUpdateEvent`; returns proposed rows and created event rows, but no user-facing `include_full` path. |
| `create_workout` | Returns `intervals.Workout`; response intentionally uses terse workout row only. |
| `update_workout` | Returns `intervals.Workout`; response intentionally uses terse workout row only. |
| `create_custom_item` | Returns `intervals.CustomItem`; current write response reuses full custom-item detail shape and hard-codes full shaping. |
| `update_custom_item` | Returns `intervals.CustomItem`; current write response reuses full custom-item detail shape and hard-codes full shaping. |

Delete tools are excluded from Step 1 because they return delete confirmations/pre-read snapshots rather than write echo payloads for this task's null-stripping mission.

Step 1 response-shaping matrix:
| Tool | Response builder | `encodeShaped` include-full | Row collections | Terse null-key behavior | Full/debug behavior | Planned Step 2 test location |
|------|------------------|-----------------------------|-----------------|-------------------------|---------------------|------------------------------|
| `add_or_update_event` | `shapeAddOrUpdateEventResponse` -> `eventRow(event, args.IncludeFull)` | `args.IncludeFull` | `nil` | Shared shaper strips null map keys from terse event wrapper/row; `event.full` absent by default. | `include_full` preserves upstream raw nulls under `event.full`; debug meta is controlled by common shaper. | `internal/tools/add_or_update_event_test.go` |
| `link_activity_to_event` | Bespoke `linkActivityToEventResponse` confirmation; raw only assigned to `Full` when requested | `args.IncludeFull` | `nil` | Confirmation has no raw sparse object by default; nil warning/full keys are omitted/stripped. | `include_full` preserves upstream raw nulls under top-level `full`. | `internal/tools/link_activity_to_event_test.go` |
| `add_activity_message` | Bespoke `addActivityMessageResponse` confirmation; raw only assigned to `Full` when requested | `args.IncludeFull` | `nil` | Confirmation has no raw sparse object by default; nil full key omitted/stripped. | `include_full` preserves upstream raw nulls under top-level `full`. | `internal/tools/add_activity_message_test.go` |
| `update_wellness` | `shapeUpdateWellnessResponse` -> pre-shapes `wellnessRow(row, includeFull)` then wraps in `updateWellnessResponse` | `args.IncludeFull` | `[]string{"wellness"}` (but wrapped value is a map, already shaped) | Inner shared shaper strips null wellness keys; outer shared shaper preserves already-shaped row and wrapper meta. | `include_full` preserves raw nulls under `wellness.full`; existing nested wellness `_meta` is intentional/current behavior to preserve during tests. | `internal/tools/update_wellness_test.go` |
| `update_sport_settings` | `shapeUpdateSportSettingsResponse` selected echo (`updateSportSettingsEcho`) | `false` | `nil` | Shared shaper strips nil optional echo fields; no raw upstream object is exposed. | No `include_full` request support; debug meta only common. | `internal/tools/update_sport_settings_test.go` |
| `apply_training_plan` | `applyTrainingPlan` proposed rows plus created `eventRow(created, false)` | `false` | `[]string{"proposed_events", "created_events"}` | Shared shaper strips nulls from proposed/created row maps; created rows never include `full`. | No `include_full` request support; raw created event details intentionally absent. | `internal/tools/apply_training_plan_test.go` |
| `create_workout` | `shapeCreateWorkoutResponse` -> `workoutToRow(workout, false)` | `false` | `nil` | Shared shaper strips nil workout-row keys; no `full` raw workout. | No `include_full` request support; raw workout details intentionally summarized. | `internal/tools/create_workout_test.go` |
| `update_workout` | `shapeUpdateWorkoutResponse` -> `workoutToRow(workout, false)` | `false` | `nil` | Shared shaper strips nil workout-row keys; meaningful zero/false/empty values should remain. | No `include_full` request support; raw workout details intentionally summarized. | `internal/tools/update_workout_test.go` |
| `create_custom_item` | `shapeCustomItemWriteResponse` -> `shapeGetCustomItemByIDResponse` full map | `true` hard-coded | `nil` | Diverges: hard-coded full mode preserves nulls by default, despite no `include_full` input. | Default already behaves like full/debug and preserves raw nulls in `custom_item`. | `internal/tools/create_custom_item_test.go` |
| `update_custom_item` | `shapeCustomItemWriteResponse` -> `shapeGetCustomItemByIDResponse` full map | `true` hard-coded | `nil` | Diverges: hard-coded full mode preserves nulls by default, despite no `include_full` input. | Default already behaves like full/debug and preserves raw nulls in `custom_item`. | `internal/tools/update_custom_item_test.go` |

Step 1 intentional exceptions and risk notes before code changes:
- Intentional terse-only/no-raw exceptions: `update_sport_settings`, `apply_training_plan`, `create_workout`, and `update_workout` have no `include_full` argument today; Step 2 should assert null stripping and preservation of zero/false/empty values, not invent new full behavior.
- Intentional confirmation-with-optional-full exceptions: `link_activity_to_event` and `add_activity_message` remain terse confirmations by default; Step 2 should assert raw null preservation only when `include_full` is true.
- High-risk divergence requiring Step 3 fix: `create_custom_item` and `update_custom_item` hard-code `encodeShaped(..., true, ...)`, preserving null keys in default write responses; they should route default writes through shared terse shaping while preserving raw detail via an explicit/debug-supported path if available.
- Wellness high-risk/current behavior: `update_wellness` double-shapes the wellness row, producing nested row `_meta` plus wrapper `_meta`; tests should preserve this existing metadata behavior while checking null stripping.
- Event-row paths: `add_or_update_event` and `apply_training_plan` use `eventRow`; tests should verify default `full` is absent and created rows are stripped without altering event write request payloads.

Step 2 concrete test matrix before editing tests:
| Tool | Test file/name plan | Sparse null fixture fields | Preserve zero/false/empty fields | Existing full path/raw-null assertion | Expected today |
|------|---------------------|----------------------------|----------------------------------|---------------------------------------|----------------|
| `add_or_update_event` | `add_or_update_event_test.go`: add default sparse-null and include-full variants | event raw `notes: null`, `icu_training_load: null`/extra nullable field | pointer-backed `load_target: 0`, `distance_meters: 0`, `name: ""` if row shape preserves it | Yes: `include_full` must keep nullable key in `event.full` with `value == nil` and key present | Pass expected |
| `link_activity_to_event` | `link_activity_to_event_test.go`: strengthen include-full and default confirmation tests | linked raw `paired_event_id: null`, `gear: null` | terse `warnings` absent/empty as intended; IDs/status preserved | Yes: top-level `full` keeps nullable raw key with key-present nil | Pass expected |
| `add_activity_message` | `add_activity_message_test.go`: strengthen include-full and default confirmation tests | message raw `moderation: null`, `parent_id: null` | terse IDs/status preserved; empty athlete ID metadata omitted/stripped | Yes: top-level `full` keeps nullable raw key with key-present nil | Pass expected |
| `update_wellness` | `update_wellness_test.go`: add sparse default and include-full assertions | wellness raw `weight: null`, `injury: null`, `notes: null` | pointer-backed `restingHR: 0`/`locked: false` if available; preserve nested `_meta` behavior | Yes: `wellness.full` keeps nullable raw key with key-present nil | Pass expected |
| `update_sport_settings` | `update_sport_settings_test.go`: default echo sparse/false test | updated upstream missing/null threshold fields via zero/nil struct fields | emitted wrapper `_meta.zones_provided: false`, selected echo IDs/status fields; do not assert `zone_definitions_overwritten:false` because it is `omitempty` today | No supported `include_full`; do not add API surface | Pass expected |
| `apply_training_plan` | `apply_training_plan_test.go`: created event terse sparse test | created event raw `description: null`, `calendar_id: null` | pointer-backed created row values such as `load_target: 0`, `distance_meters: 0`; `full` absent | No supported `include_full`; raw created event details intentionally absent | Pass expected |
| `create_workout` | `create_workout_test.go`: default sparse workout test | workout raw `description: null`, `target: null` | `indoor: false`, `distance_meters: 0` if pointer/shape preserves; IDs/name/sport | No supported `include_full`; raw workout details intentionally absent | Pass expected |
| `update_workout` | `update_workout_test.go`: default sparse workout test | workout raw `description: null`, `target: null` | pointer-backed/map-backed values such as `indoor: false` and `distance_meters: 0`; do not assert empty `description` because it is a plain `omitempty` string today | No supported `include_full`; raw workout details intentionally absent | Pass expected |
| `create_custom_item` | `create_custom_item_test.go`: default sparse custom-item test | custom item raw/content `description: null`, `image: null`, nested content nullable field | map-backed `index: 0`, `hide_script: false`, `description: ""` when supplied | No user-facing `include_full`; default should be terse after Step 3 | Fail expected today due hard-coded full shaping |
| `update_custom_item` | `update_custom_item_test.go`: default sparse custom-item test | custom item raw/content `description: null`, `image: null`, nested content nullable field | map-backed `index: 0`, `hide_script: false`, `description: ""` when supplied | No user-facing `include_full`; default should be terse after Step 3 | Fail expected today due hard-coded full shaping |

Step 2 assertion plan:
- Default/terse assertions must distinguish absent key from present JSON null: `if _, ok := row["nullable_field"]; ok { t.Fatal(...) }`.
- Full/raw assertions must distinguish key presence with nil value: `value, ok := full["nullable_field"]; if !ok || value != nil { t.Fatal(...) }`.
- Add a local recursive helper in tests only if repeated object-wide no-null checks remain clear and scoped.
- Targeted verification command after test edits: `go test ./internal/tools -run 'Test(AddOrUpdateEvent|LinkActivityToEvent|AddActivityMessage|UpdateWellness|UpdateSportSettings|ApplyTrainingPlan|CreateWorkout|UpdateWorkout|CreateCustomItem|UpdateCustomItem)'`.
- Expected red tests before Step 3 are limited to default null-key stripping for `create_custom_item` and `update_custom_item`, caused by hard-coded full shaping. If `update_sport_settings`, `update_workout`, or any other test fails, record it as an unexpected discovery unless the assertion was deliberately changed and scoped.

Step 2 targeted test run after adding golden tests:
- Command: `go test ./internal/tools -run 'Test(AddOrUpdateEvent|LinkActivityToEvent|AddActivityMessage|UpdateWellness|UpdateSportSettings|ApplyTrainingPlan|CreateWorkout|UpdateWorkout|CreateCustomItem|UpdateCustomItem)'`
- Outcome before Step 3: expected red tests only: `TestCreateCustomItemDefaultStripsSparseNullsAndPreservesMapValues` and `TestUpdateCustomItemDefaultStripsSparseNullsAndPreservesMapValues`, both failing because default custom-item write responses preserve `image: null` under hard-coded full shaping. Other targeted write-tool tests reached pass status in this run.

R009 code review required change:
- Update `create_custom_item` and `update_custom_item` output schema descriptions that still promise a full/verbatim custom-item read shape; wording must match the new null-stripped default while leaving `get_custom_item_by_id` unchanged.

Step 6 docs review:
- Must-update docs: `CHANGELOG.md` updated under `[Unreleased]`; `STATUS.md` kept current.
- Check-if-affected docs reviewed: `README.md` and `web/content/reference/tools.md` had no matching tool/output text to update; `docs/prd/PRD-icuvisor.md` already states null-stripping/default terse response shaping and no product-scope divergence was introduced.
| 2026-05-20 12:45 | Review R001 | plan Step 1: UNKNOWN |
| 2026-05-20 12:47 | Review R002 | plan Step 1: APPROVE |
| 2026-05-20 12:52 | Review R003 | code Step 1: APPROVE |
| 2026-05-20 12:55 | Review R004 | plan Step 2: REVISE |
| 2026-05-20 12:59 | Review R005 | plan Step 2: REVISE |
| 2026-05-20 13:02 | Review R006 | plan Step 2: APPROVE |
| 2026-05-20 13:10 | Review R007 | code Step 2: APPROVE |
| 2026-05-20 13:13 | Review R008 | plan Step 3: APPROVE |
| 2026-05-20 13:17 | Review R009 | code Step 3: UNKNOWN |
| 2026-05-20 13:20 | Review R010 | code Step 3: APPROVE |
| 2026-05-20 13:21 | Review R011 | plan Step 4: APPROVE |
| 2026-05-20 13:25 | Review R012 | code Step 4: APPROVE |
| 2026-05-20 13:26 | Review R013 | plan Step 5: APPROVE |
| 2026-05-20 13:31 | Review R014 | code Step 5: APPROVE |
