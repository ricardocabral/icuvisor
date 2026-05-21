# R001 Plan Review — Step 1: Define histogram contract

**Verdict:** Request changes before implementation.

The Step 1 plan currently restates the task checkboxes, but it does not define the public contract yet. For this task the contract is the hard part: it affects the MCP schema, response shape, analyzer metadata, catalog docs, and golden tests. Please make the plan concrete before moving to Step 2.

## Blocking gaps

1. **Metric enum is ambiguous.**
   - The task says to limit this to power/HR/pace through the closed metric enum, but the current `internal/analysis` enum does not obviously contain raw stream metrics named `power`, `heart_rate`, or `pace`; it mostly has activity-row metrics like `average_heart_rate_bpm`, `average_power_watts`, and `pace_seconds_per_km`.
   - The plan must name the exact accepted canonical values and whether this requires extending `internal/analysis` with stream-grain metrics, or using a histogram-specific closed subset derived from `analysis.Metric`.
   - Do not allow free-form strings like `metric: "power"` unless they are explicitly part of a closed enum/alias contract and the schema enumerates canonical values only.

2. **Bucket row shape is not precise enough.**
   - “label/range, seconds, percentage, unit” needs exact JSON field names and types.
   - Define range bounds and inclusivity, including open-ended first/last buckets. For example, decide whether rows use `lower`, `upper`, `lower_inclusive`, `upper_exclusive`, `seconds`/`duration_seconds`, `percentage`, and `unit`.
   - Define percentage denominator and rounding expectations so tests can lock behavior.

3. **`_meta.bucket_method` needs stable enum values and companion metadata.**
   - Name the exact allowed values, e.g. `configured_zones` vs `fixed_width` (or whatever final names are chosen). Avoid prose values that will drift.
   - The contract should also account for the mandatory analyzer metadata from TP-089: `_meta.method`, `_meta.source_tools`, `_meta.n`, `_meta.missing_days`, `_meta.missing_action`, and `_meta.insufficient_sample` where applicable. Step 3 mentions some of this, but it is part of the response contract and should be decided in Step 1.
   - For fixed-width fallback, define how the chosen edges/width are exposed, probably in `_meta`, so the result is auditable.

4. **Pace needs an explicit contract.**
   - Pace histograms are easy to get wrong because lower seconds means faster. Define the canonical unit (`seconds_per_km`, `seconds_per_mile`, etc.), how preferred units affect output, and whether bucket ordering is fast-to-slow or slow-to-fast.
   - Define how athlete pace zones from `SportSettings.PaceZones`/`PaceUnits` map to output ranges and labels, including conversions from upstream pace units.

5. **Zone source selection is not specified.**
   - The tool must prefer athlete configured zones, but the plan does not say how the selected sport setting is determined for a single activity. Define whether the activity details are fetched to determine sport/type, how that maps to `SportSettings.Type`/`Types`, and what metadata reports the selected sport/zone source.
   - Also define what happens when profile/sport settings are missing or the selected metric has no configured zones.

6. **Missing stream / insufficient sample response shape is missing.**
   - The task completion criteria require structured terse errors/meta for missing streams or insufficient samples. Step 1 should define whether the tool returns an `unavailable` object, empty `buckets`, `insufficient_sample: true`, or a user error for each case.
   - Define `n` semantics now: valid samples, valid intervals, or total seconds contributing to buckets.

## Recommendation

Revise Step 1 into a short contract section in `STATUS.md` (or an implementation note referenced from it) that includes:

- input schema fields and exact accepted metric enum values;
- output JSON shape with exact bucket field names/types;
- stable `_meta.bucket_method` enum values plus analyzer metadata fields;
- range/inclusivity, percentage, ordering, and pace-unit semantics;
- zone-source and fixed-width-fallback metadata;
- missing-stream/insufficient-sample response shape.

Once those details are explicit, the implementation and tests in later steps will have a stable target.
