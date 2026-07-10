# Review R005 — Code Review: Step 1

**Task:** TP-229 — Treat threshold pace as m/s and pace zones as percentages
**Step:** Step 1: Define canonical pace conversions
**Verdict:** REVISE

## Blocking findings

### 1. Workout target previews still treat stored m/s as a duration

`internal/tools/workout_target_previews.go:135-140` passes raw `setting.ThresholdPace` to `paceTargetPreview`, and `paceSecondsPerMeter` (`:160-178`) divides it as though it were seconds for the `pace_units` distance. With a real `MINS_KM` threshold of `3.5714285` m/s, it calculates `0.003571` seconds/m and emits a roughly `0:04/km` threshold instead of `4:40/km`; percentage-target previews are consequently impossible paces. The existing tests conceal this by retaining synthetic duration-valued thresholds (300 and 90).

Make the preview path use the canonical m/s-to-seconds conversion (with the display unit only selecting distance), update its fixtures, and cover a real m/s run and swim example.

### 2. Pace histogram configured zones still label percentages as seconds-per-distance boundaries

`histogramZoneConfig` sends raw `setting.PaceZones` to `analysis.ConvertPaceZoneBoundary` (`internal/tools/get_activity_histogram.go:271-281`). That converter (`internal/analysis/histogram.go:124-154`) assumes its input is a pace duration. Real boundaries such as `[77.5, 100]` with `MINS_KM` are therefore installed as `[77.5, 100] seconds_per_km`, rather than percentage-of-threshold boundaries derived from the stored m/s threshold. The histogram will report configured pace-zone buckets with false units and classify normal samples incorrectly.

Convert percentage boundaries using the sport threshold before building the histogram (or deliberately fall back to fixed-width buckets where that relationship cannot be represented), and add regression coverage using m/s plus `[77.5, 100]`.

### 3. The unknown/overflow fallback retains a field whose advertised unit is now false

When no display-duration conversion is possible, `applyProfilePace` writes both the unambiguous `threshold_pace_meters_per_second` and the legacy `threshold_pace_value` (`internal/athleteprofile/profile.go:324-327`). `get_performance_potential` then exports the latter as unit `source_unit` (`internal/tools/get_performance_potential.go:307-310`). For `pace_units_source: "FEET"`, for example, both values are the stored m/s value, so the `threshold_pace_value` entry is explicitly mislabeled as FEET/source-unit. The same bad legacy field is produced for a known display unit whose conversion overflows.

Omit the ambiguous legacy field in these fallback cases (or define and document it as m/s, though that duplicates the explicit field); do not propagate it as `source_unit`.

## Verification

- `go test -race -count=1 ./internal/units ./internal/response ./internal/athleteprofile` — pass
- `make test` — pass
- `git diff --check a077f6616fa8cad40619afa11a616c7b010d56f8..HEAD` — pass
