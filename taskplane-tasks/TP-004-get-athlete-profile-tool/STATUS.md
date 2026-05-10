# TP-004 — Status

**Issue:** v0.1 — get_athlete_profile tool
**State:** Ready

## Step 1: Define the tool contract in STATUS.md

**Status:** ⬜ Not started

- [ ] Write intended description, arguments, and response shape
- [ ] Do not accept API key as a tool parameter
- [ ] Decide whether v0.1 needs `include_full`
- [ ] Include units/timezone/athlete-ID conventions where available

## Step 2: Implement the typed tool

**Status:** ⬜ Not started

- [ ] Add typed request/response structs
- [ ] Register exactly `get_athlete_profile`
- [ ] Use a distinguishing first sentence
- [ ] Include useful JSON Schema descriptions
- [ ] Call intervals client with request context
- [ ] Return short actionable LLM-facing errors

## Step 3: Shape the response for v0.1

**Status:** ⬜ Not started

- [ ] Return terse useful profile fields
- [ ] Use disambiguating field names/metadata where applicable
- [ ] Include `_meta.server_version`
- [ ] Exclude fetched timestamps/debug cruft by default
- [ ] Exclude secrets/raw upstream payloads by default

## Step 4: Add tests

**Status:** ⬜ Not started

- [ ] Test registration metadata and no secret args
- [ ] Test successful handler with fake intervals client
- [ ] Test include-full/default behavior if implemented
- [ ] Test upstream error mapping
- [ ] Test `_meta.server_version` and normalized athlete ID

## Step 5: End-to-end local verification

**Status:** ⬜ Not started

- [ ] Exercise MCP stdio tool dispatch to `get_athlete_profile` with fake client/server
- [ ] If local `.env` has `INTERVALS_ICU_ATHLETE_ID` and `INTERVALS_ICU_API_KEY`, run optional end-to-end MCP validation and record only non-sensitive result shape
- [ ] Run `go fmt ./...`
- [ ] Run `make test`
- [ ] Run `make build`
- [ ] Run `make lint` if available
- [ ] Update `CHANGELOG.md`

## Discoveries

| Date | Finding | Impact |
| ---- | ------- | ------ |
