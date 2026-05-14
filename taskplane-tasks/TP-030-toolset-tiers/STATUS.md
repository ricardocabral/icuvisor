# TP-030-toolset-tiers: TP-030-toolset-tiers — Status

**Current Step:** Step 6: Verify
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-14
**Review Level:** 2
**Review Counter:** 16
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

**Status:** ✅ Complete

- [x] Add `tools.Tool.Toolset safety.Toolset` plus an effective-tier helper that defaults empty to `full` and rejects unknown non-empty in-code tier values during validation
- [x] Each tool self-declares its tier (`core` or `full`) in its constructor; no production name→tier map is introduced
- [x] Curate the `core` set to the §7.2.E daily-use path: read activities/fitness/wellness/events, write events/wellness/messages, plus `icuvisor_list_advanced_capabilities`. Target ~17 tools; record the exact list in `STATUS.md`
- [x] Test matrix: every registered current tool's tier membership is asserted in a table-driven test that fails on missing expected tools and unexpected newly registered tools
- [x] Preserve Step 2 boundaries: no `tools/list` filtering, no startup skip counts, and no implementation of `icuvisor_list_advanced_capabilities` yet

### Step 3: Registry filtering composition

**Status:** ✅ Complete

- [x] Propagate the already-resolved active toolset from `Config.Toolset`/`ServerInfo.Toolset` into `mcp.Options`/`safeRegistrar` without re-reading env or adding a tool-call override; empty defaults to `core`
- [x] Registration filters on tier **and** delete-mode after validation; `core` registers only core tools, `full` registers core+full tools, and a tool appears only when both gates allow it
- [x] Tools outside the active tier are **absent** from `tools/list`, not registered-and-erroring
- [x] Startup INFO line reports count-only `registered_count`, `skipped_toolset_count`, and `skipped_capability_count` with independent gate evaluation and no tool names
- [x] Add composition tests crossing active toolset (`core`/`full`) with delete mode (`none`/`safe`/`full`) plus protocol absence/logging coverage for hidden tools
- [x] Update unmarked test fixtures (`testEchoRegistry`, `capabilityRegistry`, protocol helpers) deliberately so default-core behavior is preserved rather than weakened

### Step 4: `icuvisor_list_advanced_capabilities`

**Status:** ✅ Complete

- [x] Add a catalog-collecting registrar inside `internal/tools` so `icuvisor_list_advanced_capabilities` derives full-only rows from the same self-declared `Tool` metadata and first description sentences while forwarding registration normally
- [x] Register `icuvisor_list_advanced_capabilities` after collecting the existing catalog, mark it `core`, use a no-argument closed input schema, exclude core tools/self, and keep it available in both active tiers
- [x] Propagate active toolset into `tools.RegistryOptions` from startup without re-reading env or adding request overrides, so the handler can report core vs full status
- [x] Lives in `core`; returns the `full`-only tools with one-line summaries, requirements/delete-mode note, and the exact `ICUVISOR_TOOLSET=full` instruction to enable them
- [x] Output is static/derived from the catalog — no upstream calls; terse by default
- [x] When the tier is already `full`, it still works and says so
- [x] Add drift/behavior/protocol/schema tests: tier matrix includes the new core tool, handler excludes core/self and includes known full tools, no upstream calls occur, core/full `tools/list` include the discoverability tool, and schema snapshot is committed
- [x] R012 fix: add `icuvisor_list_advanced_capabilities` to the adversarial static catalog as a read tool so `go test ./...` passes and no-confirm/delete-mode guardrails include it

### Step 5: `_meta` surfacing + docs

**Status:** ✅ Complete

- [x] Extend `internal/response` with process-level `SetToolset`/`Toolset` state parallel to delete mode, defaulting invalid/empty to `core`, and have `addCommonMeta` overwrite response-owned `_meta.toolset`
- [x] Ensure direct-return `icuvisor_list_advanced_capabilities` responses also include `_meta.toolset` in both structured content and serialized text using the same response-owned source
- [x] Add targeted tests for response metadata defaults/overwrites, startup setter propagation, and advanced-capabilities `_meta.toolset` in core/full modes
- [x] README: short section documenting `ICUVISOR_TOOLSET`, the two tiers, the default, restart requirement, delete-mode orthogonality, and the discoverability tool
- [x] CHANGELOG `[Unreleased]` entry

### Step 6: Verify

**Status:** 🟨 In Progress

- [ ] `make test`, `make build`, `make lint`, `go test -race ./...`
- [ ] Manual: start the binary in `core` and `full`; confirm `tools/list` counts and that `icuvisor_list_advanced_capabilities` is present in `core`

---

## Reviews

