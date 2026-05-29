# Review R007 — Code review for Step 2

**Verdict:** REVISE

## Findings

1. **Plan-health prompt drops the default terse-payload guardrail.** `internal/prompts/catalog.go:163-168` supplies custom guardrails, which bypasses `renderSpec`'s default guardrail that says to prefer terse tool responses and use `include_full` only when needed. This prompt lists potentially large/read-heavy tools (`get_activities`, `get_training_plan`, wellness reads, projection/analyzer tools), so it should preserve the terse-by-default instruction just like `weekly_review` does. Please add the terse/include_full guardrail and cover it in the plan-health prompt test/golden.

2. **Status notes still contradict the review artifacts.** `STATUS.md:124-125` records R004/R005 as `APPROVE`, but `.reviews/R004-plan-step2.md` and `.reviews/R005-code-step2.md` both have `REVISE` verdicts (the review table above is correct). Please fix the Notes/Execution-log entries so the task history is internally consistent.

## Verification

- `go test ./internal/prompts ./internal/mcp` passes.
- `go test ./...` passes.
