# Plan Review — TP-009 Step 4

Decision: **Changes requested**

## Summary

I cannot approve Step 4 yet because the Step 4 implementation plan has not been recorded beyond the prompt checklist and the high-level Step 1 notes. `STATUS.md` does not yet define the concrete client methods, request/response contracts, heavy-stream gating behavior, split derivation algorithm, Strava-blocked fallback behavior, registry wiring, or targeted tests for `get_activity_streams` and `get_activity_splits`.

Step 1 identified the likely upstream streams endpoint and that splits are derived from intervals/streams, but Step 4 needs an implementation-level plan before coding. These two tools handle the heaviest payloads in this cluster and the highest risk of unit/sample-shape drift, so the plan should lock down the public contract first.

## Blocking issues

1. **Streams client/model contract is missing**
   - Specify the exact `internal/intervals` method and params to add, e.g. `GetActivityStreams(ctx, activityID, params)` against `GET /api/v1/activity/{id}/streams` with `types` and `includeDefaults` query handling.
   - Define the `ActivityStream` upstream struct, including raw preservation for `include_full` and tolerant decoding for `data`, `data2`, `valueTypeIsArray`, `anomalies`, `custom`, and `allNull`.
   - Decide how stream sample arrays are represented internally (`[]float64`, `[]any`, paired lat/lng arrays, nullable samples, etc.) so canonicalization and split derivation can be tested without lossy decoding.
   - Document input validation and error mapping for blank `activity_id`, invalid keys, upstream 4xx/5xx, and context cancellation.

2. **Public `get_activity_streams` request/response shape is undefined**
   - Define the input schema: `activity_id`, `include_full`, `keys`, and any `include_defaults`/metadata-only option. Each argument needs an LLM-readable JSON Schema description.
   - Define default behavior precisely. The prompt requires that default calls do **not** return full stream samples; the plan should say whether default returns only available stream names, counts, metadata, or a short summary.
   - Define behavior when `keys` is supplied but `include_full` is false. The prompt allows explicit `keys` as an opt-in to requested channels, but the response still needs a bounded shape (samples included for those keys, or summaries only?).
   - Define the wrapper shape (`activity_id`, `streams`, `_meta`) and `_meta` fields, including `server_version`, `include_full`, requested keys, returned keys, and `unknown_stream_keys`.
   - Specify how `include_full:true` preserves the raw upstream stream objects and nulls while terse/default mode goes through TP-007 response shaping.

3. **Stream-key canonicalization plan is not concrete enough**
   - The plan must state that both requested `keys` and upstream `ActivityStream.type` are canonicalized via TP-008 `streams.CanonicalKey`.
   - Define the alias-to-upstream-token mapping used for the `types` query. For example, if callers request `heart_rate`, decide whether the upstream token is `hr`, `HeartRate`, or pass-through, and add tests for known aliases.
   - Define collision behavior when multiple upstream stream types canonicalize to the same key, and how that is surfaced in `_meta`.
   - Define exactly which unknowns are reported in `_meta.unknown_stream_keys`: unknown requested keys, unknown upstream types, or both, and whether the metadata uses original or canonicalized names.

4. **Splits algorithm is missing**
   - Define how manual laps are detected from `get_activity_intervals` data. The plan should specify which interval `type`, `unit`, or distance/duration fields qualify as manual laps and which interval rows should be ignored.
   - Define the virtual split algorithm when manual laps are absent: required streams, distance units expected from upstream, time axis source, interpolation at split boundaries, elapsed vs moving time, and how to treat paused/non-moving samples.
   - Define edge cases and output behavior for missing distance/time streams, non-monotonic distance, duplicate timestamps, null samples, activity shorter than one split, partial final split, GPS gaps, and all-null streams.
   - Define the split output schema with unit-disambiguated fields, e.g. `split_index`, `distance_km`/`distance_mi`, `elapsed_time_seconds`, `moving_time_seconds`, `pace_seconds_per_km`/`pace_seconds_per_mile`, HR/power aggregates if included, `source: manual|virtual`, and `_meta` unit metadata.
   - Document where the required package doc will live and what algorithm/edge cases it will cover.

5. **Preferred-units and profile dependencies are not planned**
   - `get_activity_splits` must fetch the athlete profile to honor `preferred_units`; the plan should specify reuse of the existing `ProfileClient`, unit fallback behavior, and preservation of context cancellation errors.
   - Define any optional caller override such as `split_unit` and validation rules (`km`/`mile` only), including how it interacts with athlete `preferred_units`.
   - Make field names unit-safe and consistent with Step 2 (`distance_km`/`distance_mi`, `pace_seconds_per_km`/`pace_seconds_per_mile`).

6. **Strava-blocked behavior is not specified for streams/splits**
   - The mission requires structured Strava-blocked detection across the activity read cluster. The plan should define how `get_activity_streams` and `get_activity_splits` return `unavailable: { reason: "strava_tos", workaround: ... }` for hidden/stub activities instead of propagating sparse rows or upstream 4xxs.
   - For streams/splits endpoints that may return not-found/forbidden on blocked activities, specify the fallback detail-read confirmation path, as was done for `get_activity_intervals`, so genuine credential failures are not masked as Strava.

7. **Registry and targeted test plan are absent**
   - Define the new tool/client interfaces and how `NewRegistry` registers `get_activity_streams` and `get_activity_splits` when the configured intervals client supports the needed methods.
   - Add focused Step 4 tests rather than deferring validation to Step 6: intervals client path/query tests, stream heavy-payload default behavior, explicit-key behavior, `include_full` raw/null preservation, canonicalization and unknown-key metadata, collision behavior if supported, metric and imperial virtual splits, manual-lap precedence, missing/paused sample edge cases, Strava-unavailable fallback, registry/schema exposure, and context-cancellation preservation.
   - Include README and CHANGELOG updates for the two new tools in the implementation plan.

## Required updates before approval

Please update `STATUS.md` with a `Step 4 Notes` section covering:

- Concrete endpoint/client/type changes for streams and the dependency graph for splits.
- Exact request and response schemas for both tools, including default vs `keys` vs `include_full` behavior.
- Stream-key canonicalization, alias mapping, collision handling, and unknown-key metadata.
- The manual-lap and virtual-split algorithms, including unit handling and edge cases.
- Strava-blocked handling and fallback detail-read confirmation for both tools.
- Registry/app wiring changes and a focused Step 4 test checklist.

## Non-blocking status hygiene

`STATUS.md` still has review-log rows after `_None._` under `## Blockers`. Move or remove those rows so the Blockers section remains unambiguous before task wrap-up.