| #   | Type | Step | Verdict | File |
| --- | ---- | ---- | ------- | ---- |
| R001 | plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |
| R002 | plan | 1 | APPROVE | `.reviews/R002-plan-step1.md` |
| R003 | code | 1 | APPROVE | `.reviews/R003-code-step1.md` |
| R004 | plan | 2 | REVISE | `.reviews/R004-plan-step2.md` |
| R005 | plan | 2 | APPROVE | `.reviews/R005-plan-step2.md` |
| R006 | code | 2 | APPROVE | `.reviews/R006-code-step2.md` |
| R007 | plan | 3 | REVISE | `.reviews/R007-plan-step3.md` |
| R008 | plan | 3 | APPROVE | `.reviews/R008-plan-step3.md` |
| R009 | code | 3 | APPROVE | `.reviews/R009-code-step3.md` |
| R010 | plan | 4 | REVISE | `.reviews/R010-plan-step4.md` |
| R011 | plan | 4 | APPROVE | `.reviews/R011-plan-step4.md` |
| R012 | code | 4 | REVISE | `.reviews/R012-code-step4.md` |
| R013 | code | 4 | APPROVE | `.reviews/R013-code-step4.md` |
| R014 | plan | 5 | REVISE | `.reviews/R014-plan-step5.md` |
| R015 | plan | 5 | APPROVE | `.reviews/R015-plan-step5.md` |
| R016 | code | 5 | APPROVE | `.reviews/R016-code-step5.md` |

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
- R006 approved the Step 2 implementation; reviewer verified `go test ./...` and diff checks.

### Step 2 tier plan

Current core tools (16 existing + `icuvisor_list_advanced_capabilities` planned in Step 4 = 17): `get_athlete_profile`, `get_activities`, `get_activity_details`, `get_activity_intervals`, `get_activity_splits`, `get_activity_messages`, `get_fitness`, `get_training_summary`, `get_best_efforts`, `get_wellness_data`, `get_events`, `get_event_by_id`, `add_or_update_event`, `update_wellness`, `add_activity_message`, `link_activity_to_event`, `icuvisor_list_advanced_capabilities`.

Current full-only tools: `get_power_curves`, `get_extended_metrics`, `get_activity_streams`, `get_training_plan`, `apply_training_plan`, `get_workout_library`, `get_workouts_in_folder`, `create_workout`, `update_workout`, `delete_workout`, `update_sport_settings`, `delete_sport_settings`, `get_custom_items`, `get_custom_item_by_id`, `create_custom_item`, `update_custom_item`, `delete_custom_item`, `delete_event`, `delete_events_by_date_range`, `delete_activity`, `delete_gear`.

Rationale: core covers profile context plus daily activity/fitness/wellness/event reads and non-destructive event/wellness/message writes; full holds raw/heavy reads, specialist workout-library/training-plan/custom-item/sport-settings surfaces, and all destructive delete tools.

### Step 3 filtering plan

Active toolset propagation: `app.defaultStartServer` passes the startup-resolved `ServerInfo.Toolset` into `mcp.Options.Toolset`; `NewServer` normalizes empty/invalid active values with `safety.ParseToolset` once for the registrar. No environment re-read and no model-controlled override.

Filtering semantics: `safeRegistrar.validateTool` runs before skip decisions, so invalid non-empty in-code toolsets still fail registration. A tool is registered only if the active toolset allows `tool.EffectiveToolset()` and delete-mode capability allows its `Requirement`. `core` allows only declared core tools; `full` allows declared core and full tools. Empty tool declarations are effective full from Step 2, so future unmarked tools stay out of default core.

Skip-count semantics: evaluate the toolset gate and capability gate independently for every validated tool; increment `skipped_toolset_count` when the active tier would hide it and `skipped_capability_count` when delete/write capability would hide it, even if both gates suppress the same tool. Register only when neither gate suppresses it. Startup logs count-only `registered_count`, `skipped_toolset_count`, and `skipped_capability_count`, never tool names/descriptions.

Composition test matrix: synthetic tools cover core read, core write, full read, full write, and full delete. Expected visible sets: `core+none` → core read only; `core+safe` → core read/core write; `core+full` and zero-value active toolset + full delete mode → core read/core write; `full+none` → core read/full read only; `full+safe` → core read/core write/full read/full write; `full+full` → all synthetic tools. Protocol coverage must show a hidden full-only tool is absent from `tools/list` under core and cannot be called as a registered tool. Existing test-only tools are marked core only when the test needs default visibility; otherwise tests set active toolset full deliberately.

- R008 approved the revised Step 3 plan and suggested adding an explicit zero-value `mcp.Options.Toolset` defaults-to-core assertion, which was folded into the composition matrix.
- R009 approved the Step 3 implementation.
- R010 required Step 4 to pin catalog derivation from existing `Tool` metadata, active-toolset propagation to handlers, a concrete response shape/instruction, and drift/protocol/schema tests before implementation.
- R011 approved the revised Step 4 plan for implementation.
- R012 requested adding the new discoverability tool to `internal/safety/adversarial_test.go`'s static catalog; reviewer found `go test ./...` failing only on that catalog count drift.
- R013 approved the revised Step 4 implementation.
- R014 required Step 5 to pin response-owned metadata semantics, direct-return tool handling, targeted tests, and exact README/CHANGELOG scope before implementation.
- R015 approved the revised Step 5 plan for implementation.
- R016 approved the Step 5 implementation.

