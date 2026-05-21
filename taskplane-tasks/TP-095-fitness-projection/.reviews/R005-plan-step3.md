# R005 Plan Review — Step 3: Register and test

**Verdict:** Needs changes / not yet reviewable

I read `PROMPT.md`, `STATUS.md`, the prior reviews, and the current projection implementation. The Step 3 plan currently consists only of the three high-level checklist bullets in `STATUS.md`. That is not enough for this step because registration touches multiple catalog/ACL/toolset invariants, and the tests need to lock down the public analyzer contract before `get_fitness_projection` becomes visible.

## Required plan details before approval

1. **Registration blast radius**
   - State every registration/catalog file that will change. Adding `newGetFitnessProjectionTool` to `registryBaseTools` is not sufficient: `Register` rejects tools that are absent from `internal/toolcatalog`, so the plan must include adding a canonical `GetFitnessProjection` name and athlete-scoped ACL eligibility there.
   - Decide and document the public catalog grouping/tier. The constructor currently uses `fullTool`; if the tool remains full-tier, update `catalog_tiers_test.go`. If it is grouped under `fitness`, update `toolCatalogGroup`; if a new analyzer group is intended, update the allowed catalog groups and docs expectations accordingly.
   - Update catalog tests deliberately: remove `get_fitness_projection` from the analyzer ghost list, assert it is registered/present in `Catalog()`, and keep the PRD/catalog/registry consistency checks meaningful.

2. **Carry forward unresolved Step 2 review findings**
   - The Step 3 plan should explicitly fix or test the R004 findings before registration: `training_load` must not disappear for valid zero-load projected days, and schema/decoder/user-message inconsistencies for horizon and `recovery_week_cadence` must be resolved.
   - Also update `STATUS.md` review tracking. It currently records earlier reviews as approved even though the review files requested changes, so Step 3 implementers do not have an accurate blocker list.

3. **Concrete test matrix**
   - Add `internal/tools/get_fitness_projection_test.go` with no network access, using fake fitness/profile clients with `Raw` values populated for CTL/ATL/TSB.
   - Include deterministic golden or equivalent exact-shape tests for:
     - standard ramp with terse default, proving no `series` is emitted unless `include_full:true`;
     - recovery-week behavior and load source labeling;
     - `include_full:true` full series shape, including day 0 and projected days, and preserving `training_load: 0` when applicable;
     - invalid inputs, including mutually exclusive horizon fields, out-of-range horizon/ramp/recovery settings, unsupported model, bad/duplicate/out-of-horizon planned loads, and the `recovery_week_cadence: 1` schema/decoder boundary once the contract is chosen;
     - insufficient current fitness data: no exact start-date row and rows whose `Raw` lacks numeric `fitness`, `fatigue`, or `form`.
   - Assert mandatory analyzer `_meta` fields on the actual tool response, not only the shared analyzer helper: `method`, `source_tools`, `n`, `missing_days`, `missing_action`, `insufficient_sample`, plus projection assumptions/boundaries such as horizon, ramp, recovery cadence/load pct, planned-load count, and CTL/ATL constants.

4. **Verification scope for this step**
   - Plan to run targeted tests that cover the touched packages, at minimum `go test ./internal/analysis ./internal/tools ./internal/toolcatalog`, before marking Step 3 complete. Full `make test`/`make build`/`make lint` can remain in later steps, but this step should prove registration and the new tests pass together.

## Recommendation

Revise the Step 3 plan in `STATUS.md` or add a short design note referenced from it with the registration inventory, exact test matrix, and explicit handling of the open R004 issues. Once that is in place, implementation should be straightforward to review.
