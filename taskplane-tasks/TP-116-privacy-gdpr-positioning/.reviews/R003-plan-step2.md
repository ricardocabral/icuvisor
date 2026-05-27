# Plan Review — Step 2: Add privacy-conscious user-facing copy

**Verdict: approved to proceed.**

The Step 2 plan is consistent with the approved Step 1 boundaries: add a standalone privacy explanation page, surface it from the explain index, and keep any edits to coach-mode/HTTP docs narrow. This addresses the task without depending on TP-113 or overloading homepage/local-first positioning.

## Implementation guardrails

- Do not claim “GDPR compliant,” certification, legal advice, or that no data ever leaves the machine. Prefer design-posture language and due-diligence questions for EU/privacy-conscious users.
- Be explicit about the trust boundary: icuvisor runs locally, API keys are kept out of chat and stored in the OS keychain by default, the local process contacts intervals.icu, and the chosen AI client/model provider may process conversation/tool-result content under its own terms.
- Keep `SECURITY.md` authoritative for security policy; link to it rather than duplicating long hardening/process details.
- Preserve the supervisor constraint from Step 0/1: avoid homepage and broad `local-first` positioning edits owned by TP-113. If a README pointer is added, keep it brief and factual.
- When mentioning HTTP, include the loopback default and the LAN-bind risk without implying HTTP is authenticated.
- When mentioning coach mode, keep clear that `athlete_id` is only a selector and the coach API key stays server-side.

## Minor follow-up

Step 3 should handle most cross-links and the changelog. It is fine if Step 2 creates the page and explain-index card first, then links from adjacent docs in the next step.
