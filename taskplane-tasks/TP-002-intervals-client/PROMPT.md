# TP-002 — intervals.icu client: Basic Auth, retries, structured errors, athlete profile

## Mission

Implement the clean-room intervals.icu HTTP client needed for v0.1: authenticated requests, retry/backoff for transient upstream failures, structured errors, and the athlete profile read path used by `get_athlete_profile`.

Roadmap item: **intervals.icu API client (Basic Auth, retries, structured errors)**.

Complexity: Blast radius 2, Pattern novelty 2, Security 2, Reversibility 1 = 7 → Review Level 2. Size: M.

## Dependencies

- **TP-001** — config/version foundation

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` — especially §7.2.C Athlete & fitness, §7.2.D response shaping, §7.3 Technology, §7.4 assumptions
- `ROADMAP.md` — v0.1 only
- `SECURITY.md`
- `CONTRIBUTING.md`
- `internal/config/` from TP-001
- Public intervals.icu API documentation for athlete/profile endpoints (clean-room only)
- Local `.env` file, if present, for `INTERVALS_ICU_ATHLETE_ID` and `INTERVALS_ICU_API_KEY` values used only for optional end-to-end validation against intervals.icu; never print or commit secrets
- `hhopke/intervals-icu-mcp` (MIT, Python) may be consulted as a non-dependency reference for API behavior/semantics when public docs are incomplete; record consulted behavior in `STATUS.md`

## File Scope

Expected files:

- `internal/intervals/` client implementation
- `internal/intervals/*_test.go`
- `internal/intervals/testdata/` fixtures if useful
- `internal/config/` only for small integration adjustments
- `go.mod` / `go.sum` only if a permissively licensed retry dependency is truly needed
- `CHANGELOG.md`
- `taskplane-tasks/TP-002-intervals-client/STATUS.md`

Do not wire MCP transport or tool registration here except through interfaces/types needed by later tasks.

## Steps

### Step 1: Plan from public API docs and current config

- [ ] Identify the intervals.icu endpoint(s) needed for athlete profile / FTP / sport settings / thresholds
- [ ] Record the public-doc source and any uncertainty in `STATUS.md`
- [ ] Define the minimal typed request/response structs for v0.1 without overfitting to future tools
- [ ] Decide retry policy: which status codes retry, max attempts, backoff, jitter, context cancellation behavior

### Step 2: Implement the HTTP client core

- [ ] Create a typed client that accepts base URL, API key, athlete ID, version/user-agent, and shared `*http.Client`
- [ ] Use Basic Auth as documented by intervals.icu; never put the API key in URLs or logs
- [ ] Set `User-Agent: icuvisor/<version>` on every request
- [ ] Use context-aware requests for every I/O call
- [ ] Always close response bodies
- [ ] Decode JSON into typed structs, not `map[string]any` for stable fields

### Step 3: Implement retries and structured errors

- [ ] Retry 429 and 5xx responses with exponential backoff and jitter
- [ ] Respect `Retry-After` when present if practical
- [ ] Do not retry non-idempotent methods; v0.1 only needs GET
- [ ] Return structured/sentinel errors for unauthorized, not found, rate limited, and upstream failures
- [ ] Wrap errors with `%w`; call sites must be able to use `errors.Is` / `errors.As`
- [ ] Ensure error strings are short and never include API keys or raw response bodies that may contain sensitive data

### Step 4: Implement athlete profile retrieval

- [ ] Add `GetAthleteProfile(ctx)` or equivalent against the normalized configured athlete ID
- [ ] Capture fields needed by v0.1 response: athlete ID, name/display name if available, FTP/thresholds/zones/sport settings when exposed
- [ ] Preserve unknown future fields only if needed for tests/debug; do not return heavy payloads by default
- [ ] Make timezone/unit fields available to later response shaping if exposed

### Step 5: Test and verify

- [ ] Use `httptest.Server` for all tests; never hit the network from tests
- [ ] Test Basic Auth header, User-Agent, URL construction, context cancellation, body closure where practical
- [ ] Test retry behavior for 429/5xx and no retry for 4xx
- [ ] Test structured error classification and secret redaction
- [ ] Test athlete profile decoding from fixtures
- [ ] Run `go fmt ./...`, `make test`, `make build`, and `make lint` if available
- [ ] If `.env` contains `INTERVALS_ICU_ATHLETE_ID` and `INTERVALS_ICU_API_KEY`, run a clearly separated manual/end-to-end validation that fetches the configured athlete profile from intervals.icu; do not make this part of unit tests and do not print secrets
- [ ] Update `CHANGELOG.md` and `STATUS.md`

## Reference Implementation Policy

- You may use `hhopke/intervals-icu-mcp` as a permissively licensed, Python-only reference implementation for intervals.icu API reverse-engineering and MCP tool semantics. It must not be added as a dependency. When reading it, extract behavior and API contracts, then implement idiomatic Go from first principles. Record the consulted files/behaviors in `STATUS.md`.
- You may not use GPL/copyleft source code as an implementation reference because this project is MIT clean-room. Do not read, copy, paraphrase, transliterate, or port its code. Use the PRD's summarized insights, intervals.icu public docs, black-box testing, and permissively licensed references instead.

## Acceptance Criteria

- A clean-room `internal/intervals` package can fetch an athlete profile using the TP-001 config.
- Tests cover auth, retry, structured errors, and profile decoding without real network calls.
- API keys are never logged, echoed, or accepted from the LLM/tool args.
- `make test` passes.

## Do NOT

- Do not read, copy, paraphrase, transliterate, or port GPL/copyleft implementation code.
- Do not implement all ~30 launch tools.
- Do not add write endpoints, streams, activities, wellness, events, or workout-library APIs in this task.
- Do not add telemetry or debug metadata.
- Do not print, log, commit, or copy local `.env` values into fixtures/docs; `.env` is only for optional manual end-to-end validation.
- Do not introduce a GPL or copyleft dependency.

## Documentation

Must update:

- `STATUS.md`
- `CHANGELOG.md`

Check if affected:

- `README.md` only if config inputs changed

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-002`, for example: `TP-002 add intervals client retry errors`.

---

## Amendments

_Add amendments below this line only._
