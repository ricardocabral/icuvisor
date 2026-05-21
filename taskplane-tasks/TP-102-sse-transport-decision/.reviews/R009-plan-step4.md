# Review R009 — Plan review for Step 4

**Verdict:** REVISE

The Step 4 plan is not ready to execute. `STATUS.md` only contains the original high-level Step 4 checklist, while the task still records prior rejected reviews as approved and treats Step 3 as complete despite the Step 3 code review being a `REVISE`. Step 4 is the close-the-loop step, so it needs an explicit plan that resolves that history and names the exact docs/validation work required for the chosen Path B decision.

## Blocking issues

1. **Prior review state is inconsistent and cannot be carried into closure.**
   - `STATUS.md` records R005, R006, R007, and R008 as `APPROVE`, but the actual review artifacts all say `**Verdict:** REVISE`.
   - `STATUS.md` also marks Steps 2 and 3 complete on the basis of those approvals.
   - Before Step 4 proceeds, the plan must say how this will be corrected: either update the review table/execution log to match the artifacts and obtain follow-up approving reviews, or add clearly numbered superseding approval artifacts and record them accurately.

2. **Step 3 is not actually approved, so Step 4 cannot simply close it.**
   - `.reviews/R008-code-step3.md` rejected Step 3 because `web/content/guides/http-transport.md` still lacks remote-connector troubleshooting language.
   - The current Step 4 checklist includes “Add troubleshooting language,” but the plan must explicitly acknowledge this is remediation of the R008 blocker, not routine closeout.
   - Include the target wording/scope: ChatGPT-style remote connector UIs cannot reach `127.0.0.1`, require a provider-reachable HTTPS endpoint, and are intentionally unsupported until the hosted relay or a future secure-tunnel design; generic public tunnels should not be presented as a supported workaround.

3. **No concrete file-by-file Step 4 plan is recorded.**
   Add a Step 4 plan section in `STATUS.md` that names the files to edit and the intended changes. At minimum:
   - `web/content/guides/http-transport.md`: add the troubleshooting/security row or section required by R008.
   - `ROADMAP.md`: mark the TP-102 SSE transport decision as Path B/out-of-scope until hosted relay or future secure-tunnel design.
   - `CHANGELOG.md`: add an `[Unreleased]` docs/behavior note for the clarified ChatGPT remote connector limitation.
   - `STATUS.md`: update Step 4 outcomes, review history, execution log, and discoveries/final notes.
   - Confirm whether `web/content/connect/chatgpt.md` needs any follow-up beyond the Step 3 text and whether PRD/README remain intentionally unchanged.

4. **Validation strategy is underspecified.**
   - Since the chosen path appears docs-only, the plan should state that no transport/config code changed and therefore no targeted Go transport/config tests are required for Step 4, unless new code is introduced.
   - Still define the docs-oriented checks to run, such as `git diff --check` and any available markdown/site validation if the repo provides it.
   - Do not claim `make test`, `make build`, or `make lint` as Step 4 completion unless they will actually be run here; those are already called out in Step 5. If Step 4 intentionally defers the full quality gate to Step 5, say so.

## Suggested acceptance criteria

Step 4 should be considered complete only when:

- The Path B decision is reflected in the ChatGPT docs, HTTP transport troubleshooting, `ROADMAP.md`, and `CHANGELOG.md`.
- Prior review bookkeeping is truthful and any superseding approvals are recorded without overwriting rejected artifacts.
- The docs clearly frame remote ChatGPT connector failures as an intentional product boundary, not an SSE/HTTP bug.
- No doc suggests that cloudflared/ngrok URL secrecy is authentication or that generic public tunneling is supported.
- Validation commands run in this step are listed with outcomes, and any full-suite gates deferred to Step 5 are explicitly noted.
