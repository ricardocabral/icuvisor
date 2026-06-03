# Plan Review — Step 2

Result: Approved with guidance.

The Step 2 plan follows the Step 1 decision: keep reliable interval-source classification on `get_activity_intervals`, strengthen routing/description surfaces, and add regression coverage without coupling `get_activity_details` to an extra intervals fetch.

Execution guidance:
- Add an explicit regression for routing wording, not only classifier behavior. Existing tests already cover `_meta.interval_source` on `get_activity_intervals`; add/adjust a tool metadata/schema test that locks the instruction that lap/rep/interval-execution claims require `get_activity_intervals`.
- If only top-level `Tool.Description` later sentences change, schema snapshots should not drift. If `activityReadInputSchema()` descriptions change, regenerate/check `internal/tools/schema_snapshot/get_activity_details.json` and `get_activity_intervals.json` with the snapshot script.
- If the first sentence of either tool description changes, also regenerate/check generated doc surfaces (`make docs-tools`, `web/data/tools.json`, and `cmd/gendocs/testdata/tools.golden.json`) or deliberately keep the first sentence stable.
- Keep default response shape compact: do not add raw interval payloads or inferred interval-source fields to `get_activity_details` unless the plan is revised. It is fine to update `activityReadOutputSchema()` to clarify that interval-source evidence appears on interval responses and details are not sufficient for lap/repetition analysis.
- Update `CHANGELOG.md` under `[Unreleased]` and record any no-snapshot/no-generated-doc decision in `STATUS.md` so Step 3 has a clear verification trail.

The targeted command `go test ./internal/tools ./internal/analysis` is appropriate for this implementation checkpoint; broader generated-doc/full-suite failures can be caught in Step 3, but avoid knowingly leaving generated artifacts stale at the Step 2 boundary.
