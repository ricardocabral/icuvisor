# TP-048-tools-boilerplate-consolidation — Status

**Current Step:** Step 5: Verify
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 2
**Review Counter:** 13
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

**Status:** ✅ Complete

- [x] Audit package-local `decodeStrict(raw, &args)` callers and preserve their existing empty/whitespace `arguments must be a JSON object` behavior with minimal prechecks where required
- [x] Replace package-local `decodeStrict(raw, &args)` callers with `DecodeStrict[T](raw)` and remove the old helper when unused
- [x] Replace decode boilerplate in every `internal/tools/<tool>.go` with `DecodeStrict`, preserving bespoke empty-input/raw-field validation ordering
- [x] Replace exact-match `Result{…}` boilerplate with `TextResult`, limiting checked `json.Marshal` sites to JSON-marshalable-by-construction payloads
- [x] Run targeted `go test ./internal/tools` and acceptance greps for `DisallowUnknownFields`, `decodeStrict(`, and `ContentTypeText`
- [x] Commit per logical batch (reads / writes / wellness / etc.)

### Step 3: `get_activities.go` cleanups

**Status:** ✅ Complete

- [x] Confirm current `stringSet` callers and replace them behavior-preservingly before deleting the `internal/tools` helper
- [x] Clarify acceptance grep evidence for `stringSet`: `internal/tools` should have no helper/callers; unrelated `internal/toolchecks/schema_stability.go` is out of scope
- [x] Promote inline anonymous struct in `validateActivitiesTokenArgs` to a named type

### Step 4: `Requirement` enum

**Status:** ✅ Complete

- [x] Choose `int`+`iota` vs typed `string` (record decision below)
- [x] Convert constants in `internal/tools/registry.go:286-293`
- [x] Update all call sites, including `internal/mcp` and `internal/safety` references from `grep -rn "Requirement" internal/`
- [x] Preserve zero-value/default read behavior without raw string comparisons
- [x] Preserve wire format if serialised

### Step 5: Verify

**Status:** 🟨 In Progress

- [ ] Update `CHANGELOG.md` under `[Unreleased]` / `Changed`
- [ ] `make build` / `test` / `test-race` / `lint`
- [ ] Schema-stability snapshot byte-identical
- [ ] Acceptance greps collapse as expected
- [ ] Manual `list_tools` parity check

---

## Decisions

- Step 4 decision: keep `type Requirement string` with exact wire values (`"read"`, `"write"`, `"delete"`) because requirements are serialized by `list_advanced_capabilities`; improve typed/default handling without switching to `int`/`iota`.

## Notes

- Step 1 plan review R001: keep `DecodeStrict` object-only semantics (`arguments must be a JSON object`), reject trailing JSON with `unexpected trailing JSON`, and keep `TextResult(shaped any) Result` no-error per prompt; Step 2 should only replace exact result construction where this preserves behavior.
- Step 2 plan review R004: remove the old unexported `decodeStrict` instead of wrapping it; preserve `decodeGetActivitiesRequest`, `decodeActivityReadRequest`, and raw-field precheck error ordering; use `TextResult` for exact constructions where payloads are JSON-marshalable by construction.
- Step 2 plan review R005: old `decodeStrict(raw, &args)` callers currently reject empty/whitespace as `arguments must be a JSON object`; audit those callers and add minimal prechecks before `DecodeStrict[T]` unless a wrapper already explicitly allowed empty input.
- Step 3 plan review R008: `stringSet` still has callers in `get_activities.go` and `get_activity_streams.go`; remove the helper only after replacing those callers, and scope acceptance to `internal/tools` because `internal/toolchecks/schema_stability.go` has an unrelated helper.
- Step 4 plan review R011: typed string is already the safe enum shape; preserve serialized values and default empty-as-read behavior while auditing `internal/` references.


| 2026-05-15 13:33 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 13:33 | Step 1 started | Helpers + tests |
| 2026-05-15 13:37 | Review R001 | plan Step 1: UNKNOWN |
| 2026-05-15 13:40 | Review R002 | plan Step 1: APPROVE |
| 2026-05-15 13:46 | Review R003 | code Step 1: APPROVE |
| 2026-05-15 13:50 | Review R004 | plan Step 2: REVISE |
| 2026-05-15 13:53 | Review R005 | plan Step 2: REVISE |
| 2026-05-15 13:55 | Review R006 | plan Step 2: APPROVE |
| 2026-05-15 14:03 | Review R007 | code Step 2: APPROVE |
| 2026-05-15 14:05 | Review R008 | plan Step 3: REVISE |
| 2026-05-15 14:07 | Review R009 | plan Step 3: APPROVE |
| 2026-05-15 14:11 | Review R010 | code Step 3: APPROVE |
| 2026-05-15 14:14 | Review R011 | plan Step 4: REVISE |
| 2026-05-15 14:16 | Review R012 | plan Step 4: APPROVE |
| 2026-05-15 14:20 | Review R013 | code Step 4: APPROVE |
