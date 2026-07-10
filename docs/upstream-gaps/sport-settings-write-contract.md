# Sport-settings write contract

The live public OpenAPI document at `https://intervals.icu/api/v1/docs` was reconfirmed without credentials on 2026-07-10.

## Update

`PUT /api/v1/athlete/{athleteId}/sport-settings/{id}` requires the boolean query parameter `recalcHrZones`. The request body is the existing sparse `SportSettings` JSON object; it includes only the writable fields supplied by the caller.

The MCP `update_sport_settings` input exposes this as optional `recalc_hr_zones`. Omission resolves to `true`; an explicit `false` is preserved. The decoder uses presence-aware input so `false` is distinguishable from omission, then forwards the resolved value to `WriteSportSettingsParams` for query encoding.

`effective_date` is not part of the MCP request, examples, metadata, or generated schema. Strict decoding rejects it, like any other unknown argument, before a profile lookup or upstream request.

## Apply

`PUT /api/v1/athlete/{athleteId}/sport-settings/{id}/apply` takes no query parameters and no request body. It is a distinct explicit client operation and is not invoked by `UpdateSportSettings` or the MCP update tool.

The upstream operation is asynchronous and its public contract does not provide a date boundary. Consequently, icuvisor does not claim a date-scoped historical recomputation.

## Response metadata

The update response reports `hr_zone_recalculation_requested`, the boolean sent as `recalcHrZones`. This describes the requested update option only; it does not claim that activity recomputation is pending or complete. The former `effective_date` and `recompute_pending` metadata claims are removed.

## Regression boundary

Wire tests assert update method, path, sparse JSON body, and both resolved query values. Apply tests assert a bodyless, queryless PUT. Tool tests assert omitted/default-true and explicit-false forwarding, rejection of legacy `effective_date` before an upstream call, and no implicit apply path.
