# R003 Plan Review — Step 2: Enforce placement in tests

**Verdict:** REVISE

The Step 2 entry in `STATUS.md` still only repeats the task checkboxes. That is not enough of an implementation plan for a tier-policy test change, especially because Step 3 may legitimately promote a subset of analyzers to `core` based on the KR5 evidence already recorded in `docs/kr5-benchmark.md`.

## Required plan adjustments

1. **Name the exact tests/files to change.**
   - Use `internal/tools/catalog_tiers_test.go` for effective registered tier assertions.
   - Use `internal/tools/list_advanced_capabilities_test.go` for discoverability of hidden full-only analyzers.
   - Reuse `analyzerFamilyCatalogNames()` from `internal/tools/catalog_test.go` so `get_activity_histogram` and `get_fitness_projection` stay covered even though their catalog groups are not both `analyzers`.

2. **Add an analyzer-specific tier-policy test, not only the existing giant map.**
   - The existing `TestRegisteredToolTierMembership` already catches current tiers, but failures there are broad and easy to update mechanically.
   - Plan a dedicated test/subtest that iterates every analyzer-family name and asserts the current Step 2 default is `safety.ToolsetFull`.
   - If Step 3 later promotes tools, the plan should say that this analyzer-specific test will be updated intentionally so only the benchmark-eligible candidates can be `core`.

3. **Encode the promotion gate separately.**
   - Add a small policy test or documented test fixture for the only Step 3 promotion candidates:
     - `analyze_trend`
     - `compute_zone_time`
     - `compute_baseline`
   - The test should ensure this candidate set is a subset of the analyzer family and that no non-candidate analyzer can be promoted by accident. A short test-local note referencing `docs/kr5-benchmark.md` is appropriate here because it explains the non-obvious policy source.

4. **Strengthen advanced-capabilities coverage.**
   - `TestListAdvancedCapabilitiesOutputFromCatalog` currently spot-checks `compute_activity_segment_stats` but does not prove all hidden analyzer-family tools are advertised.
   - Plan to assert that all analyzer-family tools whose effective tier is `full` appear in `icuvisor_list_advanced_capabilities` output, with clear one-line summaries/activation hints.
   - Keep the assertion compatible with Step 3 by deriving the expected advanced-capability rows from the actual full-only analyzer names, rather than assuming a promoted core analyzer remains hidden.

5. **State the verification command for this step.**
   - At minimum: `go test ./internal/tools -run 'TestRegisteredToolTierMembership|TestCatalogIncludesFullAnalyzers|TestCatalogAnalyzerActivationHints|TestListAdvancedCapabilitiesOutputFromCatalog'`.
   - If new test names differ, include them in the run pattern.

## Non-blocking notes

- Step 2 should remain test/policy enforcement only; do not promote analyzer tools in this step.
- A changelog entry is not needed for test-only changes, but `STATUS.md` should record the new tests and any intentional handoff to Step 3.
