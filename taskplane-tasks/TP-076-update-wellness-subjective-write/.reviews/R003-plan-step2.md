# Plan Review: Step 2 failing test

Result: changes requested.

The Step 2 plan is pointed at the right area, but it needs to be tightened before implementation. The live probe changed the expected defect shape: this is no longer a simple field-name mismatch such as `sleepQuality`; upstream accepted the seven-field subjective/lock payload and rejected `feel` itself. The tests must encode that contract without accidentally blessing silent data loss.

## Findings

### 1. Do not make the fixture test silently drop `feel`

The captured request fixture documents the accepted upstream body as the seven fields excluding `feel`, with `omitted_fields.feel` explaining the 422. If the new intervals-client test takes a full dogfood bundle including `Feel` and simply expects the outbound body to omit `feel`, it will encode unsafe behavior: the tool would appear to succeed while not writing a user-supplied field, and `_meta.fields_updated` could still claim `feel` was updated.

Required plan amendment:

- Decide the final public behavior for `feel` before writing the failing assertion. Given the probe result, the safe behavior should be explicit rejection, e.g. `field_not_writable: feel (not accepted by intervals.icu wellness write)`, with no upstream write attempted.
- Add a tool-level test for the dogfood bundle including `feel` that asserts:
  - the public error message is explicit/actionable, not the generic `could not update wellness...` message;
  - the writer client receives zero calls;
  - the response does not claim partial success or include `feel` in `fields_updated`.
- If implementation also changes the intervals client, prefer a test that `WriteWellnessParams{Feel: ...}` returns an unsupported-field error before network I/O, rather than a test that silently omits `feel`.

### 2. Separate the accepted-upstream fixture test from the failing defect test

A fixture-driven test using `internal/intervals/testdata/wellness/subjective_write_request.json` should document the live-accepted seven-field payload (`fatigue`, `soreness`, `stress`, `mood`, `motivation`, `sleepQuality`, `locked`). That test may already pass on current code if `Feel` is not set, so it should not be the only “red” test.

Required plan amendment:

- Add/keep a fixture test for the accepted seven-field shape, comparing the fixture’s `method`, redacted `path` shape, and nested `body` against the httptest request, and using `subjective_write_response.json` for the decoded response.
- Add a separate failing test for the unsupported `feel` behavior described above. This is the test that should fail on current `main`.
- Do not assert that a full bundle including `feel` produces the seven-field fixture body unless the product decision is intentionally “drop `feel`,” which would need stronger justification and user-visible metadata.

### 3. Cover schema/public-contract fallout

`update_wellness` currently exposes `feel` in the input schema and validates it as a 1-5 writable field. If the fix is to reject unsupported `feel`, the Step 2 test plan should include the schema/public contract expected after the fix so the LLM is not encouraged to keep sending a field that upstream rejects.

Required plan amendment:

- Add a schema assertion for the chosen contract: either `feel` is no longer exposed as writable, or its submitted use is rejected with the same explicit public error used by read-only fields.
- Update the existing range/schema test expectations accordingly in the same step, so the eventual fix cannot leave schema and runtime behavior inconsistent.

## Notes

- The Step 1 cleanup blocker (probe row remains locked) does not block adding unit tests, but it remains a live-account safety blocker for final acceptance and any further live re-validation.
- Keep the sanitized fixtures placeholder-based; do not introduce raw athlete IDs or date-identifying probe data into tests.
