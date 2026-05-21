# Plan Review — Step 1: Helpers + tests

## Verdict

Approved for implementation.

The revised Step 1 plan addresses the R001 blockers: it pins the `DecodeStrict` edge cases that matter for later mechanical replacement (empty/whitespace zero value, non-object rejection, trailing JSON rejection) and documents `TextResult` as a no-error helper whose Step 2 usage must be limited to behavior-preserving call sites.

## Notes to carry into implementation

- Add doc notes for both exported helpers (`DecodeStrict` and `TextResult`) so lint remains clean.
- In the `DecodeStrict` tests, assert the exact stable messages/prefixes required by the prompt for malformed/unknown-field decode errors, plus the literal `arguments must be a JSON object` and `unexpected trailing JSON` cases noted in `STATUS.md`.
- For `TextResult`, keep the implementation narrowly equivalent to the current ignored-error pattern: `json.Marshal(shaped)` for `Content[0].Text`, `ContentTypeText`, and `StructuredContent: shaped`. Do not use it in Step 2 where an existing handler returns a checked `encoding <tool> response: %w` error unless that behavior is intentionally preserved another way.

No additional plan changes are required before coding Step 1.
