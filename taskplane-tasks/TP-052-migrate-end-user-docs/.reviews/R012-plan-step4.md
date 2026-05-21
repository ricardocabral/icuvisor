# R012 Plan Review — Step 4: Connect section

Verdict: APPROVE

The Step 4 plan matches the prompt: it replaces the placeholder connect index, migrates the two user-facing Claude client guides, adds a focused ChatGPT connection page, and creates an "other clients" page for Cursor / Continue / Zed / Pi-style MCP clients. It also preserves the Step 1 decision not to migrate `docs/clients/codex-local.md`.

Non-blocking implementation notes:

1. Keep API keys out of all MCP client JSON examples. The snippets should contain only the command path and non-secret configuration such as `INTERVALS_ICU_ATHLETE_ID`, `ICUVISOR_TIMEZONE`, and optionally `ICUVISOR_TRANSPORT`; point users to `icuvisor setup` / OS keychain rather than `INTERVALS_ICU_API_KEY` for normal setup.
2. Use Hugo `relref` for internal links. Link env-var mentions to `reference/cli.md` or `reference/safety-modes.md`, and link any tool names to the generated `reference/tools.md` anchors.
3. For the Claude Desktop smoke checklist, a plain `<details><summary>...</summary>` block is appropriate; do not add JavaScript or a new build step.
4. Be precise about transport. The default examples should use local stdio execution of `/Applications/icuvisor.app/Contents/MacOS/icuvisor`. If a client requires Streamable HTTP or a URL, show loopback-only configuration by default (`127.0.0.1:8765`) and include the LAN exposure warning required by `CLAUDE.md` rather than implying remote/LAN binding is normal.
5. Avoid creating broken per-step `relref`s to Step 5 guide pages that do not exist yet unless the implementer also creates those targets before running Hugo. For Step 4, existing safe targets include the install pages and the reference pages.
6. Keep the pages end-user-oriented and macOS-specific where the current installer reality is macOS only; for unsupported platform/client combinations, say what is currently available instead of inventing installers or one-click configs.

With those guardrails, the plan is ready to implement.
