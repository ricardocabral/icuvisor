# Review R007 — Code Review: Step 1

**Task:** TP-229 — Treat threshold pace as m/s and pace zones as percentages
**Step:** Step 1: Define canonical pace conversions
**Verdict:** REVISE

## Blocking finding

### 1. Workout previews can bypass the new finite conversion guard and render an invalid pace

`paceTargetPreview` converts m/s in two stages (`internal/tools/workout_target_previews.go:203-217`): `paceSecondsPerMeter` accepts `1e-306` m/s because its reciprocal is still finite, then multiplying by the selected km/mile distance overflows to `+Inf`. The new canonical `response.PaceSecondsFromMetersPerSecond` correctly rejects the same `MINS_KM` conversion because `1000 / 1e-306` is non-finite (`internal/response/units.go:43-55`).

The preview path does not check the products at lines 214 and 216 before passing them to `formatPaceSeconds`; the latter rounds and casts the infinity to `int`, yielding a huge bogus duration rather than omitting the preview. Use the canonical distance conversion (or apply equivalent finite checks to the base and each percentage-derived result) and reject values that cannot be formatted safely. Add a regression test using a finite underflow-sized m/s threshold, matching the converter overflow coverage.

## Verification

- `make test` — pass
- `go test -race -count=1 ./internal/units ./internal/response ./internal/athleteprofile ./internal/analysis ./internal/tools` — pass
- `gofmt -l` on changed Go files — clean
- `git diff --check a077f6616fa8cad40619afa11a616c7b010d56f8..HEAD` — pass
