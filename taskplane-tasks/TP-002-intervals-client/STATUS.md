# TP-002 — Status

**Issue:** v0.1 — intervals client
**State:** Ready

## Step 1: Plan from public API docs and current config

**Status:** ⬜ Not started

- [ ] Identify public intervals.icu athlete profile endpoint(s)
- [ ] Record source/uncertainty in STATUS.md
- [ ] Define minimal typed request/response structs
- [ ] Decide retry policy/backoff/cancellation behavior

## Step 2: Implement the HTTP client core

**Status:** ⬜ Not started

- [ ] Create typed client from config/version/shared http client
- [ ] Use Basic Auth without leaking API key
- [ ] Set `User-Agent: icuvisor/<version>`
- [ ] Use context-aware requests
- [ ] Always close response bodies
- [ ] Decode JSON into typed structs

## Step 3: Implement retries and structured errors

**Status:** ⬜ Not started

- [ ] Retry 429 and 5xx with exponential backoff and jitter
- [ ] Respect `Retry-After` when practical
- [ ] Return structured/sentinel errors for stable contract points
- [ ] Wrap errors with `%w`
- [ ] Keep error strings secret-free and short

## Step 4: Implement athlete profile retrieval

**Status:** ⬜ Not started

- [ ] Add profile retrieval method using configured athlete ID
- [ ] Capture v0.1 profile fields
- [ ] Expose timezone/unit fields to later response shaping when available
- [ ] Avoid heavy/raw payloads by default

## Step 5: Test and verify

**Status:** ⬜ Not started

- [ ] Use `httptest.Server`; no network tests
- [ ] Test auth, User-Agent, URL construction, cancellation, body closing where practical
- [ ] Test retry/no-retry behavior
- [ ] Test structured error classification and secret redaction
- [ ] Test profile decoding from fixtures
- [ ] Run `go fmt ./...`, `make test`, `make build`, and `make lint` if available
- [ ] If local `.env` has `INTERVALS_ICU_ATHLETE_ID` and `INTERVALS_ICU_API_KEY`, run optional end-to-end profile fetch without printing secrets
- [ ] Update `CHANGELOG.md`

## Discoveries

| Date | Finding | Impact |
| ---- | ------- | ------ |
