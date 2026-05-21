# R001 Plan Review — Step 1: Audit analyzer tier placement

**Verdict:** Changes requested before executing Step 1.

The STATUS/PROMPT checkboxes describe the desired outcomes, but there is not enough of an execution plan recorded to make the audit reproducible. Step 1 is small, but it needs to be explicit because later steps may change core/full placement.

## Required plan adjustments

1. **Define the analyzer-family source of truth before auditing.**
   - Use the existing analyzer-family list in `internal/tools/catalog_test.go` (`analyzerFamilyCatalogNames`) and/or the `toolCatalogGroup` analyzer cases in `internal/tools/catalog.go`.
   - Include `get_fitness_projection` in the analyzer-family audit even though its catalog group is currently `fitness`, because existing analyzer activation tests include it.

2. **Audit both construction and effective registration tiers.**
   - Confirm constructors currently wrap analyzer tools with `fullTool(...)`.
   - Confirm the registered/effective tiers via `internal/tools/catalog_tiers_test.go` or a small local inspection of `Catalog()` / `NewRegistryWithOptions(... ToolsetFull ...)`.
   - Record the current tier for each analyzer-family tool in `STATUS.md` Notes or Discoveries.

3. **Incorporate the existing TP-100/KR5 evidence.**
   - `docs/kr5-benchmark.md` already contains a `TP-098 core-promotion evidence` section showing positive net savings for the three candidates:
     - `analyze_trend`: net +149 tokens, raw-stream pulls `1 → 0`
     - `compute_zone_time`: net +30 tokens, raw-stream pulls `1 → 0`
     - `compute_baseline`: net +219 tokens, raw-stream pulls `0 → 0`
   - The Step 1 acceptance note should therefore say that these are the only candidates eligible for Step 3 promotion, and that all other analyzer-family tools remain `full` unless future benchmark evidence is added.

4. **Keep Step 1 as an audit/documentation step only.**
   - Do not change tool tiers or tests in Step 1.
   - Update only `STATUS.md` with the audit table and promotion-gate note; code/test/doc changes belong in later steps.

## Suggested Step 1 deliverable

Add a concise table to `STATUS.md` with columns: `tool`, `family source`, `catalog group`, `constructor/effective tier`, and `promotion eligibility/evidence`. This will give Step 2 and Step 3 a clear handoff.
