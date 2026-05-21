# R014 Plan Review — Step 3: Register tools and descriptions

**Verdict:** APPROVE

The revised Step 3 plan in `STATUS.md` addresses the blockers from R013. It now covers the four public adapters, strict request/schema handling, source loading responsibilities, computation/response shaping, shared tool-catalog registration, coach scoping, full-toolset placement, activation-hint descriptions, conservative formula metadata, and the intentional docs/CHANGELOG deferral.

## Notes for implementation

- Keep the new tools as `fullTool` + `RequirementRead`; do not promote any analyzer to `core` in this task.
- Add all four names to `internal/toolcatalog/catalog.go` as athlete-scoped before registering them in `registryBaseTools`, otherwise registry/coach ACL checks will fail.
- Make the first sentence of each description include both the activation hint and the “do not fetch rows and reduce them yourself” instruction, since catalog summaries use the first sentence.
- Be conservative with `_meta.formula_ref`: only link the z-score formula where that value is actually emitted; rely on `_meta.method` for trend slope, histograms/quantiles, correlation, and efforts delta unless a matching resource anchor already exists.
- When adding constructors, keep them safe for `Catalog()`’s nil client path; defer client use to handlers.

## Tests

Not run; this was a plan review only. Reviewed `PROMPT.md`, `STATUS.md`, R013, and the relevant registry/catalog/analyzer helper surfaces.
