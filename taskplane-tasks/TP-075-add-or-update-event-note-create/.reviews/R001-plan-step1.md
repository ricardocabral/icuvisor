# Review R001 — Step 1 plan

Decision: **Approved with required amendments before probing.**

The Step 1 plan is directionally correct: this defect depends on the live intervals.icu write contract, and probing one field at a time against `.env-dev` is the right first move. I would not proceed exactly as written without tightening the probe shape and safety controls below.

## Required amendments

1. **Probe the same endpoint/body shape used by the shipped client.**
   `internal/intervals/events.go` creates events with `POST /athlete/{id}/events/bulk` and a one-element JSON array. The plan describes a bare object posted to “the events endpoint”. If a single-event endpoint behaves differently, the probe may identify the wrong contract. Use the production path as the primary probe:
   - method: `POST`
   - path: `/athlete/${INTERVALS_ICU_ATHLETE_ID}/events/bulk`
   - body: `[ { ...candidate event fields... } ]`
   Any additional single-object endpoint experiment should be clearly marked as non-authoritative for this tool.

2. **Start by reproducing the current failing payload.**
   Before trying the proposed datetime baseline, send the payload shape the current code emits for NOTE: `start_date_local: "YYYY-MM-DD"`, `category: "NOTE"`, no `type` unless supplied, and the same minimal `name`/`description` shape used in dogfood. Then mutate one field at a time. This makes the root-cause conclusion defensible.

3. **Add a live-write safety preflight.**
   Because Step 1 writes to a real upstream account, add a preflight that verifies `.env-dev` is loaded and is not the maintainer’s primary `.env` athlete, without printing raw athlete IDs or API keys. Do not run with shell xtrace, do not use verbose curl traces that echo auth headers, and do not paste command output containing credentials into status or fixtures.

4. **Track created probe IDs immediately and clean by exact ID.**
   Maintain a temporary local ledger of every successful probe’s event ID, date, and unique probe name. If a request times out or returns an ambiguous response, re-list the probe date range by unique `tp-075` names before cleanup. `delete_event` is registered only with `ICUVISOR_DELETE_MODE=full`; if using the MCP tool for cleanup, explicitly run in full mode for cleanup only, or use the direct intervals `DELETE /athlete/{id}/events/{eventId}` endpoint. Verify the final date-range read contains no `tp-075` probe events.

5. **Include the remaining suspected contract axes.**
   The prompt lists category casing and NOTE `name`/`description` semantics as suspects. The matrix should include, after the primary date/type probes:
   - category casing (`NOTE` vs the casing observed from a UI-created note, if available);
   - name-only, description-only, and name+description variants, so we know whether NOTE requires `name`, `description`, or both.

## Fixture guidance for Step 1 output

- Save only the accepted authoritative request/response pair, sanitized, under `internal/intervals/testdata/events/note_create_request.json` and `note_create_response.json`.
- Redact all event IDs, athlete IDs, calendar IDs, and any other account-identifying nested IDs. Replace live event IDs consistently with `EVENT_ID_PLACEHOLDER`.
- Before committing fixtures, run a local grep/check for the raw athlete ID prefix and each created event ID.

## Follow-up test implication

One issue to carry into Step 2: a tool-level fake-client test mirroring `TestAddOrUpdateEventCreatePreservesFreeTextTagsAndReadShape` will not catch the likely date-format bug, because the fake client only sees `WriteEventParams.Date` before `internal/intervals.writeEventStartDateLocal` transforms it. If Step 1 confirms a payload/date-format issue, Step 2 also needs an `internal/intervals` client/body serialization test (likely extending `TestAddOrUpdateEventSendsCreateAndUpdateBodies`) that asserts the captured NOTE outbound body. Otherwise the “failing test” may not fail on `main`.
