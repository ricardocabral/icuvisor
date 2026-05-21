# Code Review: Step 2 failing test

Result: changes requested.

The new red tests correctly encode the live-probe finding that `feel` must be rejected instead of silently omitted, and the accepted seven-field fixture test is a good addition. I ran `go test ./internal/intervals ./internal/tools`; it fails on the intended unsupported-`feel`/schema assertions.

## Findings

### 1. Schema test does not protect examples from still advertising `feel`

`internal/tools/update_wellness_test.go:35-40` now checks that `feel` is absent from the schema `properties`, but the public schema also exposes `examples` / `input_examples`. The current implementation still has an example containing `"feel": 4` (`internal/tools/update_wellness.go:414-419`). A Step 3 fix could remove the property and satisfy this test while still showing LLM clients an example that sends the upstream-rejected field.

Please extend the schema/public-contract test to walk both `examples` and `input_examples` and fail if any example contains `feel` (or otherwise advertise it as writable). That matches the Step 2 goal of ensuring `feel` is not advertised after the fix.

### 2. Unsupported-`feel` tool test does not assert the “no partial-success metadata” requirement

`internal/tools/update_wellness_test.go:72` discards the `Result`. The plan/previous review explicitly required that a dogfood bundle containing `feel` not return partial-success metadata or a `fields_updated` claim. As written, an implementation could return a non-empty `Result` with `_meta.fields_updated` while also returning the public error, and this test would still pass.

Please capture the result and assert it is empty on this validation error, e.g. no content and nil structured content (or the project’s preferred zero-result assertion), in addition to the existing zero writer-call assertion.

## Notes

- The intervals fixture test is separated from the red `feel` tests as requested and uses the sanitized request/response fixtures.
- Existing positive tests that still send `feel` will need to be reconciled in Step 3; that is okay for Step 2 as long as the red tests above fully pin the intended contract.
