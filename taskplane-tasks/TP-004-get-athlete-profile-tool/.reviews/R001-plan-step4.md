# Plan Review: TP-004 Step 4 — add tests

## Verdict

**Approve.** The Step 4 plan now covers the required regression net for `get_athlete_profile`: registration metadata, fake-client success, exact default vs `include_full` response deltas, strict argument validation, upstream error sanitization, cancellation behavior, normalized athlete IDs, `_meta.server_version`, timezone fallback, unit normalization, pace key selection, and absence of secret/debug fields.

## Notes for implementation

- Keep the tests table-driven with `t.Run` and avoid network access entirely; the fake `ProfileClient` with call count and context capture is the right approach.
- In the success/validation tests, explicitly assert the fake client call count and captured context where relevant: valid calls should call the fake once with the handler context; invalid arguments should not call it.
- For error tests, check the public message via `err.Error()` or `PublicErrorMessage(err)` for `UserError`s, but do not require the wrapped internal cause to be sanitized; retaining the cause is intentional for server-side logging.
- For the forbidden-output assertions, prefer sentinel values in fixture data plus marshaled-response checks, so the tests prove no raw/debug/credential-like fields leak without being overly broad or brittle.
- Run at least the targeted package tests after adding the file, e.g. `go test ./internal/tools`, before handing off to Step 5's broader `make test`/lint/build verification.
