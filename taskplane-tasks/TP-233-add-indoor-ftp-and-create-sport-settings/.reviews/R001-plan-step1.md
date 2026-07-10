# Plan Review — TP-233 Step 1

## Verdict: REVISE

No Step 1 implementation plan was submitted. `STATUS.md` only repeats the unchecked checkpoint list, which does not define the client boundary or the wire/validation tests needed for this new non-idempotent create operation.

Revise the plan to specify:

1. **The exact typed boundary.** Add sparse `IndoorFTP *int` to `WriteSportSettingsParams`, and define a separate create-only parameter type containing the sport plus only `FTP`, `IndoorFTP`, `ThresholdHR`, and canonical `ThresholdPace`. `CreateSportSettings` must return `SportSettings`. Do not reuse a type that permits `SportSettingID`, `RecalcHRZones`, or `Zones` on POST. State that an already-normalized `SportSettingsPace` carries m/s `Value` plus the selected `PaceUnits` and explicit/preserved `PaceLoadType`; the client serializes it and does not reinterpret pace units.

2. **The complete HTTP contract.** `UpdateSportSettings` remains `PUT /athlete/{athleteID}/sport-settings/{id}?recalcHrZones=<resolved boolean>` and includes `indoor_ftp` only when supplied. `CreateSportSettings` is exactly `POST /athlete/{athleteID}/sport-settings`, with a sparse JSON body whose required creation discriminator is `types: [sport]` (not `type`), no query string, and no `recalcHrZones` or zone fields. Define the threshold keys precisely: `ftp`, `indoor_ftp`, `lthr`, `threshold_pace` (m/s), `pace_units`, and `pace_load_type`. Preserve the normal POST no-retry behavior rather than making creation retryable without an upstream idempotency contract.

3. **Validation and its location.** Before any request, reject missing/blank sport on creation; non-positive FTP, indoor FTP, and threshold HR; and non-finite/non-positive canonical threshold pace. State whether client sport validation is restricted to the existing documented sport enum or leaves upstream sport values to the MCP layer, and use that choice consistently. There must be no `indoor_ftp <= ftp` rule: the confirmed upstream contract/product rules do not establish one. Invalid parameters must make no HTTP call and errors must identify the failing create/update operation.

4. **Focused client regression coverage.** Add local-server exact-wire cases for an update containing only `indoor_ftp` and a create containing `types:["Ride"]` plus indoor FTP. Assert method, path, raw query, sparse body, absence of zone/recalculation fields, and decoding of the returned `indoor_ftp` echo. Add creation threshold-pace coverage proving m/s plus `pace_units`/`pace_load_type` serialization, and table-driven invalid create/update parameter cases that prove zero requests. The later Ride/Run/Swim matrix can expand this coverage, but Step 1 must already lock the client contract it introduces.
