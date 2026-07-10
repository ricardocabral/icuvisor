# Review R002 — Code Review: Step 1

**Task:** TP-229 — Treat threshold pace as m/s and pace zones as percentages  
**Step:** Step 1: Define canonical pace conversions  
**Verdict:** REVISE

## Blocking finding

### 1. The new public metadata advertises a response contract that is neither present nor applied

`internal/athleteprofile/profile.go:124-126` now says that `threshold_pace` is converted from m/s and that `pace_zones_percent_of_threshold` is returned. But `Sport` has no such percentage field (`:36-65`), and `applyProfilePace` still routes the raw m/s threshold and raw percentage zones through `ToPreferredWithRaw` into `threshold_pace_seconds_per_*` and `pace_zones_seconds_per_*` (`:314-375`). The new conversion helpers are not called by this path.

For example, a `MINS_KM` setting with upstream `threshold_pace: 3.5714285` still emits `threshold_pace_seconds_per_km: 3.5714285`, rather than 280; zones `[77.5, 100]` still emit as `pace_zones_seconds_per_km`, while the metadata names a field that is absent. The existing integration expectation at `internal/tools/get_athlete_profile_test.go:181` continues to lock this incorrect contract in.

Implement the decided migration before advertising it: use `PaceSecondsFromMetersPerSecond` for known presentation units, add and populate the percentage field while omitting the misleading duration-zone fields, and update the old expectations. If that behavior is deliberately owned by Step 2, defer the metadata change until the corresponding response shape exists; do not publish a false `_meta` contract in the interim.

## Verification

- `gofmt -l` on all changed Go files → clean
- `go test -count=1 ./internal/units ./internal/response ./internal/athleteprofile` → pass
- `go test -race ./internal/units ./internal/response ./internal/athleteprofile` → pass
- `git diff --check a077f6616fa8cad40619afa11a616c7b010d56f8..HEAD` → pass

The conversion formulas and finite-value guards in `internal/response/units.go` are correct; the blocker is that the response which declares these semantics still does not use them.
