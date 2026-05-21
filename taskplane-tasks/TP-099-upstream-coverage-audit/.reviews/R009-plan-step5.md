# Review R009 — Plan Review for Step 5

**Verdict:** REVISE

The Step 5 plan is close: it now names the targeted audit command, the expected summary comparison, and the required full gates (`make test`, `make build`, `make lint`). However, it still does not resolve the main ambiguity from R008: whether Step 5 will rerun these commands as the final verification gate or intentionally carry forward the already-recorded Step 4 pass results.

Step 4 already says the exact same commands passed. For Step 5, the plan should be explicit so the completion record is auditable.

## Required change

Add one clear bullet/line under Step 5 before execution, choosing one of these approaches:

- **Preferred:** rerun all Step 5 commands now:
  - `go run scripts/audit_zone_time_coverage.go`
  - `make test`
  - `make build`
  - `make lint`
- **Or:** carry forward Step 4 results only after recording that `git status --short` shows no task-relevant file changes since Step 4 verification.

Then record the chosen approach and command outcomes in `STATUS.md`. If any gate fails, fix task-caused failures; document only clearly unrelated/pre-existing failures with enough output to be actionable.

No analyzer behavior changes are needed for this step.
