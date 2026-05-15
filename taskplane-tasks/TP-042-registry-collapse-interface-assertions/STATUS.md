# TP-042-registry-collapse-interface-assertions — Status

**Current Step:** Step 3: Collapse `schemaCatalogClient`
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 10
**Iteration:** 1
**Size:** M

---

### Step 1: Map the current assertion chain

**Status:** ✅ Complete

- [x] Enumerate every `XxxClient` interface in `internal/tools/`, including required methods, tool constructors, registry condition/order, and schemaCatalogClient coverage
- [x] Verify all are satisfied by `*intervals.Client`
- [x] Identify existing unit-test fakes and `NewRegistry`/`NewRegistryWithOptions` registry-level fake call sites
- [x] Inventory schema snapshot catalog membership versus full production registration and note parity risk
- [x] Record special constructor coupling and conditional registration semantics (`customItemsClient`, splits, link activity optional clients)
- [x] Decide direct-dep vs `Deps` struct with migration action for tests/toolchecks and no-network dummy-client rationale
- [x] R003: Correct `schemaCatalogClient` coverage for structurally satisfied `ApplyTrainingPlanClient`
- [x] R003: Add `GenerateToolCatalog`/confusable-name checker to toolchecks migration inventory
- [x] R003: Add MCP `TestProtocolGetAthleteProfileDispatch` registry fake migration plan

### Step 2: Refactor `Register`

**Status:** ✅ Complete

- [x] Change `NewRegistry` / `NewRegistryWithOptions` and `defaultRegistry` storage to direct `*intervals.Client` while preserving `Registry.Register(context.Context, Registrar) error`
- [x] Replace assertion blocks with direct constructor calls in the existing order, including optional collaborator and custom-item couplings
- [x] Add per-tool `AddTool` error wrapping that names the failing tool, including `icuvisor_list_advanced_capabilities`
- [x] Preserve delete-mode / toolset / capability gating via existing `Tool` metadata/downstream registrar behavior; do not add registry-side filtering
- [x] Migrate Step 2 call sites/tests enough to keep the tree buildable after constructor signature changes
- [x] Fix hardcoded `getAthleteProfileName` missing-client/registrar error messages
- [x] R007: Format changed Go files with gofmt/goimports-clean output
- [x] R007: Remove stale registry fake types no longer referenced by tests

### Step 3: Collapse `schemaCatalogClient`

**Status:** ✅ Complete

- [x] Replace with minimal fake or real client
- [x] Snapshot output byte-identical

### Step 4: Tests

**Status:** ⏳ Not started

- [ ] `make test` + `make test-race`
- [ ] Schema-stability snapshot unchanged
- [ ] Add regression guard test for full registration

### Step 5: Verify

**Status:** ⏳ Not started

- [ ] `make build` / `test` / `test-race` / `lint`
- [ ] Manual `list_tools` parity check

---

## Decisions

Step 1 dependency-shape decision: choose direct `*intervals.Client` for `tools.NewRegistry`/`NewRegistryWithOptions`, not a `Deps` struct. Rationale: production has exactly one concrete dependency today; no logger/clock/secondary dependency is required by registry; every per-tool narrow interface remains in constructors for unit tests; compile-time assertions verified `*intervals.Client` satisfies all 33 interfaces. Migration: registry/catalog tests that only inspect registration should use a no-network dummy real client built with dummy API key, normalized athlete ID, loopback base URL, and optionally a panic `RoundTripper` to prove registration does not execute HTTP. Tests that exercise handler behavior with narrow fakes should instantiate the target `newXxxTool` directly. Toolchecks schema parity should not use unrestricted full-client registration unless paired with a schema catalog filter, because it adds 8 tools versus current snapshots.

## Notes

Step 1 registry/interface inventory:

