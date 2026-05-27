# Plan Review: Step 1 — Add `weekly_review` prompt registration

Verdict: **Needs revision before implementation**

I reviewed `PROMPT.md`, `STATUS.md`, and the existing prompt registry/test patterns. The high-level intent is sound, but the Step 1 plan as written is incomplete in a few places that are likely to cause a broken intermediate state.

## Required plan adjustments

1. **Include `internal/prompts/registry.go` in the Step 1 artifact list.**
   `NewRegistry()` registers prompts from `registry.go`, not `catalog.go`. Adding `WeeklyReviewPrompt()` to `catalog.go` alone will not register the prompt. Step 1 should explicitly add the new prompt to the `staticRegistry{entries: ...}` list.

2. **Account for test updates or expected failures in Step 1.**
   `catalog_test.go` currently asserts exactly five registered prompts and iterates a fixed prompt list in `TestPromptResourceCitationsStayTerse`. Once `weekly_review` is registered, targeted prompt tests will fail unless the tests are updated in the same step. If the task wants golden coverage deferred to Step 2, Step 1 should still update the count/name registration test, or explicitly state that targeted tests cannot pass until Step 2.

3. **Define the tool set deliberately.**
   The prompt should not be only a duplicate of `weekly_planning`. The task depends on analyzer/compute tools, and the existing cookbook recipe points to `get_athlete_profile`, `get_fitness`, `get_training_summary`, `get_activities`, `compute_zone_time`, `compute_load_balance`, and optionally `analyze_trend`. If any of those are full-toolset-only, include `icuvisor_list_advanced_capabilities` fallback guidance as part of the prompt text.

4. **Make weekly-review-specific guardrails explicit.**
   The existing default guardrails cover API keys and terse responses, but the task requires no write/delete actions without explicit user approval. Step 1 should either set custom `Guardrails` for `weekly_review` that preserve the defaults plus this write/delete rule, or add an instruction in `Do:` that is hard to miss.

## Content checks for implementation

- Add a `WeeklyReviewName = "weekly_review"` constant and a `WeeklyReviewPrompt()` factory matching existing naming/style.
- Include arguments only if they match current conventions: `week_start`, `lookback_days`, and `include_next_week` should be optional string arguments with clear athlete-local date/boolean wording.
- Include athlete-local date/timezone guidance before day-to-day comparisons.
- Cover planned-vs-completed comparison and wellness staleness, per completion criteria.
- Keep rendered text terse enough to satisfy `TestPromptResourceCitationsStayTerse` (currently max 25 newlines).

With these revisions, the step should be safe to implement and will align with the existing prompt registry patterns.
