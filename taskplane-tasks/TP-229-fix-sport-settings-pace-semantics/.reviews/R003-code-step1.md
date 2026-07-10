# Review R003 — Code Review: Step 1

**Task:** TP-229 — Treat threshold pace as m/s and pace zones as percentages
**Step:** Step 1: Define canonical pace conversions
**Verdict:** REVISE

## Findings

### 1. Valid pace thresholds disappear from `get_performance_potential` for supported display units

`profile.go` now recognizes `SECS_400M` and `SECS_250M` and places their correctly converted values in `ThresholdPaceSecondsPer400M` and `ThresholdPaceSecondsPer250M` (`internal/athleteprofile/profile.go:338-356`). `SECS_100Y` is also shaped into its sport-specific field. However, `assignPerformancePotentialPaceThresholds` only copies km, mile, 100 m, and 500 m fields (`internal/tools/get_performance_potential.go:289-298`), and `hasPerformancePotentialPaceThreshold` checks that same incomplete set (`:506-507`).

Consequently, a supported running, swimming, or rowing sport with `pace_units: "SECS_400M"`, `"SECS_250M"`, or `"SECS_100Y"` has a valid m/s threshold in the profile response but `get_performance_potential` reports `threshold_pace` unavailable. Add corresponding typed threshold fields, propagation, availability checks, and regression coverage (or retain an unambiguous m/s fallback in that tool) before recognizing those values in the shared profile model.

### 2. The documented `pace_units: "NONE"` enum value is still falsely reported as an unknown unit

The checked-in upstream schema includes `NONE` in the `pace_units` enum (`scripts/openapidiff/baseline/intervals-openapi.json:9439-9450`), but it is absent from the new pace-unit constants and `knownUnits` map (`internal/units/unit.go:22-29,66-77`). A valid setting using `NONE` therefore goes through the unknown-unit path in `applyProfilePace` (`internal/athleteprofile/profile.go:317-327`), emits `_meta.unknown_unit: "NONE"`, and logs an unknown-unit warning.

Handle `NONE` explicitly as the known no-display-preference case: retain the m/s fallback and `pace_units_source`, but do not claim that the documented enum value is unknown. Add a response-shaping test for it.

## Verification

- `make test` — pass
- `go test -race -count=1 ./internal/units ./internal/response ./internal/athleteprofile ./internal/tools` — pass
- `gofmt -l` on changed Go files — clean
- `git diff --check a077f6616fa8cad40619afa11a616c7b010d56f8..HEAD` — fails only because R002 itself has trailing whitespace on lines 3–4; the implementation diff passes when that pre-existing review prose is excluded.