| # | Interface (file) | Required methods (summary) | Constructor(s) | Current registry condition/order | schemaCatalogClient? |
|---|---|---|---|---|---|
| 1 | `ProfileClient` (`get_athlete_profile.go`) | `GetAthleteProfile` | `newGetAthleteProfileTool` | mandatory first; nil error currently hardcodes `get_athlete_profile` | yes |
| 2 | `FitnessClient` (`get_fitness.go`) | `ListAthleteSummary` | `newGetFitnessTool`, `newGetTrainingSummaryTool` | conditional after profile | yes |
| 3 | `WellnessClient` (`get_wellness_data.go`) | `ListWellness` | `newGetWellnessDataTool` | conditional | yes |
| 4 | `WellnessWriterClient` (`update_wellness.go`) | `UpdateWellness` | `newUpdateWellnessTool` | conditional; write requirement inside tool | yes |
| 5 | `SportSettingsWriterClient` (`update_sport_settings.go`) | `UpdateSportSettings` | `newUpdateSportSettingsTool` | conditional; capability passed to tool | yes |
| 6 | `BestEffortsClient` (`get_fitness.go`) | power/HR/pace curves | `newGetBestEffortsTool` | conditional | yes |
| 7 | `PowerCurvesClient` (`get_fitness.go`) | `ListAthletePowerCurves` | `newGetPowerCurvesTool` | conditional | yes |
| 8 | `ActivitiesClient` (`get_activities.go`) | `ListActivities` | `newGetActivitiesTool` | conditional | yes |
| 9 | `EventsClient` (`get_events.go`) | `ListEvents` | `newGetEventsTool` | conditional | yes |
| 10 | `EventByIDClient` (`get_event_by_id.go`) | `GetEvent`, `ListEvents` | `newGetEventByIDTool` | conditional | yes |
| 11 | `EventWriterClient` (`add_or_update_event.go`) | `AddOrUpdateEvent` | `newAddOrUpdateEventTool` | conditional | yes |
| 12 | `ApplyTrainingPlanClient` (`apply_training_plan.go`) | embeds workout library/events/event writer | `newApplyTrainingPlanTool` | conditional; capability passed to tool | yes (structural via embedded interfaces; no explicit assertion) |
| 13 | `EventDeleterClient` (`delete_event.go`) | `GetEvent`, `DeleteEvent` | `newDeleteEventTool` | conditional; delete requirement inside tool | no |
| 14 | `EventsByDateRangeDeleterClient` (`delete_events_by_date_range.go`) | `ListEvents`, `DeleteEvent` | `newDeleteEventsByDateRangeTool` | conditional; delete requirement inside tool | no |
| 15 | `ActivityEventLinkClient` (`link_activity_to_event.go`) | `LinkActivityToEvent` | `newLinkActivityToEventTool` | conditional; optional `ActivityDetailsClient`/`EventByIDClient` looked up separately | yes |
| 16 | `TrainingPlanClient` (`get_training_plan.go`) | `GetTrainingPlan` | `newGetTrainingPlanTool` | conditional | yes |
| 17 | `WorkoutLibraryClient` (`get_workout_library.go`) | `ListWorkoutFolders`, `ListLibraryWorkouts` | `newGetWorkoutLibraryTool`, `newGetWorkoutsInFolderTool` | conditional | yes |
| 18 | `WorkoutCreatorClient` (`create_workout.go`) | `CreateLibraryWorkout` | `newCreateWorkoutTool` | conditional | yes |
| 19 | `WorkoutUpdaterClient` (`update_workout.go`) | `UpdateLibraryWorkout` | `newUpdateWorkoutTool` | conditional | yes |
| 20 | `WorkoutDeleterClient` (`delete_workout.go`) | `DeleteLibraryWorkout` | `newDeleteWorkoutTool` | conditional; delete requirement inside tool | yes |
| 21 | `SportSettingsDeleterClient` (`delete_sport_settings.go`) | `DeleteSportSettings` | `newDeleteSportSettingsTool` | conditional; delete requirement inside tool | no |
| 22 | `CustomItemsClient` (`get_custom_items.go`) | `ListCustomItems`, `GetCustomItem` | `newGetCustomItemsTool`, `newGetCustomItemByIDTool`; dependency for create/update custom item validation | conditional; captured in local for later create/update constructors | yes |
| 23 | `CustomItemCreatorClient` (`create_custom_item.go`) | `CreateCustomItem` | `newCreateCustomItemTool` | conditional; receives captured `CustomItemsClient` possibly nil | no |
| 24 | `CustomItemUpdaterClient` (`update_custom_item.go`) | `UpdateCustomItem` | `newUpdateCustomItemTool` | conditional; receives captured `CustomItemsClient` possibly nil | no |
| 25 | `CustomItemDeleterClient` (`delete_custom_item.go`) | `GetCustomItem`, `DeleteCustomItem` | `newDeleteCustomItemTool` | conditional; delete requirement inside tool | no |
| 26 | `ActivityDetailsClient` (`get_activity_details.go`) | `GetActivity` | `newGetActivityDetailsTool`; optional collaborator for intervals/messages/link | conditional; optional collaborator for intervals/messages/link | yes |
| 27 | `ActivityDeleterClient` (`delete_activity.go`) | `GetActivity`, `DeleteActivity` | `newDeleteActivityTool` | conditional; delete requirement inside tool | no |
| 28 | `ActivityIntervalsClient` (`get_activity_details.go`) | `GetActivityIntervals` | `newGetActivityIntervalsTool`; collaborator for splits/extended metrics | conditional; captured local for splits | yes |
| 29 | `ActivityStreamsClient` (`get_activity_streams.go`) | `GetActivityStreams` | `newGetActivityStreamsTool`, `newGetActivitySplitsTool` | conditional; splits only when intervals client captured | yes |
| 30 | `ActivityMessagesClient` (`get_activity_messages.go`) | `GetActivityMessages` | `newGetActivityMessagesTool` | conditional; optional `ActivityDetailsClient` collaborator | yes |
| 31 | `ActivityMessageWriterClient` (`add_activity_message.go`) | `AddActivityMessage` | `newAddActivityMessageTool` | conditional | yes |
| 32 | `ExtendedMetricsClient` (`get_extended_metrics.go`) | `GetActivity`, `GetActivityIntervals`, `GetActivityPowerVsHR` | `newGetExtendedMetricsTool` | conditional | yes |
| 33 | `GearDeleterClient` (`delete_gear.go`) | `GetGear`, `DeleteGear` | `newDeleteGearTool` | conditional; delete requirement inside tool | no |

