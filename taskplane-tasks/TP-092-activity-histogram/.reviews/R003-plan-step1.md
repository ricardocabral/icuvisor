# R003 Plan Review — Step 1: Define histogram contract

**Verdict:** APPROVE

The Step 1 contract is now concrete enough to implement and test. It names the histogram-only metric subset, defines the bucket row shape and rounding rules, locks `_meta.bucket_method` values, includes the TP-089 analyzer metadata, defines zone and fixed-width bucket construction, and specifies missing-stream/insufficient-sample behavior. The R001/R002 blocking ambiguities are resolved.

## Implementation notes to carry into Step 2

- Build the tool input schema from the histogram subset only: `power_watts`, `heart_rate_bpm`, and `pace_seconds_per_km`. Do not reuse the full `analysis.MetricSchemaProperty()` enum for this tool.
- When sorting configured zone boundaries, preserve the boundary-to-name relationship. If a boundary is filtered as non-finite, drop the corresponding name with it; pace zones may arrive in descending numeric order, so tests should cover label retention after sorting/conversion.
- Keep the top-level pace `metric` as the requested canonical `pace_seconds_per_km`; use bucket `unit` plus `_meta.emitted_unit` for imperial output, as the plan states.
- Treat `include_full` as metadata-only for this tool: no raw stream samples should appear in terse or full responses.
- Add targeted tests for the exact fixed-width edge rules, all-identical values, zone label fallback, missing/length-mismatched streams, and pace zone unit conversion.

No further Step 1 plan changes are required before implementation.
