# R004 Plan Review — Step 2: Enforce placement in tests

**Verdict:** APPROVE

The updated Step 2 plan in `STATUS.md` addresses the prior review. It now names the relevant test files, plans dedicated analyzer-family tier coverage using `analyzerFamilyCatalogNames()`, separates the benchmark-gated promotion candidate set, and accounts for `icuvisor_list_advanced_capabilities` discoverability by deriving expectations from effective full-only analyzer tools.

## What looks good

- Uses `internal/tools/catalog_tiers_test.go` for effective registered tier policy rather than relying only on catalog descriptor metadata.
- Reuses `analyzerFamilyCatalogNames()`, which keeps `get_activity_histogram` and `get_fitness_projection` covered despite their non-`analyzers` catalog groups.
- Adds a separate promotion-candidate policy check for only:
  - `analyze_trend`
  - `compute_zone_time`
  - `compute_baseline`
- Keeps Step 2 as test/policy enforcement only; actual promotion remains Step 3.
- Plans advanced-capability coverage in `internal/tools/list_advanced_capabilities_test.go` for the effective full-only analyzer set, which should remain compatible if Step 3 promotes only the eligible candidates.

## Minor implementation guidance

- Give the new tests explicit names when implementing, for example `TestAnalyzerFamilyDefaultsToFullToolset`, `TestAnalyzerCorePromotionCandidates`, and/or a focused advanced-capabilities test. This will make the Step 2 verification command precise and greppable.
- In the promotion-candidate test, prefer a small test-local note referencing `docs/kr5-benchmark.md` over parsing benchmark markdown from the unit test. The durable invariant is the allowed candidate set and its subset relationship to the analyzer family.
- Record the targeted test command/result in `STATUS.md` after implementation. Include the new test names in the `-run` pattern.

No blocker to proceed.
