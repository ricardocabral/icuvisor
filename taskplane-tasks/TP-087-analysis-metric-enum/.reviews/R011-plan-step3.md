# Plan Review — TP-087 Step 3: Tests

## Verdict: REVISE

The Step 3 plan covers the core parser/schema cases and correctly adds the round-trip regression test for `MetricValues()` entries, which is the main issue found during Step 2 code review. However, it is still too narrow for the helper surface that Step 2 introduced as the shared analyzer contract. Before implementation, expand the plan so the tests protect both the parser behavior and the public metadata/schema API future analyzer tools will consume.

## Required plan updates

1. **Make unknown-metric hint coverage explicit.**
   The current “unknown names” item could be satisfied by one generic unsupported value. Add table cases for each deterministic hint category recorded in Step 1:
   - best-effort/duration or distance names, e.g. `5min_power`, `5k_pace`
   - zone/load-balance names, e.g. `power_zone_distribution_seconds`, `load_balance`
   - segment/stream-stat names, e.g. `mean_power_segment`, `hr_drift_stream`
   - compliance/adherence names, e.g. `completed_vs_scheduled`
   - generic unsupported names, e.g. `ftp`, `distance`, `tss`

2. **Pin expression rejection examples, including the prior `_per_` regression.**
   Keep the round-trip test over every `MetricValues()` entry, and also add explicit expression cases from the design record such as `ctl/atl`, `ctl - atl`, `weekly_tss/weekly_hours`, `(ctl+atl)/2`, comma/pipe-joined metrics, and `tss_per_hour`. This ensures unsupported `_per_` formulas are rejected while canonical `pace_seconds_per_km` and `pace_seconds_per_mile` parse successfully.

3. **Add schema contract tests, not just parser tests.**
   Tests should assert that `MetricSchemaProperty()`:
   - has `type: "string"`
   - has an `enum` exactly matching `MetricValues()`
   - enumerates canonical values only, excluding aliases such as `resting_hr`, `sleepSecs`, `intensity_factor`, and `compliance_percent`
   - includes concise description prose that mentions aliases and expression rejection without dumping internals

4. **Cover the public reusable metadata helpers.**
   Step 2 added more than parsing. Add tests for the analyzer contract surface:
   - `ValidateMetricCatalog()` succeeds for the static catalog.
   - every canonical metric from `MetricValues()` has at least one `MetricSources()` descriptor.
   - representative multi-source metrics (`ctl`, `training_load`, `feel`) expose the expected source families/grains/kinds.
   - derived metrics (`weekly_tss`, `weekly_hours`) expose `KindDerived`, `SourceDerivedWeekly`, units, and method text.
   - scale metrics expose subjective scale metadata where applicable.
   - returned source slices are defensive copies, so mutating a result does not alter subsequent calls.

5. **Cover the error type/predicate contract.**
   Since future tools will use this to map parse failures to user-facing invalid-argument errors, include assertions that invalid parses satisfy `IsInvalidMetric(err)` and valid parses do not produce an error. Error text should remain short and should not include Go type names, package paths, source field tables, or stack-like details.

6. **Specify targeted test commands.**
   The plan should name the targeted command, at minimum `go test ./internal/analysis`. If any tests are added outside the package, include those package paths as well. Full `make test` can remain in the later verification step.

## What is already sound

- Table-driven parser tests for canonical values, aliases, unknown values, and arithmetic expressions are the right structure.
- The added round-trip coverage for every schema enum value directly addresses the Step 2 regression around canonical `*_per_*` metrics.
- Keeping this step focused on targeted tests, with the full quality gate deferred to Step 5, matches the task flow.
