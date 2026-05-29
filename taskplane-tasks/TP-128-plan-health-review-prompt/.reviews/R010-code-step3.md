# Review R010 — Code review for Step 3

**Verdict:** REVISE

## Findings

1. **Weekly-review recipe now asks for evidence it does not fetch.** In `web/content/cookbook/weekly-review.md:38-42`, the copy-paste prompt tells the assistant to treat planned deload/recovery weeks as intentional unless “compliance, wellness, or form” data shows a problem. But this recipe’s pull list only fetches profile, fitness, summary, activities, zone time, and load balance; it does not fetch planned events/training plans, `compute_compliance_rate`, or wellness data. A client following the recipe literally must either ignore the new caveat or reason without evidence, which muddies the intended distinction between `weekly_review` and `plan_health_review`. Move this caveat to the plan-health variation/callout, or add the required planned/compliance/wellness reads if weekly review is meant to make that claim.

2. **Season-plan page points to a non-existent “copy below.”** `web/content/cookbook/season-and-block-plan.md:15` says users can use “the prompt-library copy below,” but there is no plan-health copy later on that page. Please link to `prompt-library` or reword this as “the prompt-library copy-paste prompt” so the cookbook guidance does not send readers looking for missing content.

## Verification

- `go test ./internal/prompts ./internal/mcp`
- `make web-build` (passes; Hugo emits existing deprecation warnings)
