# Code Review — TP-083 Step 1: Define source scale mapping

## Verdict: Request changes

The direction is correct for unknown-source sleep scores, but this step currently leaves the repository with failing tests and the readiness mapping is still too coarse for the native sidecars already called out by the task/status notes.

## Findings

1. **Tests fail after changing the unknown-source fallback.**
   - `internal/tools/get_wellness_data.go:325-329` now returns `"unknown"` for unknown `sleepScore` sources, but `internal/tools/get_wellness_data_test.go:118-120` still asserts `"0-100 device nightly score"`.
   - Repro: `go test ./internal/tools` fails in `TestGetWellnessDataFixtures/custom_null_fields_and_unknown_source_provenance`.
   - Please update the test expectation in this step or keep the behavior change out until the fixture/test step. The branch should not be left red after a code change.

2. **Polar readiness still reports the wrong native scale when the evidence is `ans_charge`.**
   - `wellnessFieldSource` treats either `polar.nightly_recharge_status` or `polar.ans_charge` as just source `"polar"` (`internal/tools/get_wellness_data.go:272-274`), and `wellnessNativeScale` then always returns `"1-6 Polar nightly_recharge_status"` for Polar readiness (`internal/tools/get_wellness_data.go:28-32`, `334-338`).
   - The task notes and prior review explicitly distinguish Polar `ans_charge` from `nightly_recharge_status`. A row bridged from only `ans_charge` would now claim the 1-6 nightly-recharge scale even though that sidecar is a different metric/scale.
   - Please make the mapping key include the native sidecar (or otherwise pass enough evidence into `wellnessNativeScale`) so `native_scale` reflects the actual readiness source, with defined precedence when both sidecars exist.

3. **Step 1 mapping is not documented as the requested source/field matrix.**
   - `STATUS.md:114-115` lists labels in prose, but it does not define the exact table by field/source/native-sidecar or cite whether each entry comes from public docs vs observed fixtures.
   - Since later steps and tests depend on exact strings, please record the mapping in `STATUS.md` (or an implementation note) as a matrix covering `sleepScore` and `readiness` for Garmin, WHOOP, Oura, Polar, and unknown, including sidecar-specific Polar readiness behavior.

## Verification

- Ran `go test ./internal/tools` — fails due the stale unknown-source test expectation described above.
