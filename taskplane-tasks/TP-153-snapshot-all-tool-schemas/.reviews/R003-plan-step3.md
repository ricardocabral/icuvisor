# Review R003 — Plan Step 3

Verdict: approved with guardrails.

The Step 3 plan matches the task: regenerate via the existing no-network snapshot command, inspect the JSON churn, and rerun targeted schema/catalog tests after the refresh.

Required checks while executing:

- Confirm the final committed snapshot set matches the Step 1/2 policy: full-mode, coach-enabled coverage of all 60 public tools, with no exclusions.
- Review `git diff -- internal/tools/schema_snapshot` carefully. Existing snapshot changes should be explainable by the new broad/coach policy (especially injected `athlete_id`) or by newly covered tools, not unrelated public schema edits.
- Verify determinism, ideally by running `go run ./scripts/snapshot_tool_schemas.go` twice and ensuring the second run leaves no diff.
- Search the generated JSON for secrets/local paths/nondeterministic values before proceeding.
- Run the targeted tests after generation, including the schema freshness guard.

No plan blocker. Do not weaken the structural coverage checks to reduce snapshot churn; if a noisy field is intentionally tolerated, document the policy as the prompt requires.
