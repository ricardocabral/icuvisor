# TP-102: SSE transport decision for remote-client compatibility — Status

**Current Step:** Step 6: Documentation & Delivery
**Status:** ✅ Complete
**Last Updated:** 2026-05-20
**Review Level:** 2
**Review Counter:** 13
**Iteration:** 2
**Size:** M

> **Hydration:** Checkboxes represent meaningful outcomes, not individual code changes. Workers may expand steps when runtime discoveries warrant it.

---

### Step 0: Preflight
**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Confirm no protected docs are changed without explicit approval

---

### Step 1: Research current compatibility facts
**Status:** ✅ Complete

- [x] Verify current MCP transport guidance and whether target clients require reachable SSE endpoints vs Streamable HTTP.
  - Evidence must name authoritative external sources, URLs, access date, and short fact summaries.
  - Evidence must distinguish local ChatGPT/dev-mode MCP configuration from ChatGPT-style remote custom connector UIs.
- [x] Inventory repo and SDK feasibility for path A, including whether the MCP Go SDK exposes an SSE server transport or SSE would require custom code.
- [x] Confirm security implications of exposing a local BYO intervals.icu API key through tunnels.
  - Security notes must cover unauthenticated MCP access, write/delete tool capabilities, intervals.icu API-key authority, public tunnel URLs, tunnel-provider access/logging, and whether localhost/origin protections survive tunneling.
- [x] Summarize recommendation in STATUS.md.
  - Summary must include an evidence table, recommendation A or B with confidence, unresolved unknowns, and an explicit note that Step 2 is the product decision/approval step.
- [x] Fix Step 1 review bookkeeping and execution-log Markdown per R003.

---

### Step 2: Make the product decision
**Status:** ✅ Complete

- [x] Choose path A: implement SSE transport behind existing auth/bind rules with cloudflared/ngrok docs and warnings; or path B: declare remote connectors out-of-scope until vNext hosted relay.
- [x] If the decision conflicts with PRD current "No SSE" language, prepare an explicit PRD/roadmap update.
- [x] Get/record any needed operator approval before changing protected product docs.

---

### Step 3: Implement or document the chosen path
**Status:** ✅ Complete

- [x] For path A: add SSE transport wiring/config/tests without weakening localhost default or auth rules. (N/A: Step 2 chose path B; `web/content/connect/chatgpt.md` now states icuvisor intentionally does not add legacy SSE for this case.)
- [x] For path A: add tunneling recipe and warnings about public internet exposure of a personal API key. (N/A: Step 2 chose path B; `web/content/connect/chatgpt.md` warns against generic public tunneling rather than documenting it as supported.)
- [x] For path B: update docs to explain why ChatGPT remote connectors cannot reach localhost Streamable HTTP and point to vNext hosted relay.

---

### Step 4: Verify and close loop
**Status:** ✅ Complete

- [x] Run transport/config tests and full quality gate if code changed.
- [x] Correct review bookkeeping and record the R009 remediation plan.
- [x] Update web docs, CHANGELOG.md, and ROADMAP.md decision status.
- [x] Add troubleshooting language so users recognize this as an intentional decision, not a bug.
- [x] Run Step 4 docs validation and record any full-suite deferral to Step 5.

---


### Step 5: Testing & Verification
**Status:** ✅ Complete

- [x] Targeted tests passing
- [x] FULL test suite passing: `make test`
- [x] Build passes: `make build`
- [x] Lint passes: `make lint`
- [x] All failures fixed or documented as pre-existing unrelated failures

---

### Step 6: Documentation & Delivery
**Status:** ✅ Complete

- [x] "Must Update" docs modified
- [x] "Check If Affected" docs reviewed
- [x] Discoveries logged
- [x] Final commit includes task ID

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | 1 | REVISE | .reviews/R001-plan-step1.md |
| R002 | Plan | 1 | APPROVE | .reviews/R002-plan-step1.md |
| R003 | Code | 1 | REVISE | .reviews/R003-code-step1.md |
| R004 | Code | 1 | APPROVE | .reviews/R004-code-step1.md |
| R005 | Plan | 2 | REVISE | .reviews/R005-plan-step2.md |
| R006 | Code | 2 | REVISE | .reviews/R006-code-step2.md |
| R007 | Plan | 3 | REVISE | .reviews/R007-plan-step3.md |
| R008 | Code | 3 | REVISE | .reviews/R008-code-step3.md |
| R009 | Plan | 4 | REVISE | .reviews/R009-plan-step4.md |
| R010 | Plan | 4 | APPROVE | .reviews/R010-plan-step4.md |
| R011 | Code | 4 | APPROVE | .reviews/R011-code-step4.md |
| R012 | Plan | 5 | APPROVE | .reviews/R012-plan-step5.md |
| R013 | Code | 5 | APPROVE | .reviews/R013-code-step5.md |

