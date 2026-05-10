# TP-004 — `get_athlete_profile` tool end-to-end wiring

## Mission

Implement the single v0.1 MCP tool, `get_athlete_profile`, by connecting the MCP registry from TP-003 to the intervals.icu client from TP-002. The tool must be terse-by-default, safe with credentials, and suitable for end-to-end stdio use.

Roadmap item: **One working tool: `get_athlete_profile`**.

Complexity: Blast radius 2, Pattern novelty 2, Security 2, Reversibility 1 = 7 → Review Level 2. Size: M.

## Dependencies

- **TP-002** — intervals.icu client
- **TP-003** — MCP stdio server/registry

## Context to Read First

- `CLAUDE.md`
- `docs/prd/PRD-icuvisor.md` — especially §7.2.C Athlete & fitness, §7.2.D Response shaping, §7.2.E tool descriptions, §7.4 #7
- `ROADMAP.md` — v0.1 only
- `CONTRIBUTING.md` — MCP tool conventions
- `SECURITY.md`
- `internal/intervals/` from TP-002
- `internal/mcp/` from TP-003
- `internal/config/` from TP-001
- Local `.env` file, if present, for `INTERVALS_ICU_ATHLETE_ID` and `INTERVALS_ICU_API_KEY` values used only for optional end-to-end validation through the MCP tool; never print or commit secrets
- `hhopke/intervals-icu-mcp` (MIT, Python) may be consulted as a non-dependency reference for `get_athlete_profile` semantics/response expectations; record consulted behavior in `STATUS.md`

## File Scope

Expected files:

- `internal/tools/get_athlete_profile.go`
- `internal/tools/get_athlete_profile_test.go`
- `internal/tools/registry.go` or existing registry files
- `internal/mcp/` only for integration adjustments
- `internal/intervals/` only for profile-field adjustments needed by the tool
- `CHANGELOG.md`
- `taskplane-tasks/TP-004-get-athlete-profile-tool/STATUS.md`

## Steps

### Step 1: Define the tool contract in STATUS.md

- [ ] Write the intended `get_athlete_profile` description, arguments, and response shape in `STATUS.md`
- [ ] Keep arguments minimal; do not accept API key as a tool parameter
- [ ] Decide whether v0.1 needs `include_full`; if included, default must remain terse
- [ ] Include units/timezone/athlete-ID conventions where data is available

### Step 2: Implement the typed tool

- [ ] Add typed request and response structs
- [ ] Register the tool under exactly `get_athlete_profile`
- [ ] Make the first sentence distinguish this tool as athlete profile/thresholds/zones, not activities or wellness
- [ ] Include JSON Schema descriptions an LLM can understand
- [ ] Call the intervals client with `context.Context` from the MCP request
- [ ] Return short actionable errors to the LLM; log/wrap detailed errors internally where supported

### Step 3: Shape the response for v0.1

- [ ] Return a terse useful profile: normalized `athlete_id`, display/name if available, FTP/thresholds/zones/sport settings when available, timezone and units when available
- [ ] Use disambiguating field names where applicable, e.g. include units in field names or `_meta`
- [ ] Include `_meta.server_version` so stale-schema conversations can be diagnosed later
- [ ] Do not include fetched timestamps or debug cruft by default
- [ ] Do not include secrets or raw upstream payloads by default

### Step 4: Add tests

- [ ] Test tool registration metadata: name, description, schema, no secret arguments
- [ ] Test successful handler response using a fake intervals client
- [ ] Test include-full/default behavior if implemented
- [ ] Test upstream errors map to short actionable tool errors
- [ ] Test `_meta.server_version` and normalized athlete ID are present

### Step 5: End-to-end local verification

- [ ] Add or update an integration test that exercises MCP stdio tool call dispatch to `get_athlete_profile` with a fake client/server
- [ ] If `.env` contains `INTERVALS_ICU_ATHLETE_ID` and `INTERVALS_ICU_API_KEY`, run a clearly separated manual/end-to-end MCP validation against intervals.icu and record only pass/fail plus non-sensitive response shape in `STATUS.md`
- [ ] Run `go fmt ./...`
- [ ] Run `make test`
- [ ] Run `make build`
- [ ] Run `make lint` if available
- [ ] Update `CHANGELOG.md` and `STATUS.md`

## Reference Implementation Policy

- You may use `hhopke/intervals-icu-mcp` as a permissively licensed, Python-only reference implementation for intervals.icu API reverse-engineering and MCP tool semantics. It must not be added as a dependency. When reading it, extract behavior and API contracts, then implement idiomatic Go from first principles. Record the consulted files/behaviors in `STATUS.md`.
- You may not use `mvilanova/intervals-mcp-server` source code as an implementation reference because it is GPLv3 and this project is MIT clean-room. Do not read, copy, paraphrase, transliterate, or port its code. Use the PRD's summarized insights, intervals.icu public docs, black-box testing, and permissively licensed references instead.

## Acceptance Criteria

- `get_athlete_profile` appears in `tools/list` and can be called through MCP stdio plumbing.
- The handler fetches profile data through `internal/intervals`, not by doing ad-hoc HTTP.
- Tool input schema does not expose credentials.
- Response is terse, includes `_meta.server_version`, and uses normalized athlete ID.
- Tests cover success and error paths.

## Do NOT

- Do not implement `list_athletes`, coach mode, activities, wellness, events, or other v0.5/v1 tools.
- Do not add destructive operations.
- Do not add toolset tiers yet unless the MCP registry requires a harmless stub.
- Do not return raw full upstream payloads by default.
- Do not print, log, commit, or copy local `.env` values into fixtures/docs; `.env` is only for optional manual end-to-end validation.
- Do not read, copy, paraphrase, transliterate, or port GPL implementation code, including `mvilanova/intervals-mcp-server`.

## Documentation

Must update:

- `STATUS.md`
- `CHANGELOG.md`

Check if affected:

- `README.md` only if quickstart examples are added here instead of TP-005

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-004`, for example: `TP-004 add get athlete profile tool`.

---

## Amendments

_Add amendments below this line only._
