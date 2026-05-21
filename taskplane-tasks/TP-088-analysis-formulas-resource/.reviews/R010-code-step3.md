# Code Review R010 — Step 3: Wire docs and catalog

**Verdict:** Approved

No blocking findings.

## Review notes

- `web/content/reference/resources-prompts.md` now lists `icuvisor://analysis-formulas` with the expected `analysis_formulas` name, `text/markdown` MIME type, and `_meta.formula_ref` usage note.
- README's resource package summary was updated for the new analysis formulas resource.
- Stable exported formula-ref constants were added and the markdown rendering now uses that single code-level surface.
- Existing invariant coverage checks each stable ref appears exactly once in the rendered markdown, alongside the formula, boundary, and citation text.
- STATUS records that no separate generated resource catalog/hash was found and that the tool catalog was intentionally left untouched.

## Verification

- Ran `go test ./internal/resources ./internal/mcp` — passed.
- Ran `golangci-lint run ./internal/resources` — passed.
