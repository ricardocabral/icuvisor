# Sport-settings write contract

The live public OpenAPI document at `https://intervals.icu/api/v1/docs` was reconfirmed without credentials on 2026-07-10.

## Update

`PUT /api/v1/athlete/{athleteId}/sport-settings/{id}` requires the boolean query parameter `recalcHrZones`. The request body is the existing sparse `SportSettings` JSON object; it includes only the writable fields supplied by the caller.

The MCP `update_sport_settings` input exposes this as optional `recalc_hr_zones`. Omission resolves to `true`; an explicit `false` is preserved. The decoder uses presence-aware input so `false` is distinguishable from omission, then forwards the resolved value to `WriteSportSettingsParams` for query encoding.

`effective_date` is not part of the MCP request, examples, metadata, or generated schema. Strict decoding rejects it, like any other unknown argument, before a profile lookup or upstream request.

## Apply

`PUT /api/v1/athlete/{athleteId}/sport-settings/{id}/apply` takes no query parameters and no request body. It is a distinct explicit client operation and is not invoked by `UpdateSportSettings` or the MCP update tool.

The upstream operation is asynchronous and its public contract does not provide a date boundary. Consequently, icuvisor does not claim a date-scoped historical recomputation.

## Client implementation boundary

`WriteSportSettingsParams` carries a resolved `RecalcHRZones bool`. The update client encodes it as the required `recalcHrZones` query key with `strconv.FormatBool`; the sparse JSON body never contains that option. The MCP decoder, rather than a client zero-value heuristic, resolves omission to true before constructing these parameters.

`ApplySportSettings` takes only `(ctx, sportSettingID)` so callers cannot pass a date. Its transport uses a bodyless PUT helper that creates a fresh `http.Request` with a nil body for each retry, retains the existing retry/error behavior, and closes each response body. The update path uses a body-plus-query helper with the same retry and error semantics. Neither helper changes existing callers. `UpdateSportSettings` does not call apply; an internal `EffectiveDate`, if temporarily retained during migration, has no transport effect and is removed with the MCP producer.

## Response metadata

The update response reports `hr_zone_recalculation_requested`, the boolean sent as `recalcHrZones`. This describes the requested update option only; it does not claim that activity recomputation is pending or complete. The former `effective_date` and `recompute_pending` metadata claims are removed.

## Schema migration and stability approval

The request uses `*bool` while decoding to preserve whether `recalc_hr_zones` was supplied. Decoding resolves nil to true and copies that resolved value to `WriteSportSettingsParams.RecalcHRZones`; `EffectiveDate` is removed from both types. The input schema requires only `sport`, exposes optional boolean `recalc_hr_zones` with `default: true`, and uses examples without dates.

The response `_meta` always emits `hr_zone_recalculation_requested`, exactly the resolved option. It retains delete-mode and unit metadata but removes `effective_date` and `recompute_pending`.

Schema stability retains its generic property-removal protection. Production `internal/toolchecks/schema_stability.go` holds an `approvedSchemaPropertyRemovals` policy keyed by tool and property. Its sole entry is `update_sport_settings.effective_date`, with the TP-228 safety-correction rationale that the field falsely implied date-scoped upstream recomputation. `compareStableSchema` consults that policy only before emitting a `property-removed` failure. Tests exercise `CheckSchemaStability` itself: the one approved removal passes, while another `update_sport_settings` property and `effective_date` on a different tool fail. The schema snapshot and generated website data are refreshed with `make docs-tools`.

## Regression boundary

Wire tests assert update method, path, sparse JSON body, and both resolved query values. Apply tests assert a bodyless, queryless PUT. Tool tests assert omitted/default-true and explicit-false forwarding, rejection of legacy `effective_date` before an upstream call, and no implicit apply path.
