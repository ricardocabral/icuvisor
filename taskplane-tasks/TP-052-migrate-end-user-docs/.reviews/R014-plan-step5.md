# R014 Plan Review — Step 5: Guides section

Verdict: APPROVE

The Step 5 plan matches the prompt: it creates the API-key, Streamable HTTP, coach-mode, after-upgrade, and troubleshooting guides; it pulls from the intended README/docs sources; and it keeps the work in the end-user guide section rather than changing README or deleting source docs.

Non-blocking implementation notes:

1. Keep `icuvisor setup` as the recommended credential path in `guides/api-key.md`. Manual platform keychain commands should be clearly advanced/headless fallback material, and examples must use service `icuvisor` and account `intervals-icu-api-key`. Do not put API keys into MCP client JSON snippets.
2. In `guides/http-transport.md`, make stdio/loopback the default path: `ICUVISOR_TRANSPORT=http` or `--transport http` should bind to `127.0.0.1:8765` and serve `/mcp`. Any LAN bind example must include the explicit warning that the local MCP HTTP server is unauthenticated and exposes the configured intervals.icu credentials to anyone who can connect.
3. When adapting `docs/coach-mode.md`, do not copy stale ACL examples that put `icuvisor_list_advanced_capabilities` inside per-athlete `allowed_tools`. Step 2/R008 already established that it is a meta/control tool, not an athlete-scoped ACL entry. Use athlete-scoped patterns such as `get_*`, exact athlete-scoped tool names, or `*`.
4. Keep the coach-mode guide procedural: enabling `ICUVISOR_COACH_MODE`, adding the `coach` JSON stanza, choosing `on` versus `auto`, restarting the MCP server/client, and using `list_athletes` / `select_athlete`. Save the conceptual threat model, `athlete_id` selector explanation, and ACL philosophy for Step 6's explanation page, except where a short safety note is needed.
5. In `guides/after-upgrade.md`, translate `_meta.schema_changed` into the user action: start a new AI-client conversation or reconnect so the client refreshes the tool catalog. Avoid implying icuvisor performs auto-update checks.
6. For `guides/troubleshooting.md`, include the required symptoms from the prompt and link fixes to the existing install/connect/reference pages with Hugo `relref`. Since Step 6 pages do not exist yet, either avoid linking to them during Step 5 or create only links that will be satisfied before the next Hugo build.
7. Link env-var mentions to `reference/cli.md` or `reference/safety-modes.md`, config fields to `reference/config-file.md`, and tool names to generated anchors in `reference/tools.md`. Avoid bare links to migrated `docs/` pages.

With those guardrails, the plan is ready to implement.
