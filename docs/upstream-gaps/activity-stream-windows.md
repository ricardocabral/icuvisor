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

Explicit stream-key requests and split requests omit `includeDefaults` (whose upstream default is false), so unrelated default channels cannot interfere with those reads. Unfiltered stream reads retain default channels. Nonnumeric or malformed `data`/`data2` arrays are withheld with `channel_decode_failed` diagnostics while valid channels remain available; malformed anomaly metadata does not discard numeric samples. Fetch failures report safe HTTP, decoding, size-limit, or transport diagnostics without exposing raw upstream payloads or request URLs.

Because the upstream request is still a full fetch, a windowed call has the same upstream bandwidth, server work, and memory cost as an unwindowed call. Local slicing only bounds the response sent to the MCP client. Very large activities can still be expensive, and an upstream future window API must not be adopted until its public parameters, units, inclusivity, and alignment semantics are verified.

This limitation is intentionally documented instead of advertising unsupported upstream parameters.

## Authenticated regression evidence: issue 64

On 2026-09-30, a synthetic TCX run uploaded to the test account reproduced the remaining v1.7.1 failures against the real upstream API. The recording contains 2,858 one-second samples, distance from 5 to 7,400 m, heart rate, power, cadence, altitude, and 24 paused seconds. Every returned scalar stream includes `data2:null` with `valueTypeIsArray:false`; this is an absent secondary channel, not corrupt primary samples. The captured response is `internal/tools/testdata/activity_streams/issue64_nonzero_paused.json`.

Before the fix, direct unfiltered/filtered streams returned `channel_null` for every scalar channel, windowed reads returned `window_channel_null`, and splits returned no rows with `paused_samples_present` and `insufficient_split_coverage`. The controls succeeded on the same upstream data: heart-rate histogram n=2857, pace histogram n=2833, and segment mean over 0–120 seconds n=121. These counts match the reported issue; the sensor values are invented and the recording is not the reporter's private activity.

Scalar null secondary channels are now treated as absent throughout unbounded, bounded, and sampled responses. Explicitly paired channels (`latlng` or `valueTypeIsArray:true`) still reject a null secondary channel, and non-null malformed, mismatched, or null-containing secondary arrays remain invalid. Full responses preserve the upstream null instead of inventing a secondary sample array.

Virtual splits use absolute cumulative distance boundaries. A recording starting at 5 m and ending at 7,400 m supports kilometres 2–7; the first kilometre is omitted with `initial_split_unavailable`. Indices retain cumulative kilometre positions, durations include recorded elapsed pause time, and missing origin/tail data is not extrapolated or renumbered as a complete first split.
