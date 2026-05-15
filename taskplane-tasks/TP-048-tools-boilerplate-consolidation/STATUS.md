# TP-048-tools-boilerplate-consolidation — Status

**Current Step:** Step 1: Helpers + tests
**Status:** ⏳ Not started
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 0
**Iteration:** 0
**Size:** M

---

### Step 1: Helpers + tests

**Status:** ⏳ Not started

- [ ] Add `internal/tools/decode.go` with `DecodeStrict[T any]`
- [ ] Add `internal/tools/result.go` with `TextResult`
- [ ] Table-driven tests for both helpers (zero input, unknown field, malformed, happy path)

### Step 2: Mechanical replacement across tool files

**Status:** ⏳ Not started

- [ ] Replace decode boilerplate in every `internal/tools/<tool>.go` with `DecodeStrict`
- [ ] Replace exact-match `Result{…}` boilerplate with `TextResult`
- [ ] Commit per logical batch (reads / writes / wellness / etc.)

### Step 3: `get_activities.go` cleanups

**Status:** ⏳ Not started

- [ ] Confirm `stringSet` has no callers, delete it
- [ ] Promote inline anonymous struct in `validateActivitiesTokenArgs` to a named type

### Step 4: `Requirement` enum

**Status:** ⏳ Not started

- [ ] Choose `int`+`iota` vs typed `string` (record decision below)
- [ ] Convert constants in `internal/tools/registry.go:286-293`
- [ ] Update all call sites
- [ ] Preserve wire format if serialised

### Step 5: Verify

**Status:** ⏳ Not started

- [ ] `make build` / `test` / `test-race` / `lint`
- [ ] Schema-stability snapshot byte-identical
- [ ] Acceptance greps collapse as expected
- [ ] Manual `list_tools` parity check

---

## Decisions

_Record `Requirement` enum shape (`int`+`iota` vs typed `string`) in Step 4._

## Notes

_Add notes as work progresses._
