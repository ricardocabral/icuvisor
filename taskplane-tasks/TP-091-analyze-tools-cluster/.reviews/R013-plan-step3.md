# R013 Plan Review — Step 3: Register tools and descriptions

**Verdict:** REVISE

The current Step 3 plan is still just the original three outcome checkboxes in `STATUS.md`. For this review-level 3 task, that is not enough to proceed: Step 2 left the public tool adapters, request validation, source reads, schema wiring, and registry/catalog integration for Step 3, and those details determine whether the new analyzer family is actually callable and coach-safe.

## Blocking findings

1. **Step 3 needs a concrete tool-adapter implementation plan, not only “add tool files.”**

   The status notes say the four tool files are deferred to Step 3 (`STATUS.md`, Step 2 implementation plan), but the Step 3 plan does not say how each handler will be built. Please hydrate the plan with the package/file-level sequence for:

   - `internal/tools/analyze_trend.go`, `analyze_distribution.go`, `analyze_correlation.go`, and `analyze_efforts_delta.go` constructors, request structs, strict decoders, input schemas, and output schemas;
   - shared adapter helpers, if any, for loading current/baseline metric series from `FitnessClient`, `WellnessClient`, `ActivitiesClient`, and the curve clients;
   - how handlers call `analysis.ComputeTrend`, `ComputeDistribution`, `ComputeCorrelation`, and `ComputeEffortsDelta` and then wrap with `encodeAnalyzerResponse` / `analysis.NewAnalyzerMeta`;
   - the user-facing invalid-argument/fetch-error messages and context-cancellation behavior.

   Without this, the step could register names but still omit the source-loading and response-shaping behavior promised by Steps 1–2.

2. **The registration plan must include the shared tool catalog and coach-scoped catalog, not only `registryBaseTools`.**

   The registry rejects unknown tools through `toolcatalog.IsKnownTool` (`internal/tools/registry.go:110-113`). New analyzer names therefore must be added to `internal/toolcatalog/catalog.go` as known, athlete-scoped tool names before `registryBaseTools` can register them. Because coach mode uses `toolcatalog.IsAthleteScopedTool` to inject/route `athlete_id`, omitting these tools from `athleteScopedToolNames` would either fail registration or expose tools that cannot be correctly coach-routed.

   Please explicitly plan to add constants and athlete-scoped entries for:

   - `analyze_trend`
   - `analyze_distribution`
   - `analyze_correlation`
   - `analyze_efforts_delta`

   Then add them to `registryBaseTools` as `fullTool`/read tools and update `toolCatalogGroup` so `Catalog()` reports them under `analyzers`. Also plan the required catalog-test changes: the current `TestCatalogMatchesRegistryAndPRDRegisteredTools` has these four names in the “analyzer-family ghost” list, so it must be updated when the tools intentionally become registered.

3. **The plan does not pin the descriptions and formula metadata strongly enough for the PRD activation-hint contract.**

   PRD §7.2.C requires analyzer descriptions to lead with the prompt shape that should activate the tool and to explicitly tell the LLM not to fetch rows and reduce them itself. Because `Catalog()` and `icuvisor_list_advanced_capabilities` summarize tools from the first description sentence, the plan should specify whether the first sentence itself will carry both the activation hint and the “do not roll your own reductions” instruction, or otherwise add tests to guarantee the full descriptions contain both. Please also enumerate the intended `_meta.formula_ref` policy per tool so implementation does not over-claim formula references:

   - trend baseline z-score deltas may link `icuvisor://analysis-formulas#z_score` when emitted;
   - correlation/trend slope/histogram should rely on `_meta.method` unless the existing resource has canonical anchors;
   - efforts delta should use `best_efforts_current_vs_baseline` source/method metadata and family-specific `source_tools`.

4. **The plan should cover full-toolset gating and public catalog/docs fallout.**

   PRD §7.2.C places this family in `full` by default. Please make the plan explicit that these four tools are `RequirementRead`, `ToolsetFull`, appear in `icuvisor_list_advanced_capabilities` when the core toolset is active, and do not promote anything to `core` in this task. Also record whether `make docs-tools` / `web/content/reference/tools.md` and `CHANGELOG.md` are intentionally deferred to Step 4/6, so the Step 3 code does not accidentally leave the generated catalog stale without a tracked follow-up.

## Tests

Not run; reviewed `PROMPT.md`, `STATUS.md`, prior reviews, and the relevant registry/catalog/toolcatalog/analyzer helper surfaces.
