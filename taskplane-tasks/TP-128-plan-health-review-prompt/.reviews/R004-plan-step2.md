# Review R004 — Plan review for Step 2

**Verdict:** REVISE

Step 1 produced a solid prompt contract in `STATUS.md` Discoveries, but the Step 2 implementation plan is still too generic for the changes required by that contract. Please hydrate Step 2 before coding so the worker cannot miss registry/catalog and contract-specific test work.

## Required changes

1. **Add registry work to the Step 2 plan.** The current artifacts omit `internal/prompts/registry.go`, but adding `plan_health_review` requires adding the prompt to `NewRegistry()` and preserving intended prompt order. The existing test is still named/counting six prompts (`TestNewRegistryRegistersSixPrompts`), so the plan should explicitly update it to seven and include `PlanHealthReviewName` in the expected list.

2. **Make golden-test updates concrete.** Step 2 should explicitly add `internal/prompts/testdata/plan_health_review.md`, add the prompt to `TestRenderedPromptsGolden`, and include it in `TestPromptResourceCitationsStayTerse` / any required-args helper as needed.

3. **Add contract invariant tests, not only a golden fixture.** The plan should require targeted assertions that the rendered prompt includes the Step 1 guardrails: `icuvisor://analysis-formulas`, analyzer `_meta.method` / `_meta.assumptions` / `_meta.formula_ref`, advanced-tool fallback via `icuvisor_list_advanced_capabilities`, no opaque aggregate/black-box score, deload/recovery-week caveat, race-date scenario-anchor behavior when no race event is found, and no calendar writes until the exact proposal is shown and approved.

4. **Keep prompt scope bounded and terse.** The plan should state that this is a review workflow only, not plan filler or autonomous coaching, and that the rendered prompt must still satisfy the existing terseness/resource-citation expectations.

After these additions, the targeted verification command `go test ./internal/prompts` is sufficient for this step.
