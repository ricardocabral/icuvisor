# Plan Review: Step 1 — Characterize current behaviour

Verdict: **Approved with comments**. The Step 1 scope matches the prompt: add pre-refactor characterization coverage for pagination boundaries and pin opaque `next_page_token` values before extracting `iteratePages`.

No blocking issues found, but please tighten the execution details so the tests genuinely protect the refactor:

1. **Pin literal token strings, not recomputed expectations.** The characterization tests should assert the exact `_meta.next_page_token` string captured from the current implementation. Avoid building expected values by calling `encodeActivitiesPageToken`, because that would not catch accidental token byte drift during the refactor.

2. **Disambiguate “exact full window”.** Current `fetchActivitiesPage` fetches with `fetchLimit = page_size*2+1` (capped), so a response with exactly `page_size` eligible results is a terminal page with no token unless the upstream page was also fetch-limit-full. Name the case clearly, e.g. `exact_page_size_terminal`, or add a separate case if you intend to exercise `lastFullWindow == true` (`len(activities) == fetchLimit`).

3. **Make ordering part of the fixtures.** Use intentionally unsorted upstream fixtures where possible and assert newest-first output, including the current tie-breaker on same `start_date_local` by descending activity ID. This protects `sortActivities`/cursor interactions during Step 2.

4. **Model the same-timestamp stall precisely.** The stall case should include multiple same-timestamp rows at the cursor boundary, default unnamed filtering, and a small `page_size`, then assert both the returned rows and terminal/continuation token behavior. Existing tests around same-timestamp filtering are good references; keep them intact and add the byte-identity assertions this task needs.

5. **Record the captured values in the task status.** After adding/running the characterization tests, update `STATUS.md` notes with the captured token values or a concise mapping from case name to expected token/empty token. That gives Step 2 a clear baseline.

Suggested validation for this step: run at least `go test ./internal/tools` after adding the characterization tests, before any production-code refactor.
