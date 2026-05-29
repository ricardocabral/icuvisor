# Review R006 — Plan review for Step 3

**Verdict:** REVISE

The Step 3 checklist is pointed in the right direction, but it is not safe to proceed while Step 2 carry-over failures remain and the docs targets are still too implicit.

## Required changes

1. **Resolve Step 2 blockers before Step 3 docs work.** `go test ./internal/mcp ./internal/prompts` currently fails because `internal/mcp/protocol_test.go` still expects six prompts. Also, `STATUS.md` records R004/R005 as approvals even though the review artifacts say `REVISE`. Fix these or explicitly make them the first Step 3 precondition; otherwise `make test` cannot pass.

2. **Name the docs pages that must change.** The plan should explicitly include `web/content/reference/resources-prompts.md` for the new `plan_health_review` prompt row/arguments/tool workflow, and likely `web/content/cookbook/prompt-library.md` for a copyable plan-health prompt. Keep the existing cookbook artifacts (`weekly-review.md`, `season-and-block-plan.md`, `fitness-projection.md`) in scope.

3. **Make the cookbook content contract concrete.** Require the docs to distinguish: `weekly_review` = retrospective weekly closeout/preview, `plan_health_review` = current planned-vs-completed/race-risk audit, `season-and-block-plan` = 8+ week design/scheduling workflow. Include the Step 1 guardrails: deloads are intentional unless evidence says otherwise, missing wellness/readiness data must be caveated, race dates are scenario anchors when not found in events, no opaque score, and no calendar writes without reviewed approval.

4. **Use targeted verification for the actual blast radius.** In addition to any docs build (`make web-build`, if Hugo is available), run at least `go test ./internal/prompts ./internal/mcp` after the prompt-reference updates. `make test` remains fine as a heavier gate, but the plan should not rely on Step 4 to discover the known MCP prompt-count failure.
