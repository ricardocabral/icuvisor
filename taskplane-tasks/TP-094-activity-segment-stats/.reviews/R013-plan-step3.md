# Plan review — Step 3: Register tool and tests

**Verdict: Changes requested**

The Step 3 plan covers the main implementation direction: add `compute_activity_segment_stats` to the full tool registry, remove it from analyzer ghost assertions, and exercise handler-level success/error/terse/full cases with analyzer `_meta`. That is the right shape for this step.

Before coding, expand the plan to include the catalog surfaces that are required for registration to work and for existing invariants to keep passing.

## Blocking issues

1. **Add the shared tool catalog update explicitly.**  
   `defaultRegistry.Register` rejects any tool whose name is not known to `internal/toolcatalog` (`toolcatalog.IsKnownTool`). The current Step 3 plan only mentions `registryBaseTools` and `internal/tools/catalog.go`; registering the tool without also adding `compute_activity_segment_stats` to `internal/toolcatalog/catalog.go` will fail at runtime/tests with “not present in shared tool catalog.” Because this is an athlete-scoped read/analyzer tool, the plan should specify adding it to `athleteScopedToolNames`/known names and updating the related toolcatalog tests/ACL expectations as needed.

2. **Spell out catalog/tier invariant updates.**  
   Removing the ghost assertion is necessary but not sufficient. The plan should also call out updating the full-tool tier expectations in `internal/tools/catalog_tiers_test.go` and the catalog grouping expectations in `internal/tools/catalog_test.go`. If the new catalog group is intended to be `analyzers`, add that group to the allowed group set and map `compute_activity_segment_stats` to it; if it is intended to live under `activities`, state that explicitly. Leaving this as “as needed” risks either an empty group (catalog descriptor test failure) or an undocumented grouping choice.

3. **Assert full-only placement through the public catalog path, not only constructor usage.**  
   The plan says “Register in `full`,” but the tests should lock the observable behavior: the registered tool has `EffectiveToolset() == full`, `Catalog()` reports tier `full`, and the raw-stream-exception description/summary is discoverable through the catalog/advanced-capabilities surface. This is especially important because analyzer toolset placement is called out in PRD §7.2.C and TP-098 will build on this distinction.

## Non-blocking implementation notes

- Keep the handler fixtures/table cases the plan already lists, but include one assertion that a terse successful response has no `series`/raw samples while the `include_full:true` case includes only the sliced audit inputs.
- For formula metadata, test both sides of the policy: decoupling/drift include their formula refs, while NP/IF do not invent refs until the formula resource is extended.
- Consider making the time and distance success cases table-driven with explicit expected upstream `Types`; this keeps the “canonical streams only, no broad defaults” guarantee from Step 2 protected at the tool boundary.
