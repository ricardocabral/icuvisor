# Review R008 — Code Review: Step 1

**Task:** TP-229 — Treat threshold pace as m/s and pace zones as percentages
**Step:** Step 1: Define canonical pace conversions
**Verdict:** REVISE

## Blocking finding

### 1. The preview's new formatting guard still accepts a duration outside `int`'s representable range

`paceSecondsAreFormattable` compares the rounded duration with `float64(^uint(0)>>1)` (`internal/tools/workout_target_previews.go:223-228`). On 64-bit platforms, converting `MaxInt` to `float64` rounds it up to `2^63`, so a duration of exactly `2^63` seconds passes the guard even though `formatPaceSeconds` subsequently converts it to `int` (`:267`). That conversion is outside the signed `int` range and has implementation-specific behavior.

For example, a finite `MINS_KM` threshold of `1000 / math.Exp2(63)` m/s converts to exactly `2^63` seconds/km. It is accepted by both the canonical converter and this guard, despite the intent of R007 to omit unformattable values. Reject rounded values at or above the architecture's signed-int boundary (rather than comparing against a float conversion of `MaxInt`) before calling `formatPaceSeconds`, and add an exact-boundary regression test alongside the existing `1e-306` overflow case.

## Verification

- `make test` — pass
- `make test-race` — pass
- `make lint` — pass (0 issues)
- `make build` — pass
- `git diff --check a077f6616fa8cad40619afa11a616c7b010d56f8..HEAD` — pass
- `gofmt -l` on changed Go files — clean
