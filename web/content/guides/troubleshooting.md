---
title: "Troubleshooting"
description: "Common icuvisor setup symptoms and fixes."
---

Use this guide when icuvisor does not start, an MCP client cannot see tools, or a connection works differently after an upgrade.

## Stale conversations and cached tool catalogs

MCP clients can keep a copy of icuvisor's tool catalog and sometimes keep using context from the current chat. That is useful for speed, but it means a long-running conversation may not immediately notice that you upgraded icuvisor, changed `ICUVISOR_TOOLSET` or `ICUVISOR_DELETE_MODE`, edited your timezone, fixed zone settings in intervals.icu, or restarted the server with a different config. The binary can be correct while the open chat is still reasoning from yesterday's tools or assumptions.

Before changing credentials or pasting details into a chat, try these safe first steps:

1. Start a new conversation in the MCP client. This is the quickest way to force the assistant to stop using stale chat context.
2. Refresh, reconnect, or reload MCP tools in the client, then fully restart the client if tools still look wrong. Different clients use different labels for this action.
3. Verify the running binary with `icuvisor version` and confirm your client config points at that binary.
4. Run `icuvisor diagnostics` and review the redacted output for the active transport, toolset, delete mode, timezone, config source, and credential source.

Common stale-state symptoms include:

- The assistant reports the wrong local date, day boundary, or timezone after you changed `ICUVISOR_TIMEZONE`.
- FTP, heart-rate, pace, or power zones look old after you updated settings in intervals.icu.
- A tool added in a new release is missing, an old tool schema is still used, or a delete/write tool remains hidden after changing toolset or delete-mode settings.
- Writes fail with arguments that no longer match the current docs, or the assistant keeps retrying a field that a new schema removed or renamed.

Treat API keys as credentials, not troubleshooting data. Do not paste an intervals.icu API key into an assistant conversation, issue report, or manual MCP JSON config. Use `icuvisor setup`, the OS keychain, or your client's sensitive extension setting instead. `icuvisor diagnostics` is designed to redact secrets; prefer sharing that redacted output when asking for help.

## Symptom table

| Symptom                                                                       | Likely cause                                                                                               | Fix                                                                                                                                                                                                                                                                           |
| ----------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| macOS says the app is from an unidentified developer                          | Gatekeeper could not validate the app signature or notarization.                                           | Verify the release with the commands in [Install on macOS]({{< relref "../install/macos#verify-gatekeeper-and-notarization" >}}). Do not override Gatekeeper until the signature checks pass.                                                                                 |
| MCP client reports `missing intervals.icu API key`                            | The API key is not in the OS keychain, `INTERVALS_ICU_API_KEY`, or legacy config fallback.                 | Run `icuvisor setup`, or follow [API key setup]({{< relref "api-key" >}}). Confirm the macOS Keychain item with `security find-generic-password -s icuvisor -a intervals-icu-api-key >/dev/null`.                                                                             |
| MCP client reports `missing athlete ID` or `invalid athlete ID`               | No non-secret athlete selector was provided, or it is malformed.                                           | Set `INTERVALS_ICU_ATHLETE_ID` to `12345` or `i12345`, or provide `athlete_id` in the JSON config file. See the [config file reference]({{< relref "../reference/config-file" >}}).                                                                                           |
| The assistant says a newly added or fixed tool is still missing after upgrade | The MCP client cached the old tool catalog for the current conversation.                                   | Start a new conversation or reconnect the MCP client. If `_meta.schema_changed` appears, follow [After upgrading]({{< relref "after-upgrade" >}}).                                                                                                                            |
| HTTP client cannot connect from another machine on the LAN                    | icuvisor binds to `127.0.0.1:8765` by default, which is reachable only from the same machine.              | Read [HTTP transport]({{< relref "http-transport" >}}). Only bind a LAN IP if you accept that anyone who can reach it can call the unauthenticated MCP server.                                                                                                                |
| HTTP startup fails with an invalid bind address                               | The bind address is not an explicit IP address and port.                                                   | Use a value like `127.0.0.1:8765` or `192.168.1.10:8765`; hostnames are not accepted.                                                                                                                                                                                         |
| Linux reports key not found or cannot reach the keychain                      | libsecret/Secret Service is missing, locked, or unavailable in a headless session.                         | Install/start your desktop secret-service provider, or use `secret-tool store --label='icuvisor intervals.icu API key' service icuvisor username intervals-icu-api-key`. For headless fallback, set `INTERVALS_ICU_API_KEY` deliberately and protect the service environment. |
| Claude Desktop or Claude Code cannot start icuvisor                           | The command path is wrong or JSON syntax is invalid.                                                       | Confirm `/Applications/icuvisor.app/Contents/MacOS/icuvisor version` works. For Claude Desktop, run `plutil -lint "$HOME/Library/Application Support/Claude/claude_desktop_config.json"`. For Claude Code, validate `.mcp.json` and restart the session.                      |
| A delete tool is missing                                                      | `ICUVISOR_DELETE_MODE` is not `full`, the toolset/coach ACL hides it, or the client cached an old catalog. | Check [Safety modes and toolset tiers]({{< relref "../reference/safety-modes" >}}), restart icuvisor after changing env vars, and start a new client conversation.                                                                                                            |
| An activity response says `strava_imported: true` with `unavailable.reason` set to `strava_tos` or `strava_blocked` | intervals.icu can see a historical Strava-imported shell, but Strava's restricted API does not expose the underlying activity data. | Open the intervals.icu Connections page for the activity's original device provider and click Download old data so historical activities are re-imported directly from that provider instead of through Strava's restricted API. When icuvisor can identify the provider, it names it, for example: Open the intervals.icu Connections page, choose Wahoo, and click Download old data so historical activities are re-imported directly from Wahoo instead of through Strava's restricted API. |
