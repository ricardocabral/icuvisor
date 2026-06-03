# Plan Review: Step 2 — Implement generator and tests

**Verdict:** Approved with implementation notes

The Step 2 plan is consistent with the accepted Step 1 design: extend `cmd/gendocs`, keep `tools.json` summary-only, add generated `web/data/tool_schemas.json`, project schema fields rather than dumping raw schemas, cap examples, and preserve deterministic/no-network generation from the registered full catalog.

## Notes to carry into implementation

- Add `go test ./cmd/gendocs` to the Step 2 verification. The planned golden/determinism coverage belongs there, but the prompt's targeted command only covers internal packages.
- Keep the public `tools.Catalog()` summary contract stable. If generator code needs schemas, prefer a dedicated docs-catalog/projection helper rather than stuffing raw `InputSchema` into `ToolDescriptor` or duplicating registry construction.
- Choose one nested-field key in generated JSON (`children` or `properties`) and test it, especially for `workout_doc` and custom item `content`.
- Ensure `make docs-tools` remains a single command that writes both files using stable temp-file replacement and fails if either output cannot be written.
- Golden assertions should cover at least: `include_full`, date fields, event `category`, `workout_doc`, capped write-tool `input_examples`, sorted tool/argument output, and absence of raw secrets/local paths/real-looking athlete IDs.

No plan blockers found for Step 2.
