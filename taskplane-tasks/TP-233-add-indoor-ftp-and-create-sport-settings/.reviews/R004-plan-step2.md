# Plan Review — TP-233 Step 2

## Verdict: REVISE

The plan has the intended high-level behavior, but needs the following implementation/test details before it is actionable:

1. **Specify the create validation/call ordering.** Decode strictly, canonicalize and validate `sport` against `supportedSportSettingsSports`, and require a non-nil threshold field *before* reading the profile. Those structural failures must make zero profile and create calls. A duplicate can only be discovered after one profile read: use `findSportSetting` (covering both `Type` and `Types`), return an actionable “use update_sport_settings” error, and assert zero `CreateSportSettings`/POST calls rather than incorrectly asserting zero network calls. The request type and schema must exclude `recalc_hr_zones`, zones, confirmation, and credentials.

2. **Name the complete catalog/snapshot integration.** Add `CreateSportSettings` to the athlete-scoped canonical catalog, register `newCreateSportSettingsTool` in `registryBaseTools`, add it to the full-tier expectation, and add it to the `settings` case in `internal/tools/catalog.go`'s `toolCatalogGroup`. The latter is not covered by merely adding the `internal/toolcatalog` constant; omitting it makes `Catalog()` return an invalid empty group. Update the schema-registry expected count from 69 to 70, require the new snapshot name, and generate the committed snapshot with `go run ./scripts/snapshot_tool_schemas.go`.

3. **Make the public echo changes concrete.** For update, state the exact sparse request/parameter/echo/meta path: `indoor_ftp` → `WriteSportSettingsParams.IndoorFTP`, `indoor_ftp_watts` from the returned value with supplied-value fallback, and sorted `_meta.fields_updated: ["indoor_ftp"]`; also update the tool description, input schema, and an example. Define the create confirmation's exact `sport_settings` and `_meta` fields (including created ID, only relevant threshold echoes, m/s-derived pace rendering, selected pace units/load type, and `operation: "create"`) and add focused handler tests for its FTP, indoor FTP, HR, and Run/Swim pace paths.

These details also preserve the intended profile read for duplicate detection while preventing malformed calls from causing any I/O or exposing a zone/recalculation creation path.
