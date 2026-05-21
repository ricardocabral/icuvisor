# TP-044 — Refactor `fetchActivitiesPage` pagination driver (audit P1)

## Mission

`internal/tools/get_activities.go:295-389` `fetchActivitiesPage` is ~95 lines, ~5 nesting levels, and carries four "did we advance?" booleans (`lastFullWindow`, `cursorAdvanced`, `advanced`, plus the implicit loop-continuation gate). Cyclomatic complexity is high enough that the boundary cases (empty page, partial page, exact full window, identical-timestamp cursor stall) are hard to reason about by reading the function. The 2026-05-15 Go audit flagged this as the top readability issue inside the tools package.

Goal: extract an `iteratePages` driver that yields candidate activities, and lift the "advance cursor / boundary" bookkeeping into a small `pageCursor` state struct. `fetchActivitiesPage` shrinks to a thin shell that wires the driver to the response builder. Pagination semantics must be preserved exactly — `next_page_token` opacity, default page size, and result ordering must be byte-identical to the pre-refactor implementation.

PRD anchors: §7.2.E pagination invariants (`next_page_token` opacity, server-side paging, default page size sized for free-tier context windows).

ROADMAP positioning: maintenance / debt paydown. Independent of any version milestone. Land before v0.5 dogfood so the activities tool — by far the most-exercised — is in its cleaner shape when external users hit it.

Complexity: Blast radius 1 (one file), Pattern novelty 1 (standard Go), Security 1 (no new credential paths), Reversibility 2 (mechanical refactor, easy to revert) = 5 → Review Level 1. Size: S.

## Dependencies

- None. Pure internal refactor of one tool.

## Context to Read First

- `CLAUDE.md` — Go conventions, table-driven tests, no panic outside main, error wrapping with `%w`.
- `internal/tools/get_activities.go` — the file under refactor; `fetchActivitiesPage` at 295-389 is the target.
- `internal/tools/get_activities_test.go` — existing table-driven coverage; new tests must follow the same style and fixtures.
- `internal/intervals/activities.go` — upstream pagination shape (oldest/newest cursors, page size, identical-timestamp behaviour).
- `docs/prd/PRD-icuvisor.md` §7.2.E — pagination invariants.

## File Scope

- `internal/tools/get_activities.go` — extract `pageCursor` state struct and `iteratePages` driver; reduce `fetchActivitiesPage` to a thin shell.
- `internal/tools/get_activities_test.go` — add table-driven coverage of the four boundary cases (empty, partial, exact full window, identical-timestamp stall).
- `CHANGELOG.md` — `[Unreleased]` entry under "Changed".
- `taskplane-tasks/TP-044-activities-fetch-page-refactor/STATUS.md`.

Out of scope:
- Changing the public tool schema or `_meta` shape.
- Renaming the tool.
- Changing the default page size.
- Touching sibling tools or other pagination call sites (file a follow-up if the extracted driver looks reusable).
- Changing the upstream `internal/intervals/activities.go` contract.

## Steps

### Step 1: Characterize current behaviour

- [ ] Identify the four boundary cases and pin them with golden fixtures before changing code: empty page (zero upstream results); partial page (fewer results than `page_size`); exact full window (results exactly equal `page_size`); identical-timestamp stall (multiple upstream rows share the cursor timestamp so a naive advance loops).
- [ ] Capture the current `next_page_token` value for each case and assert byte-identity in the new tests. The token is opaque to callers but must not drift on refactor.
- [ ] Confirm result ordering (newest-first per current behaviour) is captured by the fixtures.

### Step 2: Extract `pageCursor` + `iteratePages`

- [ ] Introduce a `pageCursor` struct that owns: the upstream cursor (oldest/newest timestamps as the current code uses them), the "did the cursor advance this iteration?" flag, and the "is this a full window?" flag. Replace the four ad-hoc booleans.
- [ ] Introduce `iteratePages` (or similar) that consumes the upstream client and yields candidate activities one page at a time, with the boundary bookkeeping localized inside. Caller stays in `fetchActivitiesPage` and consumes the driver until the page-size budget is met or the upstream is exhausted.
- [ ] `fetchActivitiesPage` becomes a thin shell: configure cursor, drain driver, build response.
- [ ] Keep all naming inside `internal/tools/`; do not export new identifiers.

### Step 3: Tests

- [ ] Add table-driven tests for the four boundary cases. Use existing `testdata/` fixture conventions from `get_activities_test.go`.
- [ ] Assert byte-identical `next_page_token` against pre-refactor values captured in Step 1.
- [ ] Assert response shape unchanged (`_meta`, ordering, count).
- [ ] Existing tests must pass unchanged.

### Step 4: Verify

- [ ] `make build`, `make test`, `make test-race`, `make lint`.
- [ ] Eyeball the diff: `fetchActivitiesPage` should be materially shorter and shallower; cyclomatic complexity should drop visibly (a `gocyclo`-style read, not a hard number).
- [ ] Manual smoke against a live athlete account if available: a paged walk through recent activities should produce the same tokens and same page contents as `main` for the same inputs.

## Acceptance Criteria

- `fetchActivitiesPage` is a thin shell over `iteratePages`; cyclomatic complexity drops materially and nesting depth is no greater than 3.
- The four "did we advance?" booleans are replaced by a single `pageCursor` state struct.
- `next_page_token`, `_meta`, result ordering, and default page size are unchanged. Existing snapshot / table tests pass unchanged.
- New table-driven tests cover the four boundary cases (empty, partial, exact full window, identical-timestamp stall) and assert byte-identical pagination tokens against captured fixtures.
- `make build`, `make test`, `make test-race`, `make lint` all pass.

## Do NOT

- Do not change the tool schema, `_meta` keys, or default page size.
- Do not rename the tool or any exported identifier.
- Do not refactor sibling tools or the upstream `internal/intervals/activities.go` contract in this task — file a follow-up if the driver looks reusable.
- Do not panic anywhere; preserve error wrapping with `%w`.
- Do not introduce a generic pagination abstraction in `internal/` for "future reuse" — keep it scoped to this tool. One concrete use first.

## Documentation

- `STATUS.md`
- `CHANGELOG.md` `[Unreleased]` (under "Changed" — internal refactor, no user-visible behaviour change).

## Git Commit Convention

`TP-044 extract iteratePages driver from fetchActivitiesPage`, etc.

---

## Amendments

_Add amendments below this line only._
