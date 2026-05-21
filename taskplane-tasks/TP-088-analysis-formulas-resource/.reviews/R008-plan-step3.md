# Plan Review R008 — Step 3: Wire docs and catalog

**Verdict:** Changes requested

The Step 3 checklist points at the right outcomes, but it is still too broad to execute safely. This step is where the new resource becomes discoverable to users and to the upcoming analyzer tasks, so the plan needs to name the exact doc/catalog surfaces and define what “stable formula refs” means in code.

## Blocking concerns

1. **Documentation target is underspecified.**
   “Update resource reference docs” should explicitly name `web/content/reference/resources-prompts.md` and require adding `icuvisor://analysis-formulas` to the Resources table with `analysis_formulas`, `text/markdown`, and a terse description that mentions analyzer `_meta.formula_ref` values. The plan should also decide whether README is affected; if not updated, record that it was checked and why no change is needed.

2. **Analyzer referenceability needs a concrete contract.**
   Future analyzer tasks should not have to duplicate raw strings like `icuvisor://analysis-formulas#hr_drift`. The current Step 2 code exports only `AnalysisFormulasURI`; the fragments live in unexported data. Expand the plan to either add exported formula-ref constants/helpers under `internal/resources` (for all six refs) with tests that they match the rendered markdown, or explicitly justify why documentation-only refs are sufficient for TP-089+ consumers. Without this, `_meta.formula_ref` drift remains easy in later analyzer work.

3. **“Catalog” is not defined.**
   There does not appear to be a separate generated resource catalog analogous to the tool catalog. The plan should state that the public resource catalog for this step is `web/content/reference/resources-prompts.md` (or identify any other catalog/hash surface that must be updated). Avoid touching the generated tool catalog unless a real tool-catalog change is intended.

4. **Targeted test command should be pinned.**
   The plan should name the targeted verification to run after docs/referenceability changes, at minimum `go test ./internal/resources ./internal/mcp`. If exported constants or helper tests are added, include the specific new/updated tests in `internal/resources` that prove all six refs remain exact.

## Suggested Step 3 acceptance shape

- [ ] Update `web/content/reference/resources-prompts.md` to list `icuvisor://analysis-formulas` with name `analysis_formulas`, MIME `text/markdown`, and a description of canonical analyzer formula refs.
- [ ] Review README/public docs for affected resource listings; update or log “checked, no change” in `STATUS.md`.
- [ ] Add or confirm a code-level stable-ref surface for analyzers (prefer exported constants/helpers for `hr_drift`, `pw_hr_decoupling`, `polarization_index`, `efficiency_factor`, `variability_index`, and `z_score`) and test that it matches the markdown refs exactly once.
- [ ] Confirm there is no separate generated resource catalog/hash to refresh; do not modify the tool catalog.
- [ ] Run `go test ./internal/resources ./internal/mcp` and record the result in `STATUS.md`.

Once those details are added, the step should be straightforward and low risk.
