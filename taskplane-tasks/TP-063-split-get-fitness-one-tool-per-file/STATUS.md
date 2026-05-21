# TP-063-split-get-fitness-one-tool-per-file — Status

**Current Step:** Step 4: Wrap up
**Status:** ✅ Complete
**Last Updated:** 2026-05-17
**Review Level:** 1
**Review Counter:** 3
**Iteration:** 1
**Size:** S

---

### Step 1: Map the contents

**Status:** ✅ Complete

- [x] Catalog every top-level declaration in `internal/tools/get_fitness.go` by tool ownership and record the map in STATUS.md.
- [x] Identify helpers shared by two or more tools and record the extraction decision in STATUS.md.

#### Step 1 declaration map

- `get_fitness`: `getFitnessName`, `getFitnessDescription`, `invalidFitnessArgumentsMessage`, `fetchFitnessMessage`, `FitnessClient`, `dateRangeRequest`, `fitnessResponse`, `fitnessRow`, `fitnessMeta`, `newGetFitnessTool`, `getFitnessHandler`, `decodeDateRangeRequest`, `shapeFitnessRows`, `dateRangeInputSchema`, plus shared `toolProfile`, `encodeShaped`, `roundPtr`, `validDate`, `genericOutputSchema`.
- `get_best_efforts`: `getBestEffortsName`, `getBestEffortsDescription`, `invalidCurveArgumentsMessage`, `fetchBestEffortsMessage`, `defaultBestEffortSports`, `defaultDurationBuckets`, `defaultRunDistanceBuckets`, `defaultSwimDistanceBuckets`, `BestEffortsClient`, `bestEffortsRequest`, `bestEffortsResponse`, `bestEffortsSport`, `bestEffortRow`, `bestEffortsMeta`, `newGetBestEffortsTool`, `getBestEffortsHandler`, `decodeBestEffortsRequest`, `normalizeSports`, `normalizePositiveInts`, `bestEffortsCurveSpec`, `bestEffortsForSport`, `effortRowsFromDurationCurve`, `effortRowsFromDistanceCurve`, `activityIDAt`, `defaultDistanceBucketsForSport`, `bestEffortsInputSchema`, plus shared `rangeCurveSpec`, `firstCurve`, `valueAtBucket`, `encodeShaped`, `roundPtr`, `validDate`, `genericOutputSchema`.
- `get_power_curves`: `getPowerCurvesName`, `getPowerCurvesDescription`, `invalidCurveArgumentsMessage`, `fetchPowerCurvesMessage`, `defaultPowerCurveSport`, `defaultDurationBuckets`, `PowerCurvesClient`, `powerCurvesRequest`, `powerCurvesResponse`, `powerCurvePoint`, `powerCurvesMeta`, `newGetPowerCurvesTool`, `getPowerCurvesHandler`, `decodePowerCurvesRequest`, `normalizePositiveInts`, `rangeCurveSpec`, `firstCurve`, `bucketPowerCurve`, `valueAtBucket`, `activityIDAt`, `powerCurvesInputSchema`, plus shared `encodeShaped`, `roundPtr`, `validDate`, `genericOutputSchema`.
- `get_training_summary`: `getTrainingSummaryName`, `getTrainingSummaryDescription`, `invalidFitnessArgumentsMessage`, `fetchTrainingSummaryMessage`, `FitnessClient`, `dateRangeRequest`, `trainingSummaryResponse`, `trainingSummaryTotals`, `trainingSportTotals`, `trainingSummaryMeta`, `newGetTrainingSummaryTool`, `getTrainingSummaryHandler`, `decodeDateRangeRequest`, `shapeTrainingSummary`, `addFloatSlices`, `setDistance`, `addDistance`, `addPtr`, `dateRangeInputSchema`, plus shared `toolProfile`, `encodeShaped`, `validDate`, `genericOutputSchema`.

#### Step 1 shared-helper decision

Create `internal/tools/fitness_shared.go` for declarations used by two or more split files: `FitnessClient`, `dateRangeRequest`, `invalidFitnessArgumentsMessage`, `invalidCurveArgumentsMessage`, `defaultDurationBuckets`, `validDate`, `normalizePositiveInts`, `rangeCurveSpec`, `firstCurve`, `valueAtBucket`, `toolProfile`, `encodeShaped`, `roundPtr`, `dateRangeInputSchema`, and `genericOutputSchema`. Keep single-tool helpers with their owning tool: best-efforts-only normalization/default-distance/effort shaping, power-curve-only bucketing, fitness-only row shaping, and training-summary-only aggregation helpers.

### Step 2: Split

**Status:** ✅ Complete

- [x] Capture pre-refactor tool catalog and schema snapshots for later byte-identical diffs.
- [x] Move per-tool types, schemas, handlers, and constructors into one file per tool.
- [x] Move shared helpers used by two or more split files into `fitness_shared.go`.
- [x] Mirror the test split without changing test assertions or fixtures.

### Step 3: Verify byte-identical behaviour

**Status:** ✅ Complete

- [x] Run the existing tools package tests after the split.
- [x] Run `make build`, `make test`, `make test-race`, and `make lint`.
- [x] Diff the post-refactor tool catalog against the pre-refactor snapshot.
- [x] Diff `scripts/snapshot_tool_schemas.go` post-refactor output against the pre-refactor snapshot.

### Step 4: Wrap up

**Status:** ✅ Complete

- [x] Update `CHANGELOG.md` `[Unreleased]` Changed with the internal refactor note.
- [x] Confirm extraction and finalization commits/status reflect the split and verification work.

#### Step 4 commit/status note

Extraction was committed at `8b29cb9` (`refactor(TP-063): split get fitness tools into separate files`) and parity verification at `ce714af` (`test(TP-063): verify split fitness tool parity`). Finalization will be committed with this changelog/status update.

| 2026-05-17 02:31 | Task started | Runtime V2 lane-runner execution |
| 2026-05-17 02:31 | Step 1 started | Map the contents of `get_fitness.go` |
| 2026-05-17 02:33 | Review R001 | plan Step 1: APPROVE |
| 2026-05-17 02:37 | Review R002 | plan Step 2: APPROVE |
| 2026-05-17 02:48 | Review R003 | plan Step 3: APPROVE |

| 2026-05-17 03:24 | Worker iter 1 | done in 3168s, tools: 95 |
| 2026-05-17 03:24 | Task complete | .DONE created |