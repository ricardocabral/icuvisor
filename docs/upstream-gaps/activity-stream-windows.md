# Upstream gap: activity stream windows

## Status

Intervals.icu does not have a verified public query contract for selecting a time or distance window from `GET /api/v1/activity/{id}/streams{ext}`. icuvisor therefore performs bounded slicing locally after the existing full stream fetch. No `time_window`, `distance_window`, or equivalent parameter is sent upstream.

## Evidence

The repository-held OpenAPI baseline (`scripts/openapidiff/baseline/intervals-openapi.json`, `GET /api/v1/activity/{id}/streams{ext}`) documents only these query parameters:

- `types` — requested stream names;
- `includeDefaults` — include default streams in addition to `types`.

The endpoint response is an array of `ActivityStream` objects. The baseline does not document elapsed-time bounds, distance bounds, server-side slicing, or a reduced response contract. The client test `internal/intervals/activity_streams_test.go` verifies that icuvisor sends only `types` and `includeDefaults`.

## Local contract and cost

`get_activity_streams` accepts inclusive `time_window` bounds in elapsed seconds and `distance_window` bounds in meters. The handler requests canonical `time` and/or `distance` helper streams when explicit `keys` omit them, builds one common index mask, and applies it to every aligned channel and `data2`. `include_full:true` is still required for sample arrays and raw payloads; `max_points` remains gated by that flag. Window provenance reports requested/effective bounds, source/selected/returned counts, boundary units, and the sampling method. Invalid or unavailable boundaries and incompatible channels are diagnosed rather than shifted, interpolated, or filled with fabricated values.

Because the upstream request is still a full fetch, a windowed call has the same upstream bandwidth, server work, and memory cost as an unwindowed call. Local slicing only bounds the response sent to the MCP client. Very large activities can still be expensive, and an upstream future window API must not be adopted until its public parameters, units, inclusivity, and alignment semantics are verified.

This limitation is intentionally documented instead of advertising unsupported upstream parameters.
