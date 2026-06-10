# R011 Code Review — Step 2

Verdict: REVISE

## Findings

1. **Full test suite is broken by incomplete catalog/golden integration.**  
   `go test ./...` fails in two places after adding `add_unavailable_date_range`:
   - `cmd/gendocs`: `TestRunWritesGeneratedDocsGolden` reports `cmd/gendocs/testdata/tools.golden.json` is stale; the generated catalog includes the new tool after `add_or_update_event`, but the golden fixture does not.
   - `internal/safety`: `TestAdversarialStaticCatalogMatrix` reports safe/full registered counts are one too high because `internal/safety/adversarial_test.go`'s static catalog list does not include `add_unavailable_date_range` near lines 23-26.

   Step 2 marks schema/catalog/core/full surfaces complete, but these catalog guard surfaces were missed. Please add the new write tool to the adversarial static matrix and regenerate/update the gendocs golden fixtures.

2. **One schema example violates the tool's own enum.**  
   In `internal/tools/add_unavailable_date_range.go:312,323`, the input schema declares a case-sensitive enum containing `INJURY`, but the third example uses `"injury"`. Runtime normalization accepts lowercase, but MCP/JSON Schema clients that validate examples or arguments against the schema will treat that example as invalid. Make the example use an enum value (for example `INJURY`) or adjust the schema strategy so examples are valid under the published schema.

## Verification

- `go test ./internal/tools ./internal/mcp ./internal/toolchecks` passes.
- `go test ./...` fails as described above.
