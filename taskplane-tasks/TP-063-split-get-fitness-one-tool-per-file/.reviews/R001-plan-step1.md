# Plan Review: Step 1 — Map `get_fitness.go`

Verdict: **Approved with notes**

The Step 1 plan matches the task: before moving code, catalog every top-level declaration by ownership and identify helpers that are genuinely shared across the future split files. This is the right first step for a no-behaviour-change mechanical refactor.

## Notes / required diligence for Step 1 output

When updating `STATUS.md`, make the map concrete enough that Step 2 can be executed mechanically. In particular, record both **tool ownership** and the intended **destination file** for each declaration.

Pay attention to these likely shared declarations so they are not accidentally left in the wrong tool file or duplicated with behaviour drift:

- Shared by `get_fitness` and `get_training_summary`:
  - `FitnessClient`
  - `dateRangeRequest`
  - `decodeDateRangeRequest`
  - `dateRangeInputSchema`
  - `toolProfile`
- Shared by curve tools (`get_power_curves` / `get_best_efforts`):
  - `defaultDurationBuckets`
  - `rangeCurveSpec`
  - `firstCurve`
  - `valueAtBucket`
  - `activityIDAt` if `firstCurve`/bucket helpers stay generic
- Shared broadly across multiple split files:
  - `validDate`
  - `normalizePositiveInts`
  - `encodeShaped`
  - `roundPtr`
  - `genericOutputSchema`
- Tool-specific items should be marked for their final files:
  - `get_fitness.go`: `getFitness*`, `fitness*`, `newGetFitnessTool`, `getFitnessHandler`, `shapeFitnessRows`
  - `get_best_efforts.go`: `getBestEfforts*`, `BestEffortsClient`, `bestEfforts*`, `normalizeSports`, `bestEffortsCurveSpec`, `bestEffortsForSport`, effort row helpers, distance bucket defaults
  - `get_power_curves.go`: `getPowerCurves*`, `PowerCurvesClient`, `powerCurves*`, `defaultPowerCurveSport`, `decodePowerCurvesRequest`, `bucketPowerCurve`
  - `get_training_summary.go`: `getTrainingSummary*`, `trainingSummary*`, `newGetTrainingSummaryTool`, `getTrainingSummaryHandler`, `shapeTrainingSummary`, distance/zone accumulation helpers

## Guardrails

- Do not treat this mapping as permission to refactor schemas or helper logic. Preserve names and bodies unless a move requires import adjustment.
- Include the const/var blocks in the catalog; they are easy to miss because they combine declarations for multiple tools.
- For shared helpers, prefer `fitness_shared.go` only when the helper is used by two or more destination files, as the prompt requires.
- The Step 1 deliverable should be a declaration-level map in `STATUS.md`, not just a narrative summary.

No blockers for proceeding with Step 1.
