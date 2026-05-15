# TP-048-tools-boilerplate-consolidation — Status

**Current Step:** Step 1: Helpers + tests
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 3
**Iteration:** 1
**Size:** M

---

### Step 1: Helpers + tests

**Status:** ✅ Complete

- [x] Add `internal/tools/decode.go` with `DecodeStrict[T any]`
- [x] Add `internal/tools/result.go` with `TextResult`
- [x] Table-driven tests for both helpers (zero input, unknown field, malformed, happy path)
- [x] Cover `DecodeStrict` trailing JSON rejection plus empty/whitespace zero-value and non-object argument errors
- [x] Document `TextResult` as a no-error helper for shaped values that are JSON-marshalable by construction, comparing test output to a hand-built `Result`

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

- Step 1 plan review R001: keep `DecodeStrict` object-only semantics (`arguments must be a JSON object`), reject trailing JSON with `unexpected trailing JSON`, and keep `TextResult(shaped any) Result` no-error per prompt; Step 2 should only replace exact result construction where this preserves behavior.


| 2026-05-15 13:33 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 13:33 | Step 1 started | Helpers + tests |
| 2026-05-15 13:37 | Review R001 | plan Step 1: UNKNOWN |
| 2026-05-15 13:40 | Review R002 | plan Step 1: APPROVE |
| 2026-05-15 13:46 | Review R003 | code Step 1: APPROVE |
