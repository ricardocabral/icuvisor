# R002 Plan Review — Step 1: Define histogram contract

**Verdict:** REVISE

The updated Step 1 notes address most of R001: the input enum is now named, the bucket row shape is concrete, analyzer `_meta` is included, and missing-stream/unavailable responses are described. However, this is a public math contract and a few details remain underspecified enough that Step 2 could implement incompatible behavior and still appear to satisfy the plan.

## Blocking gaps

1. **Configured-zone bucket construction is still ambiguous.**
   - The plan says to use `PowerZones`/`HRZones`/`PaceZones` plus zone names when at least one boundary exists, but it does not define how an upstream boundary list maps to bucket ranges.
   - Please specify whether boundaries are lower bounds, upper bounds, or interior cut points; whether a below-first and/or above-last bucket is emitted; and how labels are assigned when `ZoneNames` length equals, exceeds, or is shorter than the bucket count.
   - This matters because existing sport-settings schema only says “ordered zone boundary values” and names match boundary length, not necessarily `len(boundaries)+1`.

2. **Fixed-width fallback lacks deterministic edges.**
   - `_meta.fixed_width` reports `min`, `max`, `bucket_count`, `width`, and `unit`, but the plan does not define how `bucket_count` and `width` are chosen.
   - Please lock a deterministic rule, e.g. fixed default bucket count, min/max expansion, “nice number” vs raw width, and behavior when all valid values are identical. Tests need exact expected edges.

3. **Duration/interval semantics need one more sentence.**
   - `seconds` and `_meta.n` are defined as valid stream intervals, but the plan does not state how power/HR intervals are derived: from adjacent `time` samples, one second per sample, or another stream-derived cadence.
   - Please define the required stream keys per metric and the fallback/unavailable behavior when the value stream exists but timing data needed for duration weighting is missing. Pace already mentions adjacent distance/time samples; power and HR need the same level of clarity.

4. **Pace response identity should be explicit.**
   - The accepted canonical pace metric is `pace_seconds_per_km`, while the plan says imperial athletes may emit `seconds_per_mile` buckets. State whether the top-level `metric` remains the requested canonical metric or changes to the emitted unit-specific metric, and ensure the schema/response examples will not imply both.

## Non-blocking recommendations

- Consider adding `_meta.zone_source.boundaries` or `_meta.zone_source.boundary_unit` if the emitted bucket ranges are converted from upstream pace units; this will make zone-based results as auditable as fixed-width results.
- If details/profile fetch errors are intentionally non-fatal fallbacks to fixed-width buckets, explicitly call that out as “fetch error => fallback” rather than only “missing settings => fallback.”

Once these contract points are added to `STATUS.md`, the plan should be stable enough for implementation and golden tests.