Generated schema catalog with current `schemaCatalogClient`/safe full toolset has 30 snapshots: add/update/read tools plus `delete_workout`, but omits 8 delete/custom-create-update tools that full production registration with `*intervals.Client` exposes in full delete mode: `delete_event`, `delete_events_by_date_range`, `delete_sport_settings`, `create_custom_item`, `update_custom_item`, `delete_custom_item`, `delete_activity`, `delete_gear`.

Verification evidence: temporary compile-time assertions assigning `(*intervals.Client)(nil)` to all 33 `XxxClient` interfaces passed with `go test ./internal/tools -run '^$'`.

Registry-level fake/call-site inventory:

- Per-tool constructor fakes (keep as unit-test seam; can avoid registry after signature change where tests only need one tool): `fakeApplyTrainingPlanClient`, `fakeDeleteToolsClient`, `fakeWorkoutUpdaterClient`, `fakeWorkoutLibraryClient`, `fakeActivitiesProfileClient`, `fakeEventWriterClient`, `fakeProfileClient`, `fakeWorkoutDeleterClient`, `fakeWorkoutCreatorClient`, `fakeLinkActivityToEventClient`, `fakeExtendedMetricsClient`, `fakeWellnessClient`, `fakeActivityReadClient`, `fakeEventsTrainingPlanClient`, `fakeCustomItemsClient`, `fakeActivityMessageWriterClient`, `fakeWellnessWriterClient`, `fakeFitnessMetricsClient`, `fakeSportSettingsWriterClient`.
- Registry/catalog fakes that exist specifically because `Register` accepts interface-shaped `any`: `fullCatalogTierClient` (`catalog_tiers_test.go`), `staticCatalogPanicClient` (`list_advanced_capabilities_test.go`), `advancedProtocolClient` (`internal/mcp/protocol_test.go`). Migration action: use a no-network dummy `*intervals.Client` for catalog/list-tools registration tests; for tests that need handler behavior, call the specific `newXxxTool(fake, ...)` constructor directly instead of relying on registry type assertions.
- Existing direct real-client registry fixture: `input_examples_test.go` already constructs `*intervals.Client` with dummy config, proving registration does not require network I/O.
- Non-tools call sites: `internal/app/app.go` already has a production `*intervals.Client`; MCP protocol tests have three real-registry fake call sites: `TestProtocolListAdvancedCapabilitiesVisibilityWithRealRegistry` uses `advancedProtocolClient` twice for list-tools only, and `TestProtocolGetAthleteProfileDispatch` uses `tools.NewRegistry(testProfileClient{...})` to exercise real MCP dispatch through the unexported `get_athlete_profile` tool. Migration action: advanced visibility tests can use a no-network dummy real client and update expected catalog/toolset behavior; profile dispatch should keep real registry behavior by constructing `*intervals.Client` against an `httptest.Server` fixture that returns the desired athlete profile, because package `mcp` cannot directly instantiate `newGetAthleteProfileTool`.

