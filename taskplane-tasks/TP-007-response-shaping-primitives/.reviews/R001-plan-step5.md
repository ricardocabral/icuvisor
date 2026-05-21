# Plan Review — TP-007 Step 5

Verdict: **Not approved yet**. The Step 5 section in `STATUS.md` currently only repeats the prompt checklist. This step has several contract-setting decisions (timezone rendering formats, unit metadata shape, `preferred_units` fallback semantics, and metadata ownership) that need to be recorded before coding so downstream read tools can use the primitives consistently.

## Blocking findings

1. **Unit-system API and source of truth are unspecified.**
   - The plan does not name where `UnitSystem` will live, its exported constants, or how callers construct it from an athlete profile.
   - PRD §7.4 #11 specifically calls out `preferred_units`, but `intervals.AthleteWithSportSettings` currently only has `measurement_preference`, `weight_pref_lb`, and `fahrenheit`. The plan must state whether Step 5 adds a `PreferredUnits` field, how it maps known upstream values to metric/imperial, and what fallback is used when `preferred_units` is absent.
   - Unknown/empty unit values need explicit behavior. They should not crash read tools; the plan should define whether they fall back to metric/config/profile booleans and whether `_meta.units` records the source/unknown value.

2. **`_meta.units` shape and ownership are not defined.**
   - Step 5 requires surfacing the active unit system in `_meta.units`, but the current response shaper only owns `server_version`, debug metadata, strip metadata, and scales.
   - The plan must define the exact JSON shape, for example whether `_meta.units` is `{ "system": "metric", "distance": "km", ... }`, whether it includes conversion source, and whether it is response-level only or also added to shaped row-collection rows.
   - Merge/collision behavior is missing. Since `internal/response` owns metadata assembly, caller-supplied `_meta.units` should either be reserved and overwritten by computed values or explicitly merged in a deterministic way while preserving unrelated `_meta` keys.

3. **Distance key/conversion helpers are underspecified.**
   - The checklist says helpers should choose `distance_km` vs `distance_mi` field names and convert values, but the plan does not state base-unit assumptions (meters from intervals.icu? kilometers?), rounding/precision policy, or names/signatures for helpers.
   - The plan should distinguish this Step 5 framework from TP-008's exhaustive upstream unit enum work. Do not introduce stream-key/unit-enum canonicalization here; provide narrow helpers that downstream tools can call for common distance/pace presentation.
   - Existing `get_athlete_profile` has tool-local unit/pace helpers (`normalizedMeasurementPreference`, `isMilePace`, etc.). The plan should say which of these move into `internal/response` or `internal/config`, and which remain profile-specific until Step 6.

4. **Timezone rendering convention is not concrete enough.**
   - "Render date/time fields in the athlete's configured TZ" needs an API and exact output formats. The plan should state whether date-times render as RFC3339 with the athlete offset, whether date-only values remain `YYYY-MM-DD`, and how DST is handled.
   - The shaper cannot safely infer arbitrary date strings after JSON round-tripping. The plan should explicitly make timezone rendering a presentation-boundary helper that tools call on `time.Time` values before `response.Shape`, not a recursive string mutation pass.
   - Invalid timezone behavior/fallback should be defined. Config already validates configured timezones; athlete-profile timezones may still be empty or invalid, so the helper should have deterministic fallback/error semantics and tests.
   - Update `internal/response/doc.go` in the plan with the convention so future tools know to convert UTC boundary times to athlete-local strings before shaping.

5. **Athlete-ID centralization needs specific cleanup semantics.**
   - `config.NormalizeAthleteID` already exists, but `internal/tools/get_athlete_profile.go` also has `normalizeProfileAthleteID` that silently returns the raw trimmed value on invalid input. The plan must identify this duplicate and define whether callers should propagate errors, omit invalid IDs, or deliberately preserve raw upstream IDs with metadata.
   - For future coach-mode/read-tool arguments, the plan should state that all accepted athlete IDs go through `config.NormalizeAthleteID` and all emitted IDs use the normalized `i12345` form. Tests should cover both input forms and invalid values at the helper boundary.

6. **Integration with the existing `response.Shape` pipeline is not planned.**
   - Step 5 should say how unit metadata is passed into the single chokepoint, likely via new fields on `response.Options`, and where it is added relative to null stripping, debug metadata, scale metadata, and `server_version`.
   - Include-full semantics should be explicit: `include_full` preserves raw nulls, but common metadata (`server_version`, units) should still be injected when known.
   - If no active unit system is known, the plan should define whether `_meta.units` is omitted or set to an explicit unknown/default object.

7. **Focused tests are not planned at the step boundary.**
   - Step 7 has broad test bullets, but this step should still plan concrete tests before coding.
   - Needed cases include: `preferred_units` decoding from the intervals profile; fallback from legacy `measurement_preference`/`weight_pref_lb`; unknown unit behavior; `DistanceField`/conversion helper outputs for metric and imperial; `_meta.units` injection and collision handling; include-full still includes common units metadata; athlete-ID normalization for `12345`/`i12345` and invalid IDs; timezone rendering across a non-UTC location/DST boundary; and invalid/empty athlete timezone fallback.

## Required additions before approval

Please update `STATUS.md` with a Step 5 implementation plan covering at least:

- The exact `UnitSystem` representation/API, its package location, constants, and construction from `preferred_units` with documented fallbacks.
- The exact `_meta.units` JSON shape, response-level vs row-level placement, and merge/overwrite behavior for existing `_meta.units`.
- The distance/pace helper signatures, input-unit assumptions, conversion precision, and a note that TP-008's exhaustive unit enum/stream canonicalization remains out of scope.
- The timezone rendering helper API, output formats, fallback/error behavior, and package-doc update.
- The athlete-ID cleanup plan, including removal or replacement of tool-local duplicate normalization and invalid-upstream-ID semantics.
- How `response.Shape` receives and injects units in the established operation order, including `include_full` behavior.
- Focused tests for the unit, timezone, athlete-ID, and metadata cases above.

Once those decisions are recorded, Step 5 should be implementable without surprising downstream tool authors or changing the response contract ad hoc.
