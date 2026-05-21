# Review R005 — Step 2 plan

Decision: **Approve.**

The updated Step 2 plan in `STATUS.md` addresses the blockers from R003: it uses the sanitized captured NOTE fixtures, adds an intervals-client serialization regression test that will exercise the real outbound bulk JSON, and adds a tool-level validation regression for the discovered NOTE `name` requirement.

## Notes for implementation

- Put the serialization regression in `internal/intervals/events_test.go` and compare the captured `POST /athlete/i12345/events/bulk` body to `internal/intervals/testdata/events/note_create_request.json`. The important pre-fix failure is `start_date_local`: current code emits `2026-05-25` for NOTE, while the fixture expects `2026-05-25T00:00:00`.
- Use `internal/intervals/testdata/events/note_create_response.json` as the bulk response fixture if practical, so the test also proves the captured response decodes into an `Event`.
- For the tool validation regression, assert that a NOTE create without a non-empty `name` returns a user-facing validation error and does not call the fake writer client.
- When fixing in Step 3, update the input schema/error text as well as validation so the NOTE `name` requirement remains schema-honest.

Record the targeted failing command output before applying the fix, e.g. `go test ./internal/intervals -run TestAddOrUpdateEventSendsNoteCreateBody` plus the tool validation test command.