---

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| No new out-of-scope discoveries beyond the recorded Path B remote-connector boundary. | No follow-up required for this task. | Step 6 |

---

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-05-20 | Task staged | PROMPT.md and STATUS.md created |
| 2026-05-20 14:07 | Task started | Runtime V2 lane-runner execution |
| 2026-05-20 14:07 | Step 0 started | Preflight |
| 2026-05-20 14:10 | Review R001 | plan Step 1: REVISE |
| 2026-05-20 14:11 | Review R002 | plan Step 1: APPROVE |
| 2026-05-20 14:18 | Review R003 | code Step 1: REVISE |
| 2026-05-20 14:20 | Review R004 | code Step 1: APPROVE |
| 2026-05-20 14:22 | Review R005 | plan Step 2: REVISE |
| 2026-05-20 14:24 | Review R006 | code Step 2: REVISE |
| 2026-05-20 14:28 | Review R007 | plan Step 3: REVISE |
| 2026-05-20 14:31 | Review R008 | code Step 3: REVISE |
| 2026-05-20 14:38 | Review R009 | plan Step 4: REVISE |
| 2026-05-20 14:50 | Worker iter 1 | done in 2563s, tools: 122 |
| 2026-05-20 14:55 | Review R010 | plan Step 4: APPROVE |
| 2026-05-20 14:58 | Review R011 | code Step 4: APPROVE |
| 2026-05-20 15:00 | Review R012 | plan Step 5: APPROVE |
| 2026-05-20 15:03 | Review R013 | code Step 5: APPROVE |
| 2026-05-20 15:05 | Worker iter 2 | done in 883s, tools: 80 |
| 2026-05-20 15:05 | Task complete | .DONE created |

---

## Blockers

*None*

---

## Notes

- Step 1 plan-review suggestion to include repo baseline in evidence table: PRD currently says “No SSE”; ROADMAP has pending A/B decision; `web/content/connect/chatgpt.md` documents local stdio/HTTP alternatives; `web/content/guides/http-transport.md` warns LAN HTTP is unauthenticated.

### Step 1 research evidence (accessed 2026-05-20)

| Source | Surface | Fact summary |
|--------|---------|--------------|
| https://modelcontextprotocol.io/specification/2025-06-18/basic/transports | MCP spec/current guidance | Streamable HTTP replaces the 2024-11-05 HTTP+SSE transport. Streamable HTTP uses a single MCP endpoint with POST/GET and may use SSE only as the streaming mechanism. Backwards compatibility may keep deprecated HTTP+SSE endpoints. The spec warns servers to validate `Origin`, bind locally when local, and implement auth. |
| https://developers.openai.com/api/docs/guides/tools-connectors-mcp | OpenAI API remote MCP/connectors | Remote MCP servers require a `server_url`; examples use HTTPS `/sse`; docs state remote MCP servers can be any public-Internet server implementing remote MCP and the Responses API works with Streamable HTTP or HTTP/SSE. |
| https://developers.openai.com/apps-sdk/concepts/mcp-server | ChatGPT Apps SDK MCP concept | Apps SDK can host MCP over Server-Sent Events or Streamable HTTP, but recommends Streamable HTTP. It positions remote MCP as ChatGPT web/mobile compatible. |
| https://developers.openai.com/apps-sdk/build/mcp-server | ChatGPT developer-mode app setup | ChatGPT requires HTTPS; during development, docs tell users to tunnel localhost with ngrok or similar and use the ngrok URL when creating a connector in ChatGPT developer mode. This is distinct from local stdio/loopback client configs. |
| https://developers.openai.com/api/docs/guides/secure-mcp-tunnels | OpenAI Secure MCP Tunnel | OpenAI documents an outbound tunnel path for private MCP servers so ChatGPT/Codex/API can use a private/local server without making the MCP server public. This is OpenAI infrastructure, not the same as exposing icuvisor directly through cloudflared/ngrok. |
| `docs/prd/PRD-icuvisor.md` §6.2.B | Repo baseline | Current product contract says stdio default, Streamable HTTP loopback/LAN opt-in, and “No SSE — deprecated in the MCP spec; not implemented.” |
| `ROADMAP.md` v0.4 | Repo baseline | TP-102 is explicitly the pending A/B decision for SSE+tunnel docs vs remote ChatGPT connectors out-of-scope until hosted relay. |
| `web/content/connect/chatgpt.md` | Repo baseline/local ChatGPT | Current docs only cover local stdio config or local Streamable HTTP at `http://127.0.0.1:8765/mcp`; they do not explain ChatGPT remote connector HTTPS reachability. |
| `web/content/guides/http-transport.md` | Repo baseline/security | Current Streamable HTTP guide says LAN binding exposes an unauthenticated MCP server to anyone who can connect, using this process’s intervals.icu credentials. |

