# Plan Review R005 — Step 2: Implement resource

**Verdict:** Approved

The Step 2 plan has been expanded enough to guide implementation of the new MCP resource and now addresses the blockers from R004. It commits to the existing static-resource pattern (`text/markdown`, `analysis_formulas`, explicit constants/functions), stable refs/fragments, resource-registry wiring, golden/invariant tests, and MCP protocol list/read coverage.

## Approval notes

- The output shape is now concrete: `AnalysisFormulasURI`, `AnalysisFormulasMIMEType`, `AnalysisFormulasResource()`, `AnalysisFormulasMarkdown()`, resource name `analysis_formulas`, and `text/markdown`.
- Stable analyzer references are included as acceptance criteria, with exact fragments for all six formulas.
- Drift protection is covered through a golden test plus invariants for refs, formulas/boundary wording, and citation presence.
- Registry and protocol coverage are both included, which matches the completion criterion that the resource be listed and readable through MCP.
- The plan correctly keeps docs/changelog work in later steps instead of broadening Step 2.

## Implementation reminders

- Register the resource in `NewRegistryWithOptions` so both `NewRegistry()` and the MCP default registry include it.
- Keep the rendered text sourced from the Step 1 table in `STATUS.md`; any material formula wording or boundary-rule change should update `STATUS.md` and the golden test together.
- Use the existing resource-handler pattern: check `ctx.Err()` before rendering and return `Result{URI, MIMEType, Text}` with the resource constants.
