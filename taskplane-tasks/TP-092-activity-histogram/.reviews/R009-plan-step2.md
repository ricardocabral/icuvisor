# R009 plan review — Step 2: Implement stream-backed histogram

Verdict: REVISE

The Step 1 contract is now detailed enough to implement from, but the Step 2 plan in `STATUS.md` is still too high-level for this blast radius. Please amend Step 2 with the concrete implementation shape before coding.

## Blocking plan gaps

1. **Make the histogram metric subset explicit in code-plan terms.**
   Current `internal/analysis/metrics.go` still does not expose the Step 1 stream-grain metrics `power_watts` / `heart_rate_bpm`, and the general `analysis_metric` enum contains many values that must not be accepted by this tool. The plan should state the exact helper(s) to add, e.g. a histogram-only parser/schema enum returning only `power_watts`, `heart_rate_bpm`, and `pace_seconds_per_km` plus accepted safe aliases. Do not let the tool schema reuse the full `MetricValues()` enum.

2. **Specify the stream request matrix and extraction path.**
   “Fetch only required streams” needs to be nailed down to avoid accidentally calling the public `get_activity_streams` handler or requesting broad defaults. Add the exact per-metric upstream request and canonical extraction plan:
   - power: `watts` + `time`
   - heart rate: `heart_rate` + `time`
   - pace: `distance` + `time`
   The plan should also say the tool calls `ActivityStreamsClient.GetActivityStreams` directly, uses canonical stream keys when reading returned rows, and never includes raw samples in the response.

3. **Clarify non-fatal details/profile lookup behavior.**
   Step 1 requires configured zones when available but fixed-width fallback when details/profile/zone settings are missing, and profile-fetch failure must not make pace unavailable (it defaults emitted units to `seconds_per_km`). The Step 2 plan should explicitly order/handle these calls so stream data can still produce a fixed-width histogram if activity details or athlete profile cannot be fetched. Only stream absence/invalid intervals should produce the structured `unavailable`/`insufficient_sample` payload.

4. **Define the package boundary for the engine vs. tool orchestration.**
   Add a concrete split such as: pure histogram construction, bucket assignment, fixed-width edges, zone-boundary conversion, rounding, and duration weighting in `internal/analysis/histogram*.go`; MCP argument decoding, stream/detail/profile fetching, user-facing unavailable payloads, and response shaping in `internal/tools/get_activity_histogram*.go`. This keeps the math testable without MCP fixtures and avoids importing `tools` from `analysis`.

5. **Include the unavailable/meta construction in Step 2.**
   The Step 2 checklist says “Return terse per-bucket summary only,” but the completion criteria also require structured missing-stream/insufficient-sample responses. Amend the plan to include `buckets: []`, `unavailable.reason/message`, `_meta.insufficient_sample:true`, omitted `_meta.bucket_method`, and accurate `_meta.n/source_tools/missing_days/missing_action` for no-valid-interval cases.

## Non-blocking reminders

- Preserve the Step 1 boundary semantics exactly: configured zones are lower bounds, interior buckets are `[lower, upper)`, the final bucket is open-ended, fixed-width uses exactly 10 raw-width buckets except identical min/max, and percentages are based on bucketed seconds only.
- Pace zone values must be converted from `MINS_KM`, `MINS_MILE`, `SECS_100M`, or `SECS_500M`; unknown/empty pace units with zones should fall back to fixed-width rather than guessing.
- Add at least pure engine tests during Step 2 if practical, even if broader tool/catalog/docs tests are reserved for Step 4.
