# Plan Review — Step 1: Helpers + tests

## Verdict

Request minor plan revisions before implementation. The helper direction is correct, but Step 1 needs to pin a few edge-case semantics so the mechanical replacements in Step 2 do not silently change handler behavior.

## Findings

### 1. `DecodeStrict` tests should cover trailing JSON

Current strict decoders do a second `Decode(&struct{}{})` and return `unexpected trailing JSON` unless it gets `io.EOF` (for example `get_activities.go` and `get_fitness.go`). The Step 1 test checklist covers empty, unknown fields, malformed JSON, and happy path, but not valid JSON followed by extra tokens.

Please add a table case such as `{"include_full":false} {}` and assert the same trailing-JSON rejection. This is part of the boilerplate being consolidated, not an optional behavior.

### 2. Define empty and non-object argument behavior explicitly

The prompt says empty input should decode to the zero value, which matches `decodeGetAthleteProfileRequest`, but not all current local decoders (`get_fitness.go` currently rejects empty/non-object input before decoding, and `get_activities.go` has custom required-argument handling). For the generic helper, the plan should state the exact behavior:

- whitespace-only / empty raw input returns zero value and nil error;
- non-object JSON should still be rejected with the existing short message (`arguments must be a JSON object`) if the helper is going to replace object-only decoders;
- decode errors should use one stable wrapping prefix, and tests should assert it.

Without this, Step 2 can accidentally change internal errors and possibly test expectations.

### 3. `TextResult(shaped any) Result` has unresolved marshal-error semantics

Most existing handlers call `json.Marshal(shaped)` and return a wrapped `encoding <tool> response: %w` error on failure before constructing the `Result`; a few currently ignore the marshal error. A helper with signature `TextResult(shaped any) Result` cannot preserve the checked-error behavior.

Before implementing, document the intended contract for marshal failures. Acceptable paths are either:

- constrain `TextResult` usage in Step 2 to call sites where marshal failure is already intentionally ignored / impossible by construction, and leave checked-error handlers alone; or
- adjust the helper contract to return an error (if allowed by the task owner), so existing error behavior can be preserved.

As written, the plan says “match exactly” but the proposed signature makes that ambiguous.

## Non-blocking notes

- Because `DecodeStrict` and `TextResult` are exported identifiers in package `tools`, include doc comments starting with the identifier names to satisfy lint and repository conventions.
- Add a test case for non-object JSON (`[]`, `true`, or a quoted string) if the helper keeps the existing “arguments must be a JSON object” guard.
- `TextResult` tests should compare against a hand-built `Result` using the same `json.Marshal` output, not a pretty-printed or map-order-sensitive string built by hand.
