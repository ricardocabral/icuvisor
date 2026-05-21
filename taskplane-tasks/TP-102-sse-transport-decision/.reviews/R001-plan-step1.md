# Review R001 — Plan review for Step 1

**Verdict:** Changes requested

The Step 1 checklist names the right outcomes, but it is too broad for a transport/client-compatibility decision with security impact. Before executing, expand the plan so the research is reproducible and clearly separated from the later product decision.

## Required changes

1. **Name the authoritative external sources to check.**
   - MCP transport spec/current guidance, including SSE deprecation/replacement status.
   - Official ChatGPT/OpenAI custom MCP connector documentation for whether it requires a publicly reachable endpoint and which transport(s) it accepts.
   - Any other target-client docs that are being used as evidence for “remote connector” behavior.
   - Record URLs, access date, and a short quoted/paraphrased fact summary in `STATUS.md`.

2. **Separate client surfaces.**
   The research plan should explicitly distinguish local ChatGPT/dev-mode MCP configuration from ChatGPT-style remote custom connector UIs. Current repo docs mention local stdio/HTTP shapes, while the roadmap item is about remote UIs that cannot reach `127.0.0.1`.

3. **Inventory implementation feasibility before recommending path A.**
   Add a sub-step to check whether `the MCP Go SDK` currently exposes an SSE server transport or whether SSE would require custom transport code. This does not mean implementing it in Step 1, but it materially affects the A/B recommendation.

4. **Make the security analysis concrete.**
   The plan should include reviewing the current icuvisor HTTP bind/auth behavior and documenting the risk model for tunneling: unauthenticated MCP access, registered tool/write/delete capabilities, exposure of the configured intervals.icu API key’s authority, public tunnel URLs, tunnel-provider access/logging, and whether any existing localhost/origin protections survive a tunnel. The summary must not imply tunneling is safe.

5. **Define the Step 1 deliverable.**
   Add acceptance criteria for the `STATUS.md` summary: evidence table, recommendation A or B with confidence, unresolved unknowns, and explicit note that Step 2 is where the product decision/approval happens. Step 1 should not edit protected product docs beyond `STATUS.md`.

## Non-blocking suggestion

Include the current repo baseline in the evidence table: PRD says “No SSE”, ROADMAP has the pending A/B decision, `web/content/connect/chatgpt.md` currently documents local stdio/HTTP alternatives, and `web/content/guides/http-transport.md` warns that LAN HTTP is unauthenticated.
