# TP-002 — Status

**Issue:** v0.1 — intervals client
**Iteration:** 1
**Current Step:** Step 3: Implement retries and structured errors
**Last Updated:** 2026-05-10
**State:** In Progress

## Step 1: Plan from public API docs and current config

**Status:** ✅ Complete

- [x] Identify public intervals.icu athlete profile endpoint(s)
- [x] Record source/uncertainty in STATUS.md
- [x] Define minimal typed request/response structs
- [x] Decide retry policy/backoff/cancellation behavior
- [x] R001: Align planned profile struct names with `/athlete/{id}` and `/profile` schemas
- [x] R001: Make task/step state internally consistent before Step 2
- [x] R001: Fix Discoveries table formatting

## Step 2: Implement the HTTP client core

**Status:** ✅ Complete

- [x] Create typed client from config/version/shared http client
- [x] Use Basic Auth without leaking API key
- [x] Set `User-Agent: icuvisor/<version>`
- [x] Use context-aware requests
- [x] Always close response bodies
- [x] Decode JSON into typed structs

## Step 3: Implement retries and structured errors

**Status:** 🟡 In Progress

- [x] Retry 429 and 5xx with exponential backoff and jitter
- [x] Respect `Retry-After` when practical
- [x] Return structured/sentinel errors for stable contract points
- [x] Wrap errors with `%w`
- [x] Keep error strings secret-free and short
- [x] R001: Add focused coverage or caller so retry/error helpers are lint-clean

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
| 2026-05-10 21:59 | Task started | Runtime V2 lane-runner execution |
| 2026-05-10 21:59 | Step 1 started | Plan from public API docs and current config |
| 2026-05-10 | Public API docs identify profile endpoints | Public OpenAPI spec at `https://intervals.icu/api/v1/docs` (linked from `https://intervals.icu/api-docs.html`) exposes `GET /api/v1/athlete/{id}/profile` -> `AthleteProfile`, `GET /api/v1/athlete/{id}` -> `WithSportSettings`, and `GET /api/v1/athlete/{athleteId}/sport-settings` / `{id}` -> `SportSettings`; forum API guide confirms Basic Auth username `API_KEY` and password API key, and says athlete id `0` selects the key owner. |
| 2026-05-10 | Step 1 endpoint uncertainty recorded | `AthleteProfile` includes public-ish identity fields but not FTP/zones/units; `WithSportSettings` includes `measurement_preference`, `weight_pref_lb`, `fahrenheit`, `timezone`, and embedded `sportSettings`; `SportSettings` includes `ftp`, `indoor_ftp`, `w_prime`, `p_max`, power/HR/pace zones and threshold fields. v0.1 should prefer `GET /athlete/{id}` for profile-with-sport-settings, with `/profile` as lighter identity-only reference if needed. |
| 2026-05-10 | Minimal v0.1 response shapes defined | Use typed `AthleteWithSportSettings` (matching `GET /api/v1/athlete/{id}` / OpenAPI `WithSportSettings`) as the main v0.1 response: athlete `id`, `name`, `firstname`, `lastname`, `measurement_preference`, `weight_pref_lb`, `fahrenheit`, `timezone`, `locale`, and embedded `sportSettings`. Define `SportSettings` with stable fields only: `id`, `athlete_id`, `types`, `ftp`, `indoor_ftp`, `w_prime`, `p_max`, `power_zones`, `power_zone_names`, `lthr`, `max_hr`, `hr_zones`, `hr_zone_names`, `threshold_pace`, `pace_units`, `pace_zones`, `pace_zone_names`. Do not define or call the lighter `/profile` `AthleteProfile` wrapper in v0.1 unless later black-box validation shows `/athlete/{id}` is unavailable. No generic raw payload in default return. |
| 2026-05-10 | Retry policy selected | Implement stdlib-only retry for idempotent GETs: max 3 attempts, retry HTTP 429 and 5xx plus transient transport errors, exponential backoff starting near 200ms with jitter and cap near 2s, respect `Retry-After` seconds/date when present within cap, and abort sleeps/requests immediately on context cancellation. Do not retry 401/403/404 or malformed JSON. |
| 2026-05-10 22:05 | Review R001 | code Step 1: UNKNOWN |
| 2026-05-10 22:22 | Review R001 | code Step 1: UNKNOWN |
| 2026-05-10 | Step 1 review recovery | Reviewer produced UNKNOWN after R001 fixes rather than an APPROVE/REVISE verdict; findings were addressed in STATUS.md and execution proceeds cautiously as reviewer-unavailable/unclear. |
| 2026-05-10 | Step 2 started | Implement the HTTP client core. |
| 2026-05-10 | Step 3 started | Implement retries and structured errors. |
| 2026-05-10 22:23 | Review R001 | code Step 1: UNKNOWN |
| 2026-05-10 22:25 | Review R001 | plan Step 2: APPROVE |
| 2026-05-10 22:30 | Review R001 | code Step 2: APPROVE |
| 2026-05-10 22:55 | Review R001 | code Step 3: REVISE |
