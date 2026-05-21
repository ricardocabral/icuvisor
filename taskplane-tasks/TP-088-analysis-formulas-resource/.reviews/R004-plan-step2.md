# Plan Review R004 — Step 2: Implement resource

**Verdict:** Changes requested

The Step 2 checklist names the right broad outcomes, but it is too thin for a resource that the PRD treats as a locked formula contract. Before implementation, expand the Step 2 plan so the worker knows exactly how the Step 1 draft will become a stable MCP resource and how definition drift will be caught.

## Blocking concerns

1. **Output shape is still ambiguous.**
   The plan says “compact markdown or JSON.” Existing static resources are `text/markdown`, and the PRD explicitly says this should follow the same pattern as `icuvisor://workout-syntax`. Pick markdown now unless there is a documented reason not to. The plan should name the constants/metadata to add, e.g. `AnalysisFormulasURI = "icuvisor://analysis-formulas"`, `AnalysisFormulasMIMEType = "text/markdown"`, resource name `analysis_formulas`, and a concise title/description.

2. **Stable formula refs/anchors are not an implementation acceptance criterion.**
   Step 1 selected full refs such as `icuvisor://analysis-formulas#hr_drift`. Step 2 must render those refs in the resource in a copyable, stable way and test that each required fragment appears exactly once. Otherwise later analyzer `_meta.formula_ref` values can drift from the resource silently.

3. **No golden-file drift guard is planned.**
   The PRD says definitions are locked and pinned with golden-file tests. Existing resource patterns already use `internal/resources/testdata/*.md` golden tests. Add a Step 2 test deliverable for `AnalysisFormulasMarkdown()` against `testdata/analysis_formulas.md`, plus focused tests that verify the six required refs, formulas/boundary words, and citations are present.

4. **Registry vs MCP protocol coverage needs to be explicit.**
   “tests for resource listing and read content” could mean only a local `NewRegistry().Register` capture test. The completion criteria require the resource to be listed and readable through MCP. The plan should include both:
   - `internal/resources` registration/handler tests following the existing static resource tests, including canceled-context behavior; and
   - updating the MCP default-resource protocol test (`TestProtocolDefaultResourceRegistryIncludesAllResources`) so `resources/list` and `resources/read` cover `AnalysisFormulasURI` through the actual server wiring.

5. **Translation from Step 1 draft to code is not constrained.**
   Step 2 should state that the implementation must use the Step 1 table in `STATUS.md` as the source for the six entries and must not add, remove, rename, or materially change formulas without updating `STATUS.md` and tests. This prevents subtle changes during code formatting from becoming unreviewed product decisions.

## Suggested Step 2 acceptance shape

Add or expand the Step 2 checklist along these lines:

- [ ] Add `internal/resources/analysis_formulas.go` with `AnalysisFormulasURI`, `AnalysisFormulasMIMEType`, `AnalysisFormulasResource()`, and `AnalysisFormulasMarkdown()`.
- [ ] Render compact markdown with six one-paragraph entries sourced from the Step 1 draft, exposing exact refs/fragments: `hr_drift`, `pw_hr_decoupling`, `polarization_index`, `efficiency_factor`, `variability_index`, and `z_score`.
- [ ] Register `AnalysisFormulasResource()` in `NewRegistryWithOptions` alongside the other static resources.
- [ ] Add golden and invariant tests under `internal/resources`, including metadata, handler read, canceled context, exact-once required refs, and citation presence.
- [ ] Update MCP protocol default-resource coverage so `resources/list` and `resources/read` include `icuvisor://analysis-formulas` with `text/markdown`.

## Non-blocking notes

- Prefer data-driven rendering from a small descriptor slice if it makes exact-once/invariant tests simpler, but avoid over-engineering a generic resource framework.
- Keep the resource terse: one paragraph per formula plus citations/boundary notes is enough. Long derivations belong neither in tool responses nor in this resource.
- Documentation and changelog updates are already assigned to later steps; do not let Step 2 grow into the Step 3 docs pass.
