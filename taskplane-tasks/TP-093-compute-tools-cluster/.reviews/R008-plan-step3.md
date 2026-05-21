# R008 plan review — Step 3: Register and document activation hints

Verdict: REVISE

The Step 3 intent is correct, but it should not proceed from the current plan/status. Registration would expose the new compute tools as public MCP tools while the review artifacts still show unresolved Step 2 correctness blockers, and the Step 3 plan is too vague about the catalog/doc/test surfaces that must change when these tools become visible.

## Blocking findings

### 1. Step 3 advances past unresolved prior REVISE reviews

- Location: `taskplane-tasks/TP-093-compute-tools-cluster/STATUS.md` reviews table vs `.reviews/R005-code-step1.md`, `.reviews/R006-plan-step2.md`, `.reviews/R007-code-step2.md`
- Severity: High

`STATUS.md` currently records R005/R006/R007 as approvals and marks Step 2 complete, but the actual review files all say `REVISE`. R007 lists public-behavior bugs in the implementation, including compliance linked-reservation conflicts, weekly baseline bucketing, sport-filtered summary baselines, event truncation metadata, and load-balance training-load totals.

Step 3 is registration/documentation, so proceeding before those are resolved would publish known incorrect behavior. The plan needs an explicit prerequisite: fix or supersede the prior REVISE reviews with an approving code review before registering the tools.

### 2. The Step 3 plan omits the catalog/test surfaces that must change with registration

- Location: `internal/tools/catalog.go`, `internal/tools/catalog_test.go`, `internal/tools/catalog_tiers_test.go`, `internal/toolcatalog/catalog.go`, `internal/toolchecks/schema_stability.go`
- Severity: High

Registering the four tools is more than appending constructors in `registryBaseTools`. The current catalog tests still treat `compute_zone_time`, `compute_load_balance`, `compute_baseline`, and `compute_compliance_rate` as analyzer-family ghosts, and the tier test has an exhaustive expected map. The public tool-name catalog used by coach ACLs/tool checks also does not include these names yet.

The plan should explicitly include:

- add the four constructors to the registry with `fullTool`/`ToolsetFull` and read capability semantics;
- map all four names to the `analyzers` group in `toolCatalogGroup`;
- remove only these four tools from the ghost assertion while leaving unimplemented analyzer ghosts such as `analyze_trend` unregistered;
- update exhaustive tier/catalog tests to expect the new full analyzers;
- update canonical tool-name/schema-stability surfaces if registration makes them part of ACLs or generated schema docs.

Without these planned edits, Step 3 is likely to either fail tests or leave registered tools missing from secondary catalog surfaces.

### 3. Activation hints need to match the PRD pattern, not just mention streams

- Location: `docs/prd/PRD-icuvisor.md` §7.2.C activation hint pattern; `internal/tools/compute_*` descriptions
- Severity: Medium

The task says to “document activation hints,” and the PRD says analyzer descriptions should lead with the user-prompt shape that should trigger the tool plus an explicit “do not roll your own” line. The current Step 3 checklist only says descriptions must say not to fetch rows/streams and reduce manually. That is necessary but not sufficient for the activation hint contract.

Revise the plan to check each new tool description for both parts:

- a “Use this when the user asks …” trigger phrase tailored to zone time/load balance/baseline/compliance; and
- an explicit “Do not pull/fetch rows or streams and reduce manually” instruction.

This also gives a concrete way to satisfy the “identifiable for later core promotion” requirement: keep these analyzers grouped/tiered as `full` now, while making `compute_zone_time` and `compute_baseline` easy to find in catalog tests/docs as future core-promotion candidates; do not register `analyze_trend` until it exists.

### 4. Documentation generation/update is underspecified

- Location: `web/content/reference/tools.md`, `CHANGELOG.md`, generated catalog docs
- Severity: Medium

The prompt’s file scope and documentation requirements include the generated tool reference and changelog. Step 3 currently says “document activation hints,” but does not say whether the worker will run the docs generator (`make docs-tools`) or manually update the generated reference consistently with the registered catalog. The plan should state the expected documentation path and defer only broad verification to Step 4.

## Required plan changes

1. Add a Step 3 precondition to resolve R007’s code blockers and correct the STATUS/review log before public registration.
2. Expand Step 3 with concrete registry/catalog/test/doc edits listed above, including preserving `analyze_trend` as unregistered while making the new compute tools visible as `full` analyzers.
3. Tighten the activation-hint requirement to the PRD format: trigger phrase plus “do not roll your own/fetch rows or streams and reduce manually.”
4. Specify how generated docs and CHANGELOG are updated or explicitly assign them to the later verification/delivery step.

I did not run the test suite because this is a plan review.
