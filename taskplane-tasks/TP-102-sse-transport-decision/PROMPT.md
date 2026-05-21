# TP-102 — SSE transport decision for remote-client compatibility

**Created:** 2026-05-20
**Size:** M

## Review Level: 2

**Assessment:** Transport/client-compatibility decision with potential security impact and documentation consequences.
**Score:** 5/8 — Blast radius: 2, Pattern novelty: 2, Security: 1, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-102-sse-transport-decision/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Decide and document whether icuvisor will support ChatGPT-style remote custom MCP connector UIs before the hosted relay. The task must choose one path: add an SSE transport behind the same auth/bind rules as HTTP with tunneling docs and explicit key-exposure warnings, or declare remote ChatGPT connectors out-of-scope until the vNext hosted relay and document the decision so users stop filing it as a bug.

## Dependencies

- **Task:** TP-033 (Streamable HTTP transport exists)
- **External:** MCP client behavior and current MCP spec/deprecation status must be checked before deciding

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- CLAUDE.md — repository rules, clean-room constraints, Go/MCP conventions.
- ROADMAP.md — exact roadmap entry and phase positioning.
- docs/prd/PRD-icuvisor.md — product contract, tool catalog, response-shaping rules.
- ROADMAP.md v0.4 SSE transport decision entry — exact decision text.
- docs/prd/PRD-icuvisor.md §6.2.B and KR6 — current transport/client compatibility scope.
- internal/mcp/transport.go and server.go — current transport wiring.
- web/content/guides/http-transport.md and connect/chatgpt docs — current remote-client docs.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None unless the task explicitly calls for live/manual validation; unit tests must not hit the network.

## File Scope

- `internal/mcp/transport*.go`
- `internal/mcp/server*.go`
- `internal/app/*`
- `internal/config/httpbind*.go`
- `web/content/guides/http-transport.md`
- `web/content/connect/chatgpt.md`
- `docs/deploy/*`
- `docs/prd/PRD-icuvisor.md`
- `ROADMAP.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm no protected docs are changed without explicit approval

### Step 1: Research current compatibility facts

- [ ] Verify current MCP transport guidance and whether target clients require reachable SSE endpoints vs Streamable HTTP.
- [ ] Confirm security implications of exposing a local BYO intervals.icu API key through tunnels.
- [ ] Summarize recommendation in STATUS.md.

### Step 2: Make the product decision

- [ ] Choose path A: implement SSE transport behind existing auth/bind rules with cloudflared/ngrok docs and warnings; or path B: declare remote connectors out-of-scope until vNext hosted relay.
- [ ] If the decision conflicts with PRD current "No SSE" language, prepare an explicit PRD/roadmap update.
- [ ] Get/record any needed operator approval before changing protected product docs.

### Step 3: Implement or document the chosen path

- [ ] For path A: add SSE transport wiring/config/tests without weakening localhost default or auth rules.
- [ ] For path A: add tunneling recipe and warnings about public internet exposure of a personal API key.
- [ ] For path B: update docs to explain why ChatGPT remote connectors cannot reach localhost Streamable HTTP and point to vNext hosted relay.

### Step 4: Verify and close loop

- [ ] Run transport/config tests and full quality gate if code changed.
- [ ] Update web docs, CHANGELOG.md, and ROADMAP.md decision status.
- [ ] Add troubleshooting language so users recognize this as an intentional decision, not a bug.

### Step 5: Testing & Verification

- [ ] Run targeted tests added/affected by this task
- [ ] Run FULL test suite: `make test`
- [ ] Build passes: `make build`
- [ ] Lint passes: `make lint`
- [ ] Fix all failures or document pre-existing unrelated failures in STATUS.md

### Step 6: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] `STATUS.md` discoveries and final notes updated
- [ ] Commit at step boundary with the task ID in the message

## Documentation Requirements

**Must Update:**
- CHANGELOG.md — record user-visible behavior under [Unreleased] if code or docs behavior changes.
- STATUS.md — keep execution state current.

**Check If Affected:**
- README.md — update if public setup/tool behavior changes.
- web/content/reference/tools.md — update if tool catalog descriptions or generated docs are affected.
- docs/prd/PRD-icuvisor.md — check only if behavior intentionally diverges from product scope.

## Completion Criteria

- A clear A/B decision is recorded in docs and CHANGELOG.md.
- If SSE is implemented, it is behind the same auth/bind safety rules and includes tests.
- If out-of-scope, ChatGPT remote-connector docs clearly explain the limitation and vNext relay path.
- No transport exposes beyond `127.0.0.1` by default.
- `make test`, `make build`, and `make lint` pass if code changed.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-102` for traceability. Examples:

- `feat(TP-102): complete step 1 — scope current behavior`
- `fix(TP-102): repair regression found during analyzer tests`
- `test(TP-102): add golden coverage for roadmap behavior`
- `hydrate: TP-102 expand step checkboxes`

## Do NOT

- Do not read, paste, paraphrase, or transliterate GPL/copyleft implementation code.
- Do not log or persist intervals.icu API keys outside the configured keychain path.
- Do not add LLM-controlled confirm/override flags for destructive behavior.
- Do not broaden the task into unrelated refactors; log follow-ups in STATUS.md discoveries.
- Do not expose HTTP/SSE beyond loopback by default.
- Do not imply tunneling is risk-free; explicitly warn about personal API-key exposure.
- Do not leave the decision undocumented.

---

## Amendments

_Add amendments below this line only._
