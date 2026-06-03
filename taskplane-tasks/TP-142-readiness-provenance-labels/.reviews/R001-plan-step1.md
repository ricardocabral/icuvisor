# Plan Review — Step 1: Audit readiness/recovery wording

**Verdict:** Approve with one minor adjustment.

The Step 1 plan matches the task intent: it audits the wellness provenance path, audits prompt wording, records discoveries, and runs the targeted packages. That is appropriate for a Level 1 plan-only review.

## Required adjustment before/while executing Step 1

- Include `internal/prompts/testdata/weekly_review.md` explicitly in the Step 1 audit/artifacts. The task file scope includes it, and `weekly_review` contains readiness/recovery wording that could regress independently of `recovery_check`. The current Step 1 artifact list only names `recovery_check.md`, despite the checklist saying “recovery/weekly prompts.”

## Execution guidance

- When auditing prompt wording, grep broadly in `internal/prompts/catalog.go` and prompt goldens for `readiness`, `recovery`, `freshness`, provider names, and `Body Battery`, then focus fixes/tests on the files in task scope.
- The wellness audit should cover all listed source paths: Garmin Body Battery, Oura readiness, Polar nightly recharge / ANS charge, WHOOP recovery, and generic/unknown upstream `readiness`.
- Keep Step 1 as discovery-oriented unless an obvious typo blocks tests; save regressions/wording changes for Step 2 as planned.
