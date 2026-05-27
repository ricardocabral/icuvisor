# Plan Review: Step 2 — Add golden tests

Verdict: **Approved for implementation**

I reviewed `PROMPT.md`, `STATUS.md`, and the existing prompt golden-test pattern in `internal/prompts/catalog_test.go`. The Step 2 plan matches the task scope and can proceed.

Implementation notes:

- Add `internal/prompts/testdata/weekly_review.md` for the **default/no-argument** rendering, since the task explicitly asks for default weekly-review coverage.
- Add `weekly_review` to `TestRenderedPromptsGolden` using `WeeklyReviewPrompt()` with nil or empty arguments, so the golden validates the default scope text.
- Add a small explicit-arguments test for `week_start`, `lookback_days`, and `include_next_week` that asserts the rendered `Scope:` contains those values in argument order. This does not need a second golden unless you prefer one.
- Make sure the golden captures the advanced-capability fallback guidance (`icuvisor_list_advanced_capabilities`), wellness staleness/provenance caution, athlete-local timezone/date guidance, and the no write/delete-without-approval guardrail.
- Run `go test ./internal/prompts` after updating the golden.

No blockers found.
