# Plan Review: Step 1 — Add `weekly_review` prompt registration

Verdict: **Needs revision before implementation**

The STATUS update acknowledges the R001 topics, but the Step 1 plan is still not concrete enough to guarantee a passing intermediate change.

## Required revisions

1. **Make the Step 1 file scope explicit.**
   Step 1 cannot be `internal/prompts/catalog.go` only. Registration requires adding the prompt to `NewRegistry()` in `internal/prompts/registry.go`, and targeted prompt tests will require at least the non-golden registration/resource-citation expectations in `internal/prompts/catalog_test.go` to be updated in the same step.

2. **Resolve the Step 1 vs Step 2 fallback split.**
   The prompt text is created in Step 1, but advanced-capability fallback guidance is currently listed under Step 2. If `compute_load_balance` or other analyzer helpers are included, the Step 1 prompt registration should already include `icuvisor_list_advanced_capabilities` guidance/tool listing; Step 2 should only golden-test that text.

3. **List the intended tool sequence, not just “deliberate analyzer/compute tool set.”**
   The plan should name the tools the prompt will advertise, e.g. profile, fitness, training summary, activities, zone-time/load-balance analyzers, optional trend analysis, and any events/training-plan reads needed for next-week preview. This avoids accidentally shipping a duplicate of `weekly_planning` or omitting the analyzer/compute dependencies from the task.

4. **State exactly which targeted tests Step 1 will keep passing.**
   Since the existing tests assert the registered prompt count and iterate a fixed prompt list, the plan should say Step 1 updates those expectations and runs `go test ./internal/prompts`. Golden coverage can remain Step 2.

Once these details are added to STATUS or the step plan, the implementation path is straightforward.