### Step 1 implementation-feasibility inventory

- `the MCP Go SDK v1.6.0` includes `mcp.NewSSEHandler`, `SSEOptions`, `SSEServerTransport`, and `SSEClientTransport` for the deprecated 2024-11-05 HTTP+SSE transport, so path A would not require a clean-room custom transport implementation.
- The SDK SSE handler has localhost Host/DNS-rebinding protection by default (`DisableLocalhostProtection` false) but, unlike the Streamable HTTP handler options used by icuvisor, it exposes no explicit `CrossOriginProtection` option in `SSEOptions`.
- Current icuvisor transport config accepts only `stdio` or `http`; `http` runs `Server.RunStreamableHTTP` on a configurable bind address and path `/mcp`.
- Current default bind is `127.0.0.1:8765`; non-loopback binds are accepted only as explicit IP:port and produce a warning, but there is no MCP-level authentication on the local HTTP transport.
- Path A is feasible with existing SDK primitives, but it would add a deprecated transport surface and would need careful docs/tests to preserve loopback default and avoid implying that tunnels are safe.

### Step 1 tunnel/security risk model

- Current local Streamable HTTP is unauthenticated MCP access: anyone who can reach the listener can list and invoke every registered tool/resource/prompt for that process.
- Registered capabilities depend on runtime config: `ICUVISOR_TOOLSET` controls breadth and `ICUVISOR_DELETE_MODE` gates destructive deletes at registration time. A tunnel would expose whichever read/write/delete capabilities are registered, not just harmless reads.
- The intervals.icu API key remains in server config/keychain rather than the model conversation, but reachable MCP callers inherit that key’s authority through tool calls. If the key can write/delete in intervals.icu, exposed tools can exercise that authority.
- A public tunnel URL changes the threat model from same-host/LAN to internet-reachable. URL secrecy is not authentication, and tunnel-provider dashboards/logs/control planes become part of the trust boundary.
- Existing loopback binding and SDK localhost Host protections protect local-only use; a tunnel terminates outside the process and forwards to localhost, so public reachability bypasses the “only local clients can connect” assumption. Origin/Host checks may still reject some browser DNS-rebinding attacks, but they do not authenticate the remote MCP caller that the tunnel forwards.
- OpenAI’s Secure MCP Tunnel is a safer-looking pattern than generic public tunnels because it avoids inbound public exposure, but it is not currently integrated/documented as an icuvisor supported path and still requires treating OpenAI’s tunnel and connector permissions as trusted infrastructure.
- Step 1 security conclusion: do not describe cloudflared/ngrok tunneling as safe. If path A is chosen later, docs must say it intentionally exposes a personal intervals.icu control surface and should be used only with least-privilege tool/delete config and trusted tunnel accounts.

### Step 1 recommendation summary

Recommendation: **Path B — declare ChatGPT-style remote custom connector use out-of-scope until the vNext hosted relay or a deliberate secure-tunnel integration**, with **medium-high confidence**.

Rationale:
- Current MCP guidance says Streamable HTTP is the replacement transport and only keeps HTTP+SSE for backwards compatibility. OpenAI’s current docs accept Streamable HTTP for remote MCP and Apps SDK recommends it, so adding deprecated SSE is not the core missing capability.
- The real incompatibility is reachability/security: ChatGPT remote connector UIs require an HTTPS endpoint reachable from OpenAI/ChatGPT infrastructure, while icuvisor’s safe default is local stdio or loopback HTTP.
- Generic cloudflared/ngrok recipes would make a personal intervals.icu MCP control surface internet-reachable while the server has no MCP auth layer. That conflicts with the local-first/keychain safety posture unless explicitly treated as unsupported/high-risk.
- The existing PRD “No SSE” position remains aligned with the current MCP spec and does not need to change for Step 1’s facts.

