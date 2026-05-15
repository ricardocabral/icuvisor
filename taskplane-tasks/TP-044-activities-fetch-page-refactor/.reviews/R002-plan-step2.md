# Plan Review: Step 2 — Extract `pageCursor` + `iteratePages`

Verdict: **Approved with required guardrails**. The planned direction matches TP-044: keep the refactor local to `internal/tools/get_activities.go`, introduce unexported state/driver helpers, and make `fetchActivitiesPage` a thin response-assembly shell without changing tool schema, token format, ordering, or page-size defaults.

The main risk is semantic drift in cursor advancement. Please lock down the driver contract before editing code:

1. **Do not advance past an unreturned eligible activity.** In the current loop, when `len(page) == args.PageSize`, it returns a token for the cursor state after the last activity already appended/skipped, not after the next candidate in the upstream slice. If `iteratePages` pre-advances as it yields, it can silently drop the first activity of the next page.

2. **Keep unnamed filtering tied to cursor advancement.** Default unnamed rows are not returned, but they are consumed by `advanceCursorPast` unless `isStravaBlocked(activity)` is true. The new driver/caller split must preserve that exact distinction.

3. **Preserve the no-candidates branch exactly.** The current behavior for `len(candidates) == 0` includes: try `advanceCursorPast` on the last upstream row; widen `fetchLimit` to `maxActivityFetchLimit` only for a full, non-advancing window; return `errActivitiesPaginationBoundary` after the widened full window still cannot advance; then fall back to `advanceCursorBeforeBoundary`. This is the highest-risk part to hide inside `pageCursor` methods.

4. **Keep token byte identity by encoding the same payload.** `pageCursor` may own an `activitiesPageToken`, but final encoding should still marshal `activitiesPageToken` with the same field values, field ordering, nil-vs-empty behavior, and `Fields` copy semantics. In particular, copying token fields on resumed pages and setting `terseActivityFields` only when `!IncludeFull` must remain unchanged.

5. **Make token emission conditions explicit.** Preserve the existing cases: empty/partial upstream pages are terminal; full returned pages emit a token only if the cursor advanced; after exhausting the fetch loop, emit a token only when the last upstream window was full and the cursor advanced.

6. **Keep `fetchLimit` state with the cursor/driver.** It starts as `min(page_size*2+1, maxActivityFetchLimit)` and only widens under the same-timestamp/no-advance condition. Do not reset it on each driver iteration.

7. **Prefer a concrete unexported driver over a generic abstraction.** A small local type with methods such as `listParams`, `advancePast`, `advanceBeforeBoundary`, `nextToken`, and an `iteratePages` callback/iterator is enough. Avoid introducing reusable pagination machinery outside this tool.

Also, Step 1 appears to have added golden tests in `get_activities_test.go`; before/while implementing Step 2, update `STATUS.md` notes with the captured token baseline as requested in R001 so the refactor has an auditable checkpoint.

Suggested validation immediately after Step 2: `go test ./internal/tools -run 'TestFetchActivitiesPageBoundaryGoldenFixtures|TestGetActivitiesPaginationFiltersAndTokenRoundTrip|TestGetActivities(WidensLookahead|ErrorsInstead|StopsBefore)'` before moving on to the full Step 4 suite.