### Step 5 metadata/docs plan

Response mechanism: add process-global toolset state in `internal/response` beside delete mode (`SetToolset`, `Toolset`), store `safety.ParseToolset(...).String()`, and call `response.SetToolset(toolset.String())` from `app.defaultStartServer` using the already-resolved startup value. No handler/env re-read.

Metadata semantics: `addCommonMeta` owns `server_version`, `delete_mode`, `toolset`, and units. It preserves existing caller `_meta` keys (counts, pagination, scales, delete-mode notes, etc.) but overwrites stale caller-supplied `toolset` with the normalized process value, matching the response-owned metadata contract.

Direct-return tool handling: update `icuvisor_list_advanced_capabilities` to source `_meta.toolset` from `response.Toolset()` so both `StructuredContent` and marshaled text JSON include the same common process value. Keep its catalog-specific `_meta.count`, `_meta.source`, and `_meta.delete_mode_note` intact.

Tests: extend `internal/response/shaper_test.go` for default core, explicit full, invalid/empty fallback, and stale `_meta.toolset` overwrite while preserving other meta keys; update shared expected-meta helper to include default `toolset`. Add app startup propagation coverage and advanced-capabilities handler assertions for `_meta.toolset` in core and full mode text/structured output.

Docs: README documents `ICUVISOR_TOOLSET` near delete safety: `core` default, `full` opt-in, invalid/empty fallback to `core`, restart required, `icuvisor_list_advanced_capabilities` discoverability, and `ICUVISOR_DELETE_MODE` remains separate for destructive tools. CHANGELOG `[Unreleased]` gets a concise bullet for toolset tiers, discoverability tool, and `_meta.toolset`.

### Step 4 discoverability plan

Catalog derivation: `defaultRegistry.Register` will use a small collecting/wrapping registrar to record every existing tool as it is added, forward those tools to the real registrar unchanged, then register `icuvisor_list_advanced_capabilities` using rows derived from collected tools whose `EffectiveToolset()==full`. Summaries come from the first sentence of each tool description. No production name→tier or summary map is introduced.

Registration/ordering: the new tool is added after existing catalog collection so it can include all full-only tools and exclude itself/core tools. It is declared with `coreTool(...)`, read requirement, `additionalProperties:false` no-argument input schema, and generic structured output schema. Step 3 semantics make it visible in both active `core` and `full` tiers.

Active toolset propagation: add `Toolset safety.Toolset` to `tools.RegistryOptions`, default empty/invalid to core, pass `toolset` from `app.defaultStartServer` into `tools.NewRegistryWithOptions`, and do not read env or accept request arguments in the handler.

Response shape: structured content includes `current_toolset`, `status`, `enable_instruction`, `advanced_capabilities: [{name, summary, requirement}]`, and `_meta: {count, source, delete_mode_note}`. Text content also contains `ICUVISOR_TOOLSET=full` and tells the user to set it in the MCP client/server environment and restart icuvisor. In full mode, status says the full toolset is already enabled while still returning the same full-only catalog.

Tests: update tier matrix expected core list/count for `icuvisor_list_advanced_capabilities`; add handler tests for core/full output, known full inclusion (`get_power_curves`), core/self exclusion (`get_athlete_profile`, `icuvisor_list_advanced_capabilities`), one-line summaries, enable instruction, and no upstream calls using a panic client; add protocol coverage for default core and full mode visibility; update profile-only registry expectations; add the schema snapshot and ensure description first sentence is distinct.
| 2026-05-14 12:40 | Review R003 | code Step 1: APPROVE |
| 2026-05-14 12:45 | Review R004 | plan Step 2: REVISE |
| 2026-05-14 12:50 | Review R005 | plan Step 2: APPROVE |
| 2026-05-14 12:59 | Review R006 | code Step 2: APPROVE |
| 2026-05-14 13:04 | Review R007 | plan Step 3: REVISE |
| 2026-05-14 13:07 | Review R008 | plan Step 3: APPROVE |
| 2026-05-14 13:16 | Review R009 | code Step 3: APPROVE |
| 2026-05-14 13:23 | Review R010 | plan Step 4: REVISE |
| 2026-05-14 13:26 | Review R011 | plan Step 4: APPROVE |
| 2026-05-14 13:42 | Review R012 | code Step 4: REVISE |
| 2026-05-14 13:47 | Review R013 | code Step 4: APPROVE |
| 2026-05-14 13:50 | Review R014 | plan Step 5: REVISE |
| 2026-05-14 13:53 | Review R015 | plan Step 5: APPROVE |
| 2026-05-14 14:03 | Review R016 | code Step 5: APPROVE |
