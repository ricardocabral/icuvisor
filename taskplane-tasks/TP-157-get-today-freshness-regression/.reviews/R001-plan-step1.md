# Plan Review: Step 1 — Add explicit freshness regression tests

Verdict: **approved with clarifications**.

The Step 1 scope is appropriate: add focused `get_today` tests in `internal/tools/get_today_test.go` using the existing fake client and fixed clock. The planned cases directly cover the reported stale-composition failure.

Required clarifications before marking Step 1 complete:

- Make the assertions output-level, not only fetch-param-level. Existing tests already assert today-bounded calls; the new regression must prove stale rows returned by fake clients are excluded from `fitness`, `wellness`, `completed_activities`, `planned_events`, `annotations`, and `_meta.section_counts`.
- Include explicit previous-day IDs/dates in each fake source and assert those IDs/dates are absent from shaped output.
- For wellness, cover both:
  - no today wellness row + yesterday row returned => `wellness` is empty, not backfilled;
  - partial today wellness row + richer yesterday row returned => only today’s partial fields appear.
- Keep the test clock/timezone pinned so `today`, `_meta.date`, `_meta.as_of_date`, and row dates are deterministic in athlete-local time.
- If the new tests fail against current code, that is acceptable evidence for Step 2; record the failing targeted test result in `STATUS.md` and do not weaken the regression to make Step 1 green prematurely.

No production-code or changelog changes are needed for Step 1 unless the worker intentionally rolls into Step 2.