Unresolved unknowns for Step 2:
- Whether the operator wants to support OpenAI Secure MCP Tunnel as a separate future path before the hosted relay.
- Whether any target non-OpenAI client still requires legacy HTTP+SSE despite OpenAI’s current Streamable HTTP support.
- Whether consumer ChatGPT connector UI behavior exactly matches the API/Apps SDK documentation for all accounts and regions.

Step 1 is evidence and recommendation only. **Step 2 remains the product decision/approval point** before changing PRD/roadmap/product docs.

### Step 2 decision record

Decision: **Path B**. icuvisor will not add legacy SSE transport for ChatGPT-style remote custom connector UIs in this task. Remote connector use is out-of-scope until the vNext hosted relay or a future explicit secure-tunnel design. Local clients remain supported through stdio and loopback Streamable HTTP; remote HTTPS reachability is an intentional product boundary, not a transport bug.

PRD/roadmap impact: This decision **does not conflict** with `docs/prd/PRD-icuvisor.md` §6.2.B because the PRD already says SSE is deprecated and not implemented. It does require a roadmap/docs update to mark the TP-102 decision as Path B and explain the remote-connector limitation.

Approval record: `.pi/taskplane-config.json` defines no `protectedDocs`, and Step 2 does not change the PRD. No operator approval is required before updating non-protected web docs, `ROADMAP.md`, and `CHANGELOG.md` in later steps. If a future change edits the PRD or changes the product contract beyond Path B, it should be escalated first.

### Step 4 remediation plan after R009

Review bookkeeping:
- R005-R009 are now recorded in the review table and execution log with the verdicts shown in their artifacts. Prior status entries had incorrectly recorded R005-R008 as approvals.
- Steps 2 and 3 remain marked complete because the current orchestration prompt identifies them as completed/do-not-redo; Step 4 will close the remaining R008/R009 issues by adding the missing HTTP troubleshooting coverage, recording the Path B decision in roadmap/changelog, and obtaining Step 4 plan/code approval before completion.

File-by-file plan:
- `web/content/connect/chatgpt.md`: review the Step 3 Path B text; no expected follow-up unless the final diff lacks clear local-vs-remote connector wording.
- `web/content/guides/http-transport.md`: add remote-connector troubleshooting/security language explaining that ChatGPT-style remote connector UIs cannot reach `127.0.0.1`, require a provider-reachable HTTPS endpoint, and are intentionally unsupported until the hosted relay or a future secure-tunnel design. Do not present generic public tunnels as supported authentication.
- `ROADMAP.md`: mark the TP-102 SSE transport decision as Path B/out-of-scope until the hosted relay or future secure-tunnel design.
- `CHANGELOG.md`: add an `[Unreleased]` note documenting the clarified ChatGPT remote connector limitation.
- `STATUS.md`: keep the review history truthful, record validation outcomes, and note whether PRD/README/tool reference remain intentionally unchanged.

Validation strategy:
- Step 4 is docs/status-only unless new code is introduced. Transport/config targeted tests already passed with `go test ./internal/config ./internal/mcp`; no additional Go targeted tests are required for Step 4.
- Run `git diff --check` after docs edits. Full `make test`, `make build`, and `make lint` are deferred to Step 5 as the task-level quality gate.

### Step 4 verification notes

- Transport/config targeted tests run: `go test ./internal/config ./internal/mcp` passed. Full quality gate deferred to Step 5; no Go source changed in Steps 3-4.
- Docs validation run: `git diff --check` passed. Full `make test`, `make build`, and `make lint` remain deferred to Step 5.
- Code review R011 approved Step 4 and also ran `make web-build`, which passed with existing Hugo deprecation warnings only.

### Step 5 verification notes

- Targeted tests run: `go test ./internal/config ./internal/mcp` passed; `make web-build` passed with existing Hugo deprecation warnings for `.Site.Data` and `.Language.LanguageDirection`.
- Full test suite run: `make test` passed.
- Build run: `make build` passed.
- Lint run: `make lint` passed.
- No failing checks remain; no pre-existing unrelated failures needed documentation.
- Code review R013 approved Step 5 and independently reran targeted tests, `make test`, `make build`, `make lint`, and `make web-build`.

### Step 6 delivery notes

- Must-update docs modified: `CHANGELOG.md` and `STATUS.md`.
- Check-if-affected docs reviewed: `README.md` needs no change because it only describes local-client/project overview; `web/content/reference/tools.md` needs no change because no tool catalog changed; `docs/prd/PRD-icuvisor.md` needs no change because the Path B decision aligns with the existing “No SSE” transport contract.
