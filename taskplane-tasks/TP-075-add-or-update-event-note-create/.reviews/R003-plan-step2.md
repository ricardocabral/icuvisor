# Review R003 — Step 2 plan

Decision: **Request changes.**

The planned direction is close, but as written it can produce a non-failing or insufficient test and it would consume a fixture that still has unresolved Step 1 review issues.

## Required changes before implementing Step 2

1. **Resolve the outstanding Step 1 review before using the fixtures.**  
   R002 requested sanitization/documentation changes, but the current `STATUS.md` marks Step 1 complete and `internal/intervals/testdata/events/note_create_response.json` still contains `"calendar_id": 1`. Since Step 2 says it will use the captured fixture, first redact that account-specific calendar ID and update `STATUS.md` with the missing probe-matrix outcomes (or explicitly mark them incomplete). Do not build a new golden test around an unsanitized fixture.

2. **Add an intervals-client serialization test, not only a tool fake-client test.**  
   Mirroring `TestAddOrUpdateEventCreatePreservesFreeTextTagsAndReadShape` in `internal/tools` is useful for NOTE response shaping, but it will not catch the current bug: the fake client records `WriteEventParams.Date` before `internal/intervals.writeEventStartDateLocal` turns it into the outbound JSON. A tool-level NOTE test can pass on main while the real HTTP payload is still rejected upstream.

   Step 2 must include a failing test in `internal/intervals/events_test.go` (either a separate test or a NOTE case added carefully to `TestAddOrUpdateEventSendsCreateAndUpdateBodies`) that:
   - calls `Client.AddOrUpdateEvent` with `WriteEventParams{Date: "2026-05-25", Category: "NOTE", Name: ..., Description: ...}`;
   - captures the `POST /athlete/i12345/events/bulk` request from an `httptest.Server`;
   - compares the captured body to `internal/intervals/testdata/events/note_create_request.json`, especially `start_date_local: "2026-05-25T00:00:00"`;
   - uses the sanitized `note_create_response.json` as the server response if practical.

   This should fail on current main because it sends NOTE `start_date_local` as `YYYY-MM-DD`.

3. **Cover the discovered NOTE `name` requirement or document why it is deferred.**  
   Step 1 found that NOTE creates require a non-empty `name`, while the current tool schema/validation still describes `name` as optional for all categories. Under the schema-honesty requirement, Step 2 should add a failing tool-level validation test for `category: "NOTE"` without `name` (or the plan should explicitly say this is out of scope and create a follow-up). If you add the validation test now, Step 3 should also update the input schema/error text accordingly.

## Suggested Step 2 shape

- First fix the fixture/status issues from R002.
- Add `TestAddOrUpdateEventSendsNoteCreateBody` in `internal/intervals/events_test.go` using the captured request fixture as the expected outbound body.
- Optionally add a separate `internal/tools` NOTE test for read shape/free-text preservation, but do not treat it as the failing regression test for the date-format bug.
- Run the targeted test command and record that it fails before the implementation fix, e.g. `go test ./internal/intervals -run TestAddOrUpdateEventSendsNoteCreateBody`.
