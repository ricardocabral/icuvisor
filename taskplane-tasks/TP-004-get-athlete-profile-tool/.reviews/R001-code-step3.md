# R001 Code Review — Step 3: Shape the response for v0.1

## Verdict

**Request changes.** The timezone fallback wiring and terse/full response boundaries are generally in place, and the package still builds, but unit normalization now conflates the independent weight preference with the profile measurement preference.

## Findings

### Blocking: `weight_pref_lb` overrides an explicit metric measurement preference

- **File:** `internal/tools/get_athlete_profile.go:180` and `internal/tools/get_athlete_profile.go:271`

`profileUnits` now calls `normalizedMeasurementPreference(profile.MeasurementPreference, profile.WeightPrefLB)`, and `normalizedMeasurementPreference` returns `"imperial"` whenever `WeightPrefLB` is true. That means a profile with an explicit metric measurement preference and pounds as the weight unit is reported as:

```json
{
  "units": {
    "measurement_preference": "imperial",
    "weight": "lb"
  }
}
```

The Step 1/Step 3 contract keeps these as separate public fields: `measurement_preference` should be the normalized upstream measurement preference, while `weight` should reflect `weight_pref_lb`. `weight_pref_lb` is a valid fallback signal only when the upstream measurement preference is absent; it should not override an explicit `METRIC`/`metric` value.

Suggested fix: normalize the explicit measurement preference first, and only infer from `weightPrefLB` when the trimmed preference is empty. For example, map values containing `IMPERIAL` to `imperial`, values containing `METRIC` to `metric`, and use `weightPrefLB` only in the empty/unknown fallback branch. Add a Step 4 test for `MeasurementPreference: "METRIC", WeightPrefLB: true` to lock this down.

## Notes

- `profileTimezone` correctly prefers the athlete profile timezone and falls back to the non-secret configured timezone passed through the registry/tool constructor.
- The `include_full` delta remains limited to the agreed fields; no raw upstream payloads, request/debug fields, timestamps, or credentials were added.

## Checks run

- `git diff d0e5154..HEAD --name-only`
- `git diff d0e5154..HEAD`
- `go test ./...` — passed
- `gofmt -l internal/tools/get_athlete_profile.go internal/tools/registry.go` — no output
- `go vet ./...` — passed
