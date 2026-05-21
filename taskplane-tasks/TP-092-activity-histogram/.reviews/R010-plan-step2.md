# R010 plan review — Step 2: Implement stream-backed histogram

Verdict: APPROVE

The amended Step 2 plan resolves the R009 blockers and is concrete enough to implement. It now pins the histogram-only metric subset/schema, the exact stream request matrix, direct `ActivityStreamsClient.GetActivityStreams` usage with `IncludeDefaults:false`, the engine/tool package boundary, non-fatal details/profile lookup behavior, and the structured unavailable/meta payload for missing or invalid stream intervals.

## Approval notes

- Keep the tool schema on `analysis.HistogramMetricSchemaProperty()`/histogram subset only. Do not reuse the full `MetricValues()` enum or accept aggregate metrics such as `average_power_watts` / `average_heart_rate_bpm` for this tool.
- Preserve the direct stream-client path from the plan: request only `watts,time`, `heart_rate,time`, or `distance,time`, canonicalize returned row keys, and never call the public `get_activity_streams` handler or include raw samples in either terse or full responses.
- Treat activity details/profile fetches as best-effort exactly as planned: context cancellation/deadline should still propagate, but ordinary lookup failures should only add warnings and fall back to fixed-width buckets; profile failure must not make pace unavailable and should default to `seconds_per_km`.
- Keep pure histogram math in `internal/analysis` so fixed edges, zone boundaries, pace conversion, interval weighting, rounding, and bucket assignment are unit-testable without MCP/tool fixtures.
- For unavailable responses, omit `_meta.bucket_method` and include `buckets: []`, `unavailable.reason/message`, analyzer meta (`method`, `source_tools`, `n`, `missing_days`, `missing_action`, `insufficient_sample`), and no raw stream data.

## Non-blocking implementation reminders

- Add pure engine tests during this step if practical; Step 4 can still add broader tool/catalog/docs coverage.
- Be careful not to log raw athlete identifiers or upstream payloads when recording lookup warnings.
- Keep `source_tools` accurate: include `get_activity_details` / `get_athlete_profile` when those lookups are attempted or used, and rely on the analyzer meta normalizer for stable ordering/deduplication.
