# Plan review — Step 3: Register tool and tests

**Verdict: Approved**

The revised Step 3 plan addresses the blocking gaps from R013. It now covers all registration surfaces that matter for this repository:

- `registryBaseTools` registration as a `full` read tool near `get_activity_streams`.
- Shared `internal/toolcatalog` known/athlete-scoped catalog updates so `defaultRegistry.Register` and coach ACL validation will accept the new tool.
- Catalog grouping/tier invariants, including adding the `analyzers` group and removing only `compute_activity_segment_stats` from the analyzer ghost assertions.
- Observable full-only assertions through registered tool tier, `Catalog()` tier, and advanced-capabilities/catalog summary surfaces.
- Handler-level tests for time and distance segments, missing streams, insufficient samples, terse/full behavior, canonical/narrow stream requests, mandatory analyzer `_meta`, `source_tools`, and formula-ref policy.

No blocking plan issues remain. Proceed with Step 3 implementation.

## Non-blocking notes for implementation

- When adding the `analyzers` group, make sure any generated/reference docs in Step 4 can render the new group cleanly; Step 3 only needs the catalog tests updated.
- Prefer table-driven handler cases for the time/distance/missing/insufficient paths so expected upstream `Types` and response `_meta` assertions stay close to the fixture data.
- Include an explicit assertion that the default response omits `series`/raw audit samples, while `include_full:true` includes only sliced audit inputs for the selected stat.
