# Plan Review: Step 2 — Fill regression or docs gaps

Verdict: APPROVE

The revised Step 2 plan now covers the prior blockers:

- It explicitly calls for default/no-`include_full` tag-preservation regressions for both present and explicit empty tags in `get_activities` and `get_activity_details`.
- It keeps the change focused on regression tests/docs, avoiding behavior changes when the implementation is already correct.
- It includes the required `CHANGELOG.md` update under `[Unreleased]`.
- It preserves the nutrition field contract by keeping disambiguated `_g` fields in terse output and raw upstream names only where appropriate.
- It retains the targeted `go test ./internal/tools` gate for this step.

Non-blocking reminder: if no tool schema/catalog output changes are made, `web/content/reference/tools.md` likely only needs to be reviewed as unaffected rather than edited.
