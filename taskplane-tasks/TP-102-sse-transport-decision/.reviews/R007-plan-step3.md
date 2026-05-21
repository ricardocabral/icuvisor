# Review R007 — Plan review for Step 3

**Verdict:** REVISE

The Step 3 plan is not ready to execute. `STATUS.md` only advances the current step/status and leaves the original A/B checklist unchanged; it does not contain a worker-specific plan for documenting the chosen Path B. There is also unresolved review-bookkeeping inconsistency from Step 2 that must be fixed before treating the Path B decision as approved.

## Blocking issues

1. **No concrete Step 3 plan is recorded.**
   - The only Step 3 content in `STATUS.md` is the generic prompt checklist, including two Path A bullets even though Step 2 chose Path B.
   - Add a Step 3 plan section that states this step will be documentation-only: no SSE implementation, no transport/config code changes, no PRD change unless explicitly escalated.

2. **Resolve prior review history before proceeding.**
   - `.reviews/R005-plan-step2.md` says `REVISE`, but `STATUS.md` records R005 as `APPROVE`.
   - `.reviews/R006-code-step2.md` also says `REVISE`, but `STATUS.md` records R006 as `APPROVE`.
   - Step 3 should not proceed on a falsely approved Step 2. Either correct the review table/execution log to match the artifacts and obtain the required follow-up approvals, or add subsequent approving review artifacts and record them accurately.

3. **Specify exactly which Path B docs will change and what they must say.**
   The plan should name the target files and intended message. At minimum:
   - `web/content/connect/chatgpt.md`: add a clear section distinguishing local ChatGPT/dev-mode MCP use from ChatGPT-style remote custom connector UIs. State that remote connectors require an HTTPS endpoint reachable from ChatGPT/OpenAI infrastructure, cannot reach `127.0.0.1`, and are intentionally unsupported until the vNext hosted relay or a future explicit secure-tunnel design.
   - `web/content/guides/http-transport.md`: add troubleshooting/security language so users do not treat remote-connector failure as an HTTP transport bug.
   - Do not add cloudflared/ngrok recipes for Path B, and do not imply URL secrecy or tunneling is an authentication boundary.

4. **Define Step 3 acceptance criteria.**
   Add concrete completion criteria such as:
   - Path A checklist items are marked `N/A` or otherwise not left as active unchecked work for this Path B task.
   - The docs explain the reachability limitation, the safety rationale, and the future relay path in user-facing language.
   - Existing local stdio and loopback Streamable HTTP instructions remain intact.
   - Any docs-only validation to be run in Step 4 is named, and no code tests are claimed for Step 3 unless code is changed.

## Non-blocking suggestions

- Keep `ROADMAP.md` and `CHANGELOG.md` for Step 4 if that boundary is intentional, but the Step 3 plan should explicitly say they are deferred so the omission is not accidental.
- Consider adding a short “not supported” callout near the top of the ChatGPT page, before the local configuration examples, because remote connector users are likely to scan for a URL first.