Schema membership inventory:

- Current `GenerateSchemaSnapshots()` count (verified with temporary `internal/toolchecks` test): 30 names, matching committed snapshot files: `add_activity_message`, `add_or_update_event`, `apply_training_plan`, `create_workout`, `delete_workout`, `get_activities`, `get_activity_details`, `get_activity_intervals`, `get_activity_messages`, `get_activity_splits`, `get_activity_streams`, `get_athlete_profile`, `get_best_efforts`, `get_custom_item_by_id`, `get_custom_items`, `get_event_by_id`, `get_events`, `get_extended_metrics`, `get_fitness`, `get_power_curves`, `get_training_plan`, `get_training_summary`, `get_wellness_data`, `get_workout_library`, `get_workouts_in_folder`, `icuvisor_list_advanced_capabilities`, `link_activity_to_event`, `update_sport_settings`, `update_wellness`, `update_workout`. `apply_training_plan` is present because `schemaCatalogClient` structurally satisfies embedded `ApplyTrainingPlanClient` through `WorkoutLibraryClient`, `EventsClient`, and `EventWriterClient`, even though there is no explicit compile-time assertion for `ApplyTrainingPlanClient`.
- Dummy real `*intervals.Client` full production registration count (temporary `internal/tools` test, mode full/toolset full): 38 names. Additional names versus schema snapshots: `create_custom_item`, `delete_activity`, `delete_custom_item`, `delete_event`, `delete_events_by_date_range`, `delete_gear`, `delete_sport_settings`, `update_custom_item`.
- Parity risk: replacing `schemaCatalogClient{}` with an unrestricted dummy `*intervals.Client` would add those 8 snapshots and break byte-identical schema stability. Step 3 must either keep a materially smaller schema fake/adapter that excludes those interfaces or add a schema-only registrar filter equivalent to the current catalog membership.
- Toolchecks migration must update both `GenerateSchemaSnapshots()` in `schema_stability.go` and `GenerateToolCatalog()` in `confusable_names.go`; both currently call `tools.NewRegistryWithOptions(schemaCatalogClient{}, ...)` and should share the same schema/catalog fixture or filter so schema snapshots and confusable-name checks see identical membership.

Special coupling/conditional semantics to preserve in refactor:

- `CustomItemsClient` is not just a read tool dependency; registry captures it and passes it to `newCreateCustomItemTool`/`newUpdateCustomItemTool` as schema validation collaborator. With typed `*intervals.Client`, pass the same client for both creator/updater and read collaborator so handlers never hit the existing `missing custom item creator or schema client` path in production.
- `get_activity_splits` currently registers only if both `ActivityStreamsClient` and the previously captured `ActivityIntervalsClient` exist. With `*intervals.Client`, both exist, so direct registration is safe and preserves production behavior; schema parity must still account for this tool already being in snapshots.
- `link_activity_to_event` currently passes optional `ActivityDetailsClient` and `EventByIDClient`; warnings are skipped when either is nil. With `*intervals.Client`, both collaborators become non-nil in production, matching the full-client behavior already possible via `fullCatalogTierClient`.
- `get_activity_intervals` and `get_activity_messages` also receive optional details collaborators today; direct full-client wiring makes those non-nil in production. Tests that depended on nil collaborators should instantiate the tool constructor directly.
- Capability/delete/toolset behavior is declared on each `Tool` (`Requirement`, `Toolset`) or passed into write schemas; registry currently does not filter tools itself before calling downstream registrar. Do not move this behavior into tool handlers.

| 2026-05-15 13:18 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 13:18 | Step 1 started | Map the current assertion chain |
| 2026-05-15 13:22 | Review R001 | plan Step 1: UNKNOWN |
| 2026-05-15 13:25 | Review R002 | plan Step 1: APPROVE |
| 2026-05-15 13:34 | Review R003 | code Step 1: UNKNOWN |
| 2026-05-15 13:39 | Review R004 | code Step 1: APPROVE |
| 2026-05-15 13:42 | Review R005 | plan Step 2: UNKNOWN |
| 2026-05-15 13:44 | Review R006 | plan Step 2: APPROVE |
| 2026-05-15 14:00 | Review R007 | code Step 2: UNKNOWN |
| 2026-05-15 14:07 | Review R008 | code Step 2: APPROVE |
| 2026-05-15 14:09 | Review R009 | plan Step 3: APPROVE |
| 2026-05-15 14:12 | Review R010 | code Step 3: APPROVE |
