# Plan Review R009 — Step 3: Wire docs and catalog

**Verdict:** Approved

The Step 3 plan now addresses the blockers from R008 and is concrete enough to execute safely.

## What is now covered

- The public documentation target is explicit: `web/content/reference/resources-prompts.md`, with the expected URI/name/MIME metadata and analyzer `_meta.formula_ref` description.
- README/public-doc impact is accounted for, either by updating affected listings or logging a checked/no-change decision in `STATUS.md`.
- Future analyzer referenceability has a concrete code deliverable: exported stable-ref constants/helpers for all six formula refs, plus tests tying that surface back to the rendered markdown.
- The “catalog” scope is constrained correctly: confirm no separate generated resource catalog/hash exists and avoid unrelated tool-catalog changes.
- Targeted verification is pinned to `go test ./internal/resources ./internal/mcp`.

## Non-blocking execution notes

- Prefer exporting full formula-ref values such as `icuvisor://analysis-formulas#hr_drift`, not just fragment constants, so analyzer `_meta.formula_ref` code can avoid string concatenation and raw URI duplication.
- When the targeted tests are run, record the command/result in `STATUS.md` or the execution log so Step 4/5 can distinguish already-run targeted checks from full-suite verification.
