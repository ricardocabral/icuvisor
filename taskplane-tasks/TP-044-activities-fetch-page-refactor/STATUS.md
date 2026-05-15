# TP-044-activities-fetch-page-refactor: `fetchActivitiesPage` pagination driver refactor — Status

**Current Step:** Step 2: Extract `pageCursor` + `iteratePages`
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 1
**Review Counter:** 1
**Iteration:** 1
**Size:** S

---

### Step 1: Characterize current behaviour

**Status:** ✅ Complete

- [x] Pin the four boundary cases with golden fixtures (empty, partial, exact full window, identical-timestamp stall)
- [x] Capture pre-refactor `next_page_token` values for byte-identity assertions
- [x] Confirm result ordering captured by fixtures

### Step 2: Extract `pageCursor` + `iteratePages`

**Status:** 🟨 In Progress

- [ ] Introduce `pageCursor` state struct (replaces the four ad-hoc booleans)
- [ ] Introduce `iteratePages` driver yielding candidates one page at a time
- [ ] Reduce `fetchActivitiesPage` to a thin shell
- [ ] No new exported identifiers

### Step 3: Tests

**Status:** ⏳ Not started

- [ ] Table-driven coverage of the four boundary cases
- [ ] Byte-identical `next_page_token` assertions vs. captured fixtures
- [ ] Response shape (`_meta`, ordering, count) unchanged
- [ ] Existing tests pass unchanged

### Step 4: Verify

**Status:** ⏳ Not started

- [ ] `make build`, `make test`, `make test-race`, `make lint`
- [ ] Diff review: function shorter, shallower, lower complexity
- [ ] Manual smoke against a live account (if available) — same tokens, same page contents

---

## Decisions

- **State struct shape:** TBD in Step 2. Default sketch in PROMPT.md: `pageCursor` owns the upstream cursor + the "advanced this iteration" / "full window" flags, replacing `lastFullWindow`, `cursorAdvanced`, `advanced`.
- **Driver signature:** TBD in Step 2. Plain function returning the next page's candidates plus a "done" signal; no generic abstraction.

## Notes

_Add notes as work progresses._

| 2026-05-15 14:25 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 14:25 | Step 1 started | Characterize current behaviour |
| 2026-05-15 14:27 | Review R001 | plan Step 1: APPROVE |
