# Plan Review — TP-083 Step 3: Fixture coverage

## Verdict: Revise

The Step 3 plan is directionally aligned with the task, but it is too generic to ensure the remaining coverage gap is closed. Step 2 already has Polar/Oura assertions and an unknown fallback assertion in existing fixture tests; the plan should explicitly target the unprotected source-specific labels rather than risk satisfying “two divergent sources” by duplicating already-covered cases.

## Required changes to the plan

1. **Name the exact provider fixtures/assertions to add.**
   - Add or update coverage for **Garmin** to assert the exact readiness label: `0-100 Garmin Body Battery`.
   - Add a **WHOOP** fixture and assert exact labels, preferably both:
     - `sleepScore` provenance: `source: whoop`, `native_scale: 0-100 WHOOP sleep performance percentage`
     - `readiness` provenance: `source: whoop`, `native_scale: 0-100 WHOOP recovery score`
   This directly addresses the Step 2 review note that Garmin/WHOOP labels are still not protected.

2. **Be explicit about fixture location/load path.**
   Existing `TestGetWellnessDataFixtures` loads from `internal/intervals/testdata/wellness`, while `internal/tools/testdata/wellness` is currently empty. The plan should say whether new fixtures go into the intervals fixture directory used by `loadWellnessFixture`, or whether the helper will be changed. Avoid adding unused fixtures under `internal/tools/testdata/wellness`.

3. **Keep the unknown fallback test intentional.**
   The existing `custom_fields.json` case already asserts `source: unknown` and `native_scale: unknown` for `sleepScore`. The plan should specify whether Step 3 will strengthen that existing assertion or add a separate fallback fixture/case. Do not count manual-only rows as the unknown fallback, because they intentionally omit provenance.

4. **Run a precise targeted test command.**
   Include a concrete command such as:
   - `go test ./internal/tools -run 'TestGetWellnessData(Fixtures|NullStrippingAndIncludeFull)'`
   If a new fixture exercises `intervals.Wellness` native extraction behavior, also run:
   - `go test ./internal/intervals -run Wellness`

Once the plan names Garmin + WHOOP exact assertions and clarifies fixture placement, it should be safe to proceed.
