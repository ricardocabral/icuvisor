# Plan Review — TP-077 Step 2

Verdict: **approve**

## Findings

No blocking findings. The revised Step 2 plan addresses the previous review items:

- It targets the actual Step 1 discovery: `create_workout` requires a non-empty `folder_id` for create calls, while the already-correct `type` field should remain covered by the accepted fixture payload.
- It adds failing coverage at the tool boundary for public validation/schema behavior, which is where the user-visible contract belongs.
- It calls out updating existing happy-path tests with a sanitized folder ID so the intended red tests are focused on the new required-folder assertions.
- It explicitly defers production validation/serialization changes to Step 3.

## Non-blocking guidance

- In the intervals client missing-folder test, avoid any real network dependency while the test is red. Prefer an `httptest.Server`/custom transport that fails the test if a request is made, so the current behavior fails deterministically and the fixed behavior returns before I/O.
- Unit tests can only enforce a non-empty `folder_id`; they cannot prove the folder is owned by the athlete. Put the "existing folder owned by the athlete" requirement in schema/validation error text, and leave ownership validation to upstream/live validation.
- When tying the accepted outbound body to `create_request.json`, compare against the sanitized fixture in a way that keeps the fixture authoritative and does not reintroduce account-derived IDs.
