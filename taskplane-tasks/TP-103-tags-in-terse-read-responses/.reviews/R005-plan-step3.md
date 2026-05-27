# Plan Review: Step 3 — Regression tests and docs

**Verdict: Needs revision before completion**

The Step 3 checklist covers the main task requirements at a high level, but it is not yet specific enough for the current state of the branch. I ran the targeted package tests and found a real regression-test/doc-surface failure that the plan does not explicitly account for.

## Blocking plan gap

- `go test ./internal/tools ./internal/intervals` currently fails in `TestCatalogSummariesUseFirstDescriptionSentence` because the `get_activities` tool description now includes activity tags, but the catalog-summary golden expectation still has the old first sentence.
- Step 3 should explicitly include updating any tool-description/catalog golden tests and generated tool-reference/catalog artifacts if response-shape descriptions changed, not only `CHANGELOG.md`.

## Required additions to the Step 3 plan

1. Add a concrete audit of all regression coverage already added in Steps 1–2:
   - event and activity tags present/order;
   - explicit empty arrays emitted;
   - missing/null omitted from terse rows;
   - non-array/mixed malformed payloads omitted without panic;
   - `include_full` preserves raw upstream payloads, including raw `null` where present.
2. Include catalog/schema/doc test updates caused by description changes:
   - update `internal/tools/catalog_test.go` golden expectations or adjust descriptions intentionally;
   - run/refresh generated tool docs only if this repository treats changed tool descriptions/schemas as generated-doc inputs.
3. Keep `CHANGELOG.md` update under `[Unreleased]`, ideally in `### Added` because this is a user-visible response addition.
4. Name the targeted verification command(s), at minimum:
   - `go test ./internal/tools ./internal/intervals`
   - plus any narrower tests used while fixing the catalog/doc expectations.

Once the plan includes these catalog/doc-surface updates and the targeted tests pass, Step 3 should be safe to complete.
