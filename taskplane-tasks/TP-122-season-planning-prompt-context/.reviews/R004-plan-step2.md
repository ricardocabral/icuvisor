# Plan Review — Step 2

Verdict: **Approved.**

The Step 2 plan is aligned with the Step 1 design: enhance the existing prompt surface, primarily `weekly_planning`, rather than adding a new `season_planning` prompt or implying an automated ATP/calendar writer. The checklist covers the required implementation areas: prompt text, explicit no-auto-write guardrails, golden fixture updates, and targeted prompt tests.

Implementation cautions:

- Keep the strengthened prompts within the existing terse-prompt constraint (`TestPromptResourceCitationsStayTerse` allows at most 25 newline-separated lines and rejects schema-like prompt text).
- Make the new `weekly_planning` wording explicitly gather race date/priority, active training plan, planned events, current load, recent completion/compliance, and available write capabilities before proposing season/week edits.
- Include fallback wording for tools that may not be exposed in the current toolset (`get_training_plan`, `compute_compliance_rate`) by using `icuvisor_list_advanced_capabilities` and continuing from available reads.
- Add or update at least one non-golden assertion for the new safety-critical language, not only the markdown fixture, so future fixture churn does not silently drop the approval/no-auto-calendar guardrail.
- Do not add a new prompt in this step unless the PRD prompt catalog and prompt count tests are intentionally updated.

Verification run during review: `go test ./internal/prompts` passes.
