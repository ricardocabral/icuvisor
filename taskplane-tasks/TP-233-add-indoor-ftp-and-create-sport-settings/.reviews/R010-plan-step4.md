# Plan Review — TP-233 Step 4

## Verdict: REVISE

The plan omits the generator golden fixtures and the test that proves they are fresh. The registered catalog now produces 70 tools while the committed website data and both `cmd/gendocs/testdata/*.golden.json` files still have 69. As a result, `go test ./cmd/gendocs ./internal/toolcatalog ./internal/toolchecks -count=1` currently fails at `TestRunWritesGeneratedDocsGolden`; the stated verification never runs that package. Regenerating only `web/data/*.json` would leave the full suite failing.

Revise the step to:

- Add `cmd/gendocs/testdata/tools.golden.json` and `cmd/gendocs/testdata/tool_schemas.golden.json` to the artifacts, updating them from the same deterministic registry output as `web/data/tools.json` and `web/data/tool_schemas.json`. Verify the generated catalog/schema counts are 70 and inspect the public entries: `create_sport_settings` must be `settings`/`full`/`write`, and `update_sport_settings` must expose `indoor_ftp`.
- Make the PRD catalog contract explicit rather than merely naming the tools: `update_sport_settings` accepts indoor FTP in watts as a threshold-only write; `create_sport_settings` is full-toolset-only, is for a missing sport only, requires sport plus at least one threshold, and cannot carry zones, HR-zone recalculation, or historical application. Retain the existing duration-input-to-m/s pace and zone-replacement distinctions, and do not invent an indoor-versus-outdoor FTP ordering rule.
- Add an Unreleased **Added** entry covering both the indoor-FTP update and threshold-only missing-sport creation, including the no-zone-replacement boundary.
- Run `make docs-tools`, `go test ./cmd/gendocs ./internal/toolcatalog ./internal/toolchecks -count=1`, and review the generated diff so the catalog-freshness CI guard will be clean after commit.
