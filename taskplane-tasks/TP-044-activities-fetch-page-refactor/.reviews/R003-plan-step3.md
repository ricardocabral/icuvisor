# Plan Review: Step 3 — Tests

Verdict: **Approved with required guardrails**. The Step 3 checklist targets the right acceptance criteria: table-driven boundary coverage, literal `next_page_token` byte-identity, unchanged response metadata, ordering/count assertions, and preserving the existing test suite.

Please tighten the execution so the tests actually protect the refactor rather than just re-covering happy paths:

1. **Do not recapture or recompute tokens after the refactor.** Expected `next_page_token` values must remain literal strings captured in Step 1 from the pre-refactor implementation. Avoid deriving expectations via `encodeActivitiesPageToken` or by parsing/re-encoding tokens in the test.

2. **Cover both helper-level and handler-level invariants.** `fetchActivitiesPage` tests are good for raw activity ordering and token byte identity, but they cannot assert `_meta`. Add/keep handler-level assertions for `_meta.page_size`, `_meta.more_available`, `_meta.include_full`, presence/absence of `_meta.next_page_token`, and activity count/order for the same boundary scenarios or a clearly representative subset.

3. **Make token absence explicit.** For empty and partial terminal pages, assert an empty helper token and, at the handler layer, that `next_page_token` is omitted from `_meta` rather than present as an empty string. For continuation cases, assert the exact token string and `more_available == true`.

4. **Disambiguate the “exact full window” fixture.** The high-risk branch is upstream `len(activities) == fetchLimit`, not merely returned rows equal to requested `page_size`. Ensure the fixture name/data make that distinction clear, and keep asserting that only the requested page-size rows are returned while the token matches the pre-refactor payload.

5. **Exercise identical-timestamp continuation semantics, not just no panic.** The stall case should prove that same-timestamp boundary rows are skipped/consumed exactly as before, the follow-up request with the captured token terminates or returns the expected rows, and the code does not silently drop an eligible activity at the page boundary.

6. **Keep ordering assertions intentional.** Use unsorted upstream fixture order where practical and assert newest-first output plus the current same-timestamp ID tie-breaker. This is especially important because `iteratePages` now separates fetching/sorting from page draining.

7. **Preserve existing focused tests.** The current same-timestamp/window-cap tests around lookahead widening, boundary errors, Strava stubs, and mismatched tokens should remain unchanged unless only mechanical fixture plumbing is needed. New tests should complement them rather than weakening their assertions.

Suggested validation for this step before moving to Step 4: run `go test ./internal/tools -run 'TestFetchActivitiesPageBoundaryGoldenFixtures|TestGetActivitiesPaginationFiltersAndTokenRoundTrip|TestGetActivitiesDoesNotLoopOnSameTimestampFilteredLookahead|TestGetActivities(WidensLookahead|AdvancesPastFullyFiltered|ReturnsTokenWhenFiltered|ErrorsInstead|StopsBefore)'`.
