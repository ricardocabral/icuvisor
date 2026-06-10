# TP-162: Codex Streamable HTTP JSON-RPC smoke coverage — Status

**Current Step:** Step 1: Add Streamable HTTP JSON-RPC handshake smoke tests
**Status:** 🟡 In Progress
**Last Updated:** 2026-06-10
**Review Level:** 2
**Review Counter:** 3
**Iteration:** 1
**Size:** M

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied

---

### Step 1: Add Streamable HTTP JSON-RPC handshake smoke tests
**Status:** 🟨 In Progress

- [x] Initialize response envelope test added
- [x] Ping response envelope test added
- [x] Raw in-process HTTP wire assertions parse JSON or SSE `data:` envelopes instead of SDK-decoded results
- [x] Handshake lifecycle covers initialize, `notifications/initialized` with session ID, then ping using Codex-like headers
- [x] Success assertions reject bare payloads and top-level errors before inspecting `result`
- [ ] R003 revision: success envelope assertions fail whenever a top-level `error` member is present, even `null`
- [x] Codex-like HTTP headers covered without external process
- [x] Targeted MCP tests pass

---

### Step 2: Fix transport/protocol behavior only if tests fail
**Status:** ⬜ Not Started

- [ ] Transport/server fixes applied if needed
- [ ] Stdio and loopback defaults preserved
- [ ] Protocol errors remain short/actionable
- [ ] Targeted MCP tests pass

---

### Step 3: Testing & Verification
**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Integration tests (if applicable)
- [ ] All failures fixed
- [ ] Build passes

---

### Step 4: Documentation & Delivery
**Status:** ⬜ Not Started

- [ ] `docs/clients/codex-local.md` reviewed/updated if affected
- [ ] `CHANGELOG.md` updated
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |
| R002 | Plan | 1 | APPROVE | `.reviews/R002-plan-step1.md` |
| R003 | Code | 1 | REVISE | `.reviews/R003-code-step1.md` |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-06-09 | Task staged | PROMPT.md and STATUS.md created |
| 2026-06-10 12:17 | Task started | Runtime V2 lane-runner execution |
| 2026-06-10 12:17 | Step 0 started | Preflight |
| 2026-06-10 12:17 | Step 0 completed | Required paths and Go module dependencies verified |
| 2026-06-10 12:17 | Step 1 started | Streamable HTTP JSON-RPC smoke tests |
| 2026-06-10 12:18 | Step 1 plan reviewed | R001 REVISE, R002 APPROVE after raw-wire checklist hydration |
| 2026-06-10 12:19 | Targeted MCP tests | `go test ./internal/mcp -run 'Streamable|JSONRPC|Codex|Protocol|Ping|Initialize'` passed |

---

## Blockers

*None*

---

## Notes

Public signal: Montis forum #512 and #516-518 described Codex handshake failure until JSON-RPC 2.0 wrapping was strict.
R001 plan review requires raw in-process HTTP assertions against the wire response, parsing current SSE `data:` JSON-RPC envelopes if necessary, and a full initialize/initialized/ping session lifecycle with Codex-like headers.
R003 code review requires rejecting any top-level `error` member on successful JSON-RPC responses, including `error: null`.
| 2026-06-10 12:21 | Review R001 | plan Step 1: REVISE |
| 2026-06-10 12:23 | Review R002 | plan Step 1: APPROVE |
| 2026-06-10 12:29 | Review R003 | code Step 1: REVISE |
