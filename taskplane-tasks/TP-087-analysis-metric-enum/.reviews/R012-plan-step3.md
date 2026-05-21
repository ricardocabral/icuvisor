# Plan Review — TP-087 Step 3: Tests

## Verdict: APPROVE

The revised Step 3 plan now covers the parser, schema, metadata, and error-contract surfaces introduced in Step 2. It directly addresses the prior R011 gaps and is appropriately scoped to targeted unit tests before the later full quality gate.

## What is sound

- Parser coverage is table-driven and includes canonical metrics, safe aliases, unknown names, arithmetic expressions, and the `_per_` regression through a `MetricValues()` round-trip test.
- Unknown-metric hints are now explicit for each deterministic category from the Step 1 design record: efforts, zones/load balance, segment/stream stats, compliance/adherence, and generic unsupported names.
- Schema tests are planned against the public contract: `type: "string"`, enum matching `MetricValues()`, canonical-only enum values, concise description prose, alias mention, and expression rejection guidance.
- Metadata helper coverage protects the reusable analyzer contract: catalog validation, source descriptors for every metric, multi-source representatives, derived weekly metrics, subjective scales, and defensive-copy behavior.
- Error contract tests cover `IsInvalidMetric(err)` and concise user-facing error text, which future analyzer tools will rely on for invalid-argument handling.
- The targeted command `go test ./internal/analysis` is specific and sufficient if all Step 3 tests remain in that package.

## Minor implementation guidance

- Keep assertions on description/error concision resilient enough to avoid brittle copy tests, while still checking that no internal package/type/source-table details leak.
- In the schema enum test, compare as sets or sorted slices in the same order as `MetricValues()` so the test documents the deterministic contract without duplicating the whole enum manually.
- If any tests are added outside `internal/analysis`, update the targeted command in STATUS.md before executing Step 3.
