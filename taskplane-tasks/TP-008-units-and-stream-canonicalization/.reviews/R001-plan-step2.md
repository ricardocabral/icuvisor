# Plan Review: TP-008 Step 2 — response-boundary unit conversion

## Verdict: approved

The revised Step 2 plan is concrete enough to implement. It addresses the prior review concerns by keeping intervals decode verbatim, locating preferred-unit conversion in `internal/response`, defining a typed converter result that preserves original values/units, explicitly protecting swim/row pace units from generic mile/km conversion, naming `get_athlete_profile` as the response-boundary integration point, and adding Step 2-specific tests plus the `CHANGELOG.md` update.

## What looks good

- Decode preservation is explicitly verified/tested: intervals structs keep upstream unit tokens as strings such as `MINS_KM` rather than converting during JSON decode.
- The conversion layer is correctly planned for `internal/response`, not `internal/units`, preserving the separation between upstream enum parsing and athlete-preferred response shaping.
- The converter result is planned to carry both preferred and original value/unit data, which is necessary to avoid lossy shaping and to keep responses self-describing.
- The sport-specific pace policy is now called out: `MINS_KM`/`MINS_MILE` can convert according to athlete preference, while `SECS_100M` and `SECS_500M` pass through unless a future sport-aware rule is added.
- Existing `get_athlete_profile` pace shaping will be migrated to exercise the boundary behavior now rather than adding unused helpers.
- The tests listed are appropriate for this step: distance, speed, run/walk pace, swim/row pass-through, unknown raw metadata, profile integration, and response-owned `_meta.units` stripping/emission.
- The plan includes the user-visible documentation update under `[Unreleased]`.

## Implementation guardrails

- Make the raw unknown label explicit in the converter API, either by accepting a parsed-unit struct or a separate raw label parameter. `UnitUnknown` alone cannot recover the upstream token needed for `_meta.unknown_unit: <value>`.
- Document the numeric convention in code/tests: profile pace values are emitted as seconds per distance base, so `MINS_KM`/`MINS_MILE` conversion should preserve duration semantics while changing the distance denominator.
- For sport-specific pace pass-through, ensure the field suffix and `_meta.units` label cannot imply km/mi. If `get_athlete_profile` receives `SECS_100M` or `SECS_500M`, avoid reusing `*_seconds_per_km` / `*_seconds_per_mile` field names for those values.
- Keep caller-supplied `_meta.units` stripped in `response.Shape`; only the response boundary should add canonical units metadata.
- Run targeted tests for `internal/response` and `internal/tools` at the step boundary, even if the full `make test` / `make lint` / `make build` gate remains in Step 4.
