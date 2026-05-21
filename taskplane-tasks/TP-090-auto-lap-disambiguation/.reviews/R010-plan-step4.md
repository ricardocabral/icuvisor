# Plan Review: Step 4 — Tests and docs

## Verdict

Approve, with implementation cautions.

## Review scope

- Read `PROMPT.md` and current `STATUS.md` for TP-090.
- Checked the current interval-source implementation/test locations in `internal/analysis` and `internal/tools/get_activity_details_test.go`.
- Checked current docs surfaces: `web/content/reference/tools.md`, generated `web/data/tools.json`, and `CHANGELOG.md`.

## Assessment

The Step 4 checklist is short, but it is backed by the detailed Step 1 acceptance examples already recorded in `STATUS.md`. That gives enough direction for this step: add fixture-backed coverage for structured intervals, 1 km auto-laps, 1 mi auto-laps, and unknown source; update public docs/changelog; and run the quality gate.

## Implementation cautions

1. **Use real test fixtures, not only inline JSON.**
   The prompt explicitly asks for fixtures. Existing inline coverage is useful, but Step 4 should add committed fixture files under the allowed `internal/tools/testdata/**/*` scope, then load them from tests. Make sure the fixture matrix includes the currently missing 1 mi auto-lap case.

2. **Exercise the public tool response, not only the classifier.**
   Keep or add tests that call `get_activity_intervals` and assert `_meta.interval_source` and `_meta.auto_lap_suspected` in the shaped response. Include at least:
   - structured workout -> `structured_workout`, `false`;
   - 1 km generic laps -> `device_laps`, `true`;
   - 1 mi generic laps -> `device_laps`, `true`;
   - insufficient/mixed/unknown rows -> `unknown`, `false`.

3. **Preserve terse/full behavior while adding metadata.**
   The metadata should be present on successful interval responses in default terse mode and should not require `include_full`. `include_full` can expose raw rows, but it should not be needed to see the new `_meta` signals.

4. **Update generated/reference docs through the right surface.**
   If the public tool description is changed to mention `_meta.interval_source` / `_meta.auto_lap_suspected`, update the registry description and run `make docs-tools` so `web/data/tools.json` stays generated. Do not hand-edit generated catalog output without the corresponding source change. If the generated catalog is too terse to document return metadata, add a small prose note in `web/content/reference/tools.md` near the existing reference text.

5. **Record user-visible behavior in `CHANGELOG.md`.**
   Add an `[Unreleased]` entry describing the new additive `get_activity_intervals` metadata and analyzer caution/propagation support. This is user-visible even though it is backward-compatible.

6. **Quality gate overlap with Step 5 is acceptable but should be explicit.**
   If Step 4 runs `make test`, `make build`, and `make lint`, record outcomes in `STATUS.md`; Step 5 can then repeat/verify them. If any failure is pre-existing or unrelated, document it rather than leaving the “Run full quality gate” box ambiguous.

## Summary

Proceed with Step 4. The main risk is treating existing inline/unit coverage as sufficient despite the task's fixture requirement and the missing 1 mi public-tool response case. Address the cautions above during implementation.
