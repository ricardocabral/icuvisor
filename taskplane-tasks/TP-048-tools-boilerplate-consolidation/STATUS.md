# TP-048-tools-boilerplate-consolidation — Status

**Current Step:** Step 2: Mechanical replacement across tool files
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 6
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

**Status:** 🟨 In Progress

- [x] Audit package-local `decodeStrict(raw, &args)` callers and preserve their existing empty/whitespace `arguments must be a JSON object` behavior with minimal prechecks where required
- [x] Replace package-local `decodeStrict(raw, &args)` callers with `DecodeStrict[T](raw)` and remove the old helper when unused
- [x] Replace decode boilerplate in every `internal/tools/<tool>.go` with `DecodeStrict`, preserving bespoke empty-input/raw-field validation ordering
- [x] Replace exact-match `Result{…}` boilerplate with `TextResult`, limiting checked `json.Marshal` sites to JSON-marshalable-by-construction payloads
- [x] Run targeted `go test ./internal/tools` and acceptance greps for `DisallowUnknownFields`, `decodeStrict(`, and `ContentTypeText`
- [x] Commit per logical batch (reads / writes / wellness / etc.)

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
- Step 2 plan review R004: remove the old unexported `decodeStrict` instead of wrapping it; preserve `decodeGetActivitiesRequest`, `decodeActivityReadRequest`, and raw-field precheck error ordering; use `TextResult` for exact constructions where payloads are JSON-marshalable by construction.
- Step 2 plan review R005: old `decodeStrict(raw, &args)` callers currently reject empty/whitespace as `arguments must be a JSON object`; audit those callers and add minimal prechecks before `DecodeStrict[T]` unless a wrapper already explicitly allowed empty input.


| 2026-05-15 13:33 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 13:33 | Step 1 started | Helpers + tests |
| 2026-05-15 13:37 | Review R001 | plan Step 1: UNKNOWN |
| 2026-05-15 13:40 | Review R002 | plan Step 1: APPROVE |
| 2026-05-15 13:46 | Review R003 | code Step 1: APPROVE |
| 2026-05-15 13:50 | Review R004 | plan Step 2: REVISE |
| 2026-05-15 13:53 | Review R005 | plan Step 2: REVISE |
| 2026-05-15 13:55 | Review R006 | plan Step 2: APPROVE |
