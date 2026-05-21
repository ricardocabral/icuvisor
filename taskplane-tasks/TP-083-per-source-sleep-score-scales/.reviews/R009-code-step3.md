# Code Review — TP-083 Step 3: Fixture coverage

## Verdict: Approve

No blocking findings.

## Review notes

- Garmin fixture coverage now asserts both `source: garmin` and the exact `native_scale: 0-100 Garmin Body Battery` for readiness.
- The new WHOOP fixture is in the directory used by `loadWellnessFixture` and covers both divergent labels: sleep performance percentage for `sleepScore` and recovery score for `readiness`.
- The existing unknown-provider fixture now explicitly verifies `source: unknown` and `native_scale: unknown`, satisfying the fallback requirement without changing manual-only provenance behavior.
- Interval native-extraction tests now include Garmin sleep score and WHOOP sleep/recovery sidecars, which protects the fixture assumptions used by the tool-level tests.

## Tests run

- `go test ./internal/tools -run 'TestGetWellnessData(Fixtures|NullStrippingAndIncludeFull)'`
- `go test ./internal/intervals -run Wellness`
