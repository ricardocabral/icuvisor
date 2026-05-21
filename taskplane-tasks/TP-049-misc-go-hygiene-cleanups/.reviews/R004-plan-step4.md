# Plan Review: Step 4 — Reformat long constructor lines

Verdict: Approved.

The Step 4 plan is appropriately narrow for item 4: only wrap the four long `newXxxTool` return expressions in `internal/tools/get_fitness.go:210-224` and verify the file remains clean under Go formatting. This is a formatting-only change and should not alter tool registration, schemas, handler arguments, or response behavior.

Implementation notes to keep the diff mechanical:

- Limit the edit to the four constructor return statements at lines 211, 215, 219, and 223. The long function signatures themselves are not part of this item unless `gofmt` changes them.
- Prefer a standard multi-line composite literal style, for example:
  - `return coreTool(Tool{` / one field per line / `})`
  - keep the existing field values and order unchanged.
- Do not rename tools, tweak descriptions, change input/output schemas, or refactor the shared constructor/helper pattern in this step.
- Since imports are unaffected, `gofmt -w internal/tools/get_fitness.go` should be sufficient; running `goimports` is fine but should produce no import churn.
- The requested `make build`, `make test`, and `make lint` checks are adequate for this formatting-only step before committing.

No blockers found.
