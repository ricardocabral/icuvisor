# Review R004 — Code Review: Step 1

**Task:** TP-229 — Treat threshold pace as m/s and pace zones as percentages
**Step:** Step 1: Define canonical pace conversions
**Verdict:** REVISE

## Blocking finding

### 1. The canonical converters accept finite input but can return `+Inf` with `ok=true`

`PaceSecondsFromMetersPerSecond` and `PaceMetersPerSecondFromSeconds` reject non-finite *inputs* but do not validate the division result (`internal/response/units.go:41-52,54-66`). For example, each conversion with `math.SmallestNonzeroFloat64` and `MINS_KM` returns `+Inf, true`.

The read path trusts that success result and assigns it to a response field (`internal/athleteprofile/profile.go:322-323`), after which JSON encoding rejects the profile as an unsupported `+Inf` value. The write-side helper is intended to be used by the next step, where it would instead produce an invalid upstream threshold. Store the quotient first and return `false` unless it is finite (and positive); add read and write regression cases for an underflow-sized positive input.

## Verification

- `go test -count=1 ./internal/units ./internal/response ./internal/athleteprofile ./internal/tools` — pass
- `go test -race -count=1 ./internal/units ./internal/response ./internal/athleteprofile ./internal/tools` — pass
- `make test` — pass
- `git diff --check a077f6616fa8cad40619afa11a616c7b010d56f8..HEAD` — pass
- A temporary boundary probe with `math.SmallestNonzeroFloat64` reproduced the overflow described above.
