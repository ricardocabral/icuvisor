# Plan Review — TP-087 Step 1: Design the enum and aliases

## Verdict: APPROVE

The Step 1 design record is now sufficiently concrete to proceed to Step 2. It addresses the earlier review rounds: boolean wellness flags are staged out, duplicated source families are represented by multiple source descriptors, extended activity versus extended interval fields are split, structured zone/effort surfaces are excluded from the scalar enum, and the API boundary is now clearly reusable from future analyzer tools.

## What is ready for implementation

- The first-pass canonical metric inventory is explicit and tied to current read surfaces, with PRD-only examples (`pace_at_lt2`, `power_at_lt2`, `np`) documented as staged follow-ups rather than silently accepted.
- Canonical naming rules are coherent: lower `snake_case`, unit suffixes when needed, standard abbreviations only where product-facing/physiology-standard, and unit-contextual interval `pace` deferred.
- Alias policy is conservative and rejects ambiguous guesses and near-misses rather than expanding the enum through normalization tricks.
- Free-form arithmetic rejection is explicit enough to drive Step 3 table tests.
- Unknown-metric hints are deterministic, one-line, and route users toward the appropriate planned analyzer family instead of accepting math expressions.
- `internal/analysis` as an SDK-free package with `Metric`, `ParseMetric`, `MetricValues`, `MetricSources`, schema helpers, and short invalid-argument errors is the right boundary for a shared analyzer contract.
- The metadata target preserves source family, row grain, metric kind, unit labels, subjective scale references, and derived-method text, which should prevent later analyzers from duplicating or guessing source routing tables.

## Non-blocking implementation notes for Step 2

1. **Make source-field aliases exhaustive in tests.** The plan states that current camelCase/read-surface names become canonical snake_case analyzer names. When implementing aliases, include table coverage for representative source-field aliases beyond the examples already listed, especially wellness fields such as `ctlLoad`, `atlLoad`, `kcalConsumed`, `avgSleepingHR`, `hydrationVolume`, `baevskySI`, `bloodGlucose`, `bodyFat`, `fatTotal`, and `spO2` if those are intended to parse. If the implementation intentionally supports only the listed aliases, document that narrower choice in `STATUS.md` before tests are locked.

2. **Keep schema enums canonical-only.** Aliases should parse for user convenience, but schema enum values should expose canonical metric names only, with aliases mentioned in prose at most.

3. **Do not overfit analyzer routing in this task.** Step 2 should preserve metadata via `MetricSources(metric) []MetricSource`, but actual analyzer-specific source selection can remain future-tool scope as the plan says.

## Recommendation

Proceed to Step 2. The remaining concerns are implementation guardrails and test coverage details, not blockers for the design step.
