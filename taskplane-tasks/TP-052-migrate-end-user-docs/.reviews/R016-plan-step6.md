# R016 Plan Review — Step 6: Explanation section

Verdict: APPROVE

The Step 6 plan matches the prompt: it replaces the explanation index and authors the five required conceptual pages (`what-is-mcp`, `local-first`, `terse-by-default`, `safety-modes`, and `coach-mode`). It also correctly keeps the reference tables and procedural setup content in the pages already authored in Steps 2 and 5, with Step 6 focused on the “why”.

Non-blocking implementation notes:

1. Keep explanation pages conceptual and avoid duplicating the detailed tables in `reference/safety-modes.md` or the setup checklist in `guides/coach-mode.md`. Link to those pages with Hugo `relref` instead.
2. In `local-first.md`, avoid overclaiming that training data “never leaves the machine” in an absolute sense. A safer framing is: icuvisor does not run a hosted relay and does not send your API key or data to icuvisor-operated servers; it talks from the local process to intervals.icu and returns selected tool responses to the MCP client, and cloud AI clients may send those responses to their model provider under that client’s policy.
3. In `safety-modes.md`, use the code-audited behavior from Step 1: unknown or empty `ICUVISOR_DELETE_MODE` falls back to `safe`, safe mode registers non-destructive writes but no delete tools, and destructive tools are gated at registration time rather than by a model-controlled `confirm` argument. Do not repeat the stale PRD table language.
4. In `terse-by-default.md`, explain `include_full` as an explicit opt-in for heavier payloads and connect the idea to token budget, null stripping, resources/prompts, pagination, and the `core`/`full` toolset split. Link env-var/tier details to `reference/safety-modes.md` rather than restating every value.
5. In `coach-mode.md`, keep the threat-model summary user-facing: the coach-scoped key stays server-held; `athlete_id` is a target selector, not a credential; target IDs are normalized and checked against the roster before upstream calls; per-athlete ACLs compose with delete mode and toolset tier. Do not link to or migrate `docs/threat-models/coach-mode.md`, since developer/security detail stays in `docs/` for this task.
6. Link all tool-name mentions to generated anchors in `reference/tools.md` and all env-var mentions to `reference/cli.md` or `reference/safety-modes.md`. Use `relref` for internal links and avoid bare links to migrated `docs/` sources.
7. For `what-is-mcp.md`, keep the promised non-technical one-paragraph introduction and link externally to `https://modelcontextprotocol.io`; avoid inventing client behavior beyond the connect pages already written.

With those guardrails, the plan is ready to implement.
