# Plan Review — TP-087 Step 2: Implement validation helpers

## Verdict: APPROVE

The Step 2 plan is ready to implement. It selects the right reusable boundary (`internal/analysis`), carries forward the approved Step 1 inventory/alias/hint design, and keeps analyzer-tool routing out of scope while preserving enough metadata for future analyzers to make source-family decisions.

## What is sound

- The planned `Metric` enum plus metadata/source types in `internal/analysis` matches the task requirement for a shared analyzer contract and avoids coupling the core enum to the MCP SDK or `internal/tools`.
- `ParseMetric` with private alias handling, explicit expression detection, and deterministic hint categories directly addresses the PRD requirement: closed enum, unknown-metric hints, and no free-form arithmetic.
- Canonical schema helpers that enumerate supported values are the right approach for future analyzer input schemas; aliases should remain parse-only convenience and not expand the schema enum.
- The planned metadata surface (`MetricSources`, source family/grain/kind/unit/scale/method) is sufficient to support duplicate-source metrics later without each analyzer duplicating the table.
- Deferring tests to Step 3 is acceptable because Step 2 now has concrete helper surfaces and behavior to test.

## Implementation guardrails for Step 2

1. **Keep outputs deterministic.** `MetricValues()` and any schema helper should return canonical values in a stable order, not map iteration order, so future tool schemas and generated docs remain stable.

2. **Avoid import cycles in the error surface.** `internal/analysis` should not import `internal/tools`. Prefer an analysis-owned sentinel/type/predicate (for example usable with `errors.As`/`errors.Is`) that future tool handlers can wrap with `tools.NewUserError` at the boundary.

3. **Return defensive copies.** If `MetricValues()` or `MetricSources(metric)` return slices backed by package-level tables, copy them before returning so callers/tests cannot mutate the shared enum contract.

4. **Schema enum must be canonical-only.** Mention aliases in description prose at most; do not put aliases into the JSON Schema `enum`, or clients may learn non-canonical names as public API.

5. **Expression rejection should run before fuzzy hinting.** Inputs such as `ctl/atl`, `ctl - atl`, `weekly_tss/weekly_hours`, `(ctl+atl)/2`, comma-joined metrics, or pipe-joined metrics should consistently return the expression-not-supported message rather than being classified as unknown metrics or partial aliases.

6. **Be deliberate about source-field aliases.** Either implement the representative camelCase/read-surface aliases called out in the approved Step 1 review, or record a narrower alias set in `STATUS.md` before Step 3 tests lock behavior.

## Recommendation

Proceed with Step 2. The notes above are implementation guardrails, not blockers to the current plan.
