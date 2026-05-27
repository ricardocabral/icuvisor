# Plan Review: Step 3 — Regression tests and docs

**Verdict: Approved**

The revised Step 3 plan now addresses the gaps called out in R005. It explicitly includes:

- auditing event/activity coverage for present/order, explicit empty arrays, missing/null omission, malformed omission, and `include_full` raw preservation;
- updating catalog/schema/doc golden expectations affected by the activity tag description/schema changes;
- updating `CHANGELOG.md` under `[Unreleased]`;
- running `go test ./internal/tools ./internal/intervals`.

I reran the targeted command and confirmed the known current failure remains in `TestCatalogSummariesUseFirstDescriptionSentence` for `get_activities`. That is acceptable for the plan stage because the revised checklist now explicitly covers fixing the catalog/doc golden expectation before Step 3 completion.

Suggested execution detail: while resolving the failure, use the narrow test `go test ./internal/tools -run TestCatalogSummariesUseFirstDescriptionSentence` before rerunning the full targeted package command.
