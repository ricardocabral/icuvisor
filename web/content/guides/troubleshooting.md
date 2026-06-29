---
title: "Troubleshooting"
description: "Common icuvisor setup symptoms and fixes."
---

Use this guide when icuvisor does not start, an MCP client cannot see tools, or a connection works differently after an upgrade.

## Data access and tool-selection confusion

These symptoms often appear after a first successful connection, when the assistant can call icuvisor but the answer is incomplete, stale, or based on the wrong capability. For a single bounded check across activities, streams, load, wellness, thresholds, and calendar data, ask for the [data quality report]({{< relref "../cookbook/data-quality-report" >}}).

### Strava-restricted activities

If an activity response says `strava_imported: true`, `unavailable.reason: strava_tos`, `unavailable.reason: strava_blocked`, or `_meta.data_availability[].reason: restricted_source`, icuvisor is reporting an upstream access restriction. intervals.icu may show an activity shell or summary, but the API response available to icuvisor does not include the complete streams, intervals, or samples.

What to do next:

1. Do not paste private activity files, API keys, or screenshots into chat as a workaround.
2. Open intervals.icu Connections and connect the original provider when possible, such as Garmin, Wahoo, Zwift, or another device source.
3. Use that provider's **Download old data** or equivalent historical import option so intervals.icu re-imports the activity directly instead of relying on the Strava-restricted copy.
4. Start a new assistant conversation and ask icuvisor to re-check the affected activity. If the API still omits the data, ask for an answer based only on available summary fields.

icuvisor cannot bypass Strava's restricted API access and should not guess the missing values.

### Missing streams or partial activity detail

If `_meta.data_availability[].reason` is `missing_stream` for `heart_rate`, `watts`, pace, distance, cadence, or another channel, the activity exists but that sample stream is not available in the API payload. The original recording might not contain that sensor channel, or source restrictions might hide stream samples even when intervals.icu shows a high-level summary.

What to do next:

1. Ask the assistant to list which fields are present before it analyzes the activity.
2. Use summary fields such as duration, distance, training load, average heart rate, or max heart rate when they are returned.
3. Re-import from the native provider if the activity came through Strava and you need stream-derived analysis.
4. If the stream was never recorded, use a different activity or ask for a summary-level answer instead of interval-by-interval analysis.

### HR-only, TRIMP, or missing load fields

Fitness and training-summary responses can include `_meta.load_diagnostics` values such as `trimp_or_hr_load_available`, `missing_training_load`, or `fitness_fields_missing`. That means the account or date range may rely on heart-rate/TRIMP-style load, or intervals.icu omitted specific fitness fields for those dates. icuvisor preserves the upstream load value as `training_load`; it does not rename TRIMP or heart-rate load to `tss`, and it does not report missing CTL/ATL/TSB as zero.

What to do next:

1. Ask the assistant to use `training_load` as the neutral intervals.icu load value unless you specifically know the account's load model.
2. Check athlete sport settings and recent activities for power, pace, threshold, and heart-rate availability before requesting TSS- or threshold-specific conclusions.
3. Narrow the date range to days with known completed activities if broad summaries come back sparse.
4. If a field is still absent, ask for a plain-language caveat instead of a substituted metric.

### Stale or missing activities

If the assistant cannot find a recent activity, treats a planned workout as completed, or reports old zones/timezones after you fixed them, the issue is usually one of three things: the activity has not synced to intervals.icu yet, the date window is wrong for your athlete timezone, or the MCP client is reusing stale conversation context.

What to do next:

1. Confirm the activity appears in intervals.icu first. icuvisor can only read what upstream exposes.
2. Ask the assistant to resolve the date window in your athlete timezone and search a slightly wider window.
3. Start a new chat or reconnect/reload MCP tools so the client refreshes cached schemas and context.
4. Ask for `workout_status` fields when comparing planned versus completed work, so pending future workouts are not counted as missed.
5. Run `icuvisor diagnostics` locally if the configured athlete ID, timezone, or credential source might be wrong; share only the redacted output when asking for help.

### The assistant chose the wrong tool or did not call icuvisor

Some AI clients answer from memory, use a stale tool list, or choose a generic profile check when you asked for a richer training review. That is a client/tool-selection problem, not a reason to paste credentials into chat.

What to do next:

1. Restate the request with an explicit connector instruction: "Use icuvisor and my intervals.icu data before answering."
2. Add the outcome and window: "last 14 days", "today", "next 7 days", or exact dates.
3. Ask it to cite the icuvisor source tool behind key numbers.
4. If the client still misses the right capability, ask it to call `icuvisor_list_advanced_capabilities` or use the beginner prompts in [What can I ask icuvisor?]({{< relref "../cookbook/what-can-i-ask" >}}).
5. Start a new conversation after changing `ICUVISOR_TOOLSET`, delete/write mode, hosted preferences, or coach athlete selection.

## Stale conversations and cached tool catalogs

MCP clients can keep a copy of icuvisor's tool catalog and sometimes keep using context from the current chat. That is useful for speed, but it means a long-running conversation may not immediately notice that you upgraded icuvisor, changed `ICUVISOR_TOOLSET` or `ICUVISOR_DELETE_MODE`, edited your timezone, fixed zone settings in intervals.icu, or restarted the server with a different config. The binary can be correct while the open chat is still reasoning from yesterday's tools or assumptions.

Before changing credentials or pasting details into a chat, try these safe first steps:

1. If `_meta.schema_changed: true` is visible in a tool response, start a new conversation in the MCP client. This is the quickest way to force the assistant to stop using stale chat context.
2. Refresh, reconnect, or reload MCP tools in the client, then fully restart the client if tools still look wrong. Different clients use different labels for this action.
3. If your client hides `_meta`, ask the assistant to call `icuvisor_check_server_version` with no arguments after reconnecting. Compare the visible tool-description fields `description_server_version`, `description_catalog_fingerprint`, `description_toolset`, and `description_delete_mode` with response `server_version`, `description_catalog_fingerprint`, `toolset`, and `delete_mode`. If any pair differs, reload/reconnect the MCP client again; if the current chat keeps using stale schemas or context after reconnecting, start a new conversation.
4. Verify the running binary with `icuvisor version` and confirm your client config points at that binary.
5. Run `icuvisor diagnostics` and review the redacted output for the active transport, toolset, delete mode, timezone, config source, and credential source.

Do not compare the diagnostic `description_catalog_fingerprint` with response `catalog_hash`; `catalog_hash` is the live exposed-catalog hash, while the description fingerprint is a visible-description fallback for clients that hide `_meta`.

Common stale-state symptoms include:

- The assistant reports the wrong local date, day boundary, or timezone after you changed `ICUVISOR_TIMEZONE`.
- FTP, heart-rate, pace, or power zones look old after you updated settings in intervals.icu.
- A tool added in a new release is missing, an old tool schema is still used, or a delete/write tool remains hidden after changing toolset or delete-mode settings.
- Writes fail with arguments that no longer match the current docs, or the assistant keeps retrying a field that a new schema removed or renamed.

Treat API keys as credentials, not troubleshooting data. Do not paste an intervals.icu API key into an assistant conversation, issue report, or manual MCP JSON config. Use `icuvisor setup`, the OS keychain, or your client's sensitive extension setting instead. `icuvisor diagnostics` is designed to redact secrets; prefer sharing that redacted output when asking for help. The `icuvisor_check_server_version` tool is also local and read-only: it has no arguments, does not call intervals.icu, and returns only server version, catalog fingerprint/hash, toolset, and delete-mode metadata. For the trust-boundary model behind this guidance, read [Local-first design]({{< relref "../explain/local-first" >}}).

## Symptom table

| Symptom                                                                       | Likely cause                                                                                               | Fix                                                                                                                                                                                                                                                                           |
| ----------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| macOS says the app is from an unidentified developer                          | Gatekeeper could not validate the app signature or notarization.                                           | Verify the release with the commands in [Install on macOS]({{< relref "../install/macos#verify-gatekeeper-and-notarization" >}}). Do not override Gatekeeper until the signature checks pass.                                                                                 |
| MCP client reports `missing intervals.icu API key`                            | The API key is not in the OS keychain, Windows Credential Manager, `INTERVALS_ICU_API_KEY`, or legacy config fallback. | Run `icuvisor setup`, or follow [API key setup]({{< relref "api-key" >}}). On macOS, confirm the Keychain item with `security find-generic-password -s icuvisor -a intervals-icu-api-key >/dev/null`. On Windows, confirm the Credential Manager target with `cmdkey /list \| findstr /i icuvisor`. |
| MCP client reports `missing athlete ID` or `invalid athlete ID`               | No non-secret athlete selector was provided, or it is malformed.                                           | Set `INTERVALS_ICU_ATHLETE_ID` to `12345` or `i12345`, or provide `athlete_id` in the JSON config file. See the [config file reference]({{< relref "../reference/config-file" >}}).                                                                                           |
| The assistant says a newly added or fixed tool is still missing after upgrade | The MCP client cached the old tool catalog for the current conversation.                                   | Reconnect/reload MCP tools and call `icuvisor_check_server_version` if `_meta` is hidden. Start a new conversation if `_meta.schema_changed` appears or stale schemas persist. Follow [After upgrading]({{< relref "after-upgrade" >}}).                                                                                                  |
| HTTP client cannot connect from another machine on the LAN                    | icuvisor binds to `127.0.0.1:8765` by default, which is reachable only from the same machine.              | Read [HTTP transport]({{< relref "http-transport" >}}). Only bind a LAN IP if you accept that anyone who can reach it can call the unauthenticated MCP server.                                                                                                                |
| HTTP startup fails with an invalid bind address                               | The bind address is not an explicit IP address and port.                                                   | Use a value like `127.0.0.1:8765` or `192.168.1.10:8765`; hostnames are not accepted.                                                                                                                                                                                         |
| Linux reports key not found or cannot reach the keychain                      | libsecret/Secret Service is missing, locked, or unavailable in a headless session.                         | Install/start your desktop secret-service provider, or use `secret-tool store --label='icuvisor intervals.icu API key' service icuvisor username intervals-icu-api-key`. For headless fallback, set `INTERVALS_ICU_API_KEY` deliberately and protect the service environment. |
| Claude Desktop or Claude Code cannot start icuvisor                           | The command path is wrong or JSON syntax is invalid.                                                       | Confirm the binary path works: macOS `/Applications/icuvisor.app/Contents/MacOS/icuvisor version`; Windows `& "$env:LOCALAPPDATA\Programs\icuvisor\icuvisor.exe" version`. For Claude Desktop, lint macOS JSON with `plutil -lint "$HOME/Library/Application Support/Claude/claude_desktop_config.json"` or Windows JSON with `Get-Content "$env:APPDATA\Claude\claude_desktop_config.json" \| ConvertFrom-Json \| Out-Null`. For Claude Code, validate `.mcp.json` and restart the session. |
| A delete tool is missing                                                      | `ICUVISOR_DELETE_MODE` is not `full`, the toolset/coach ACL hides it, or the client cached an old catalog. | Check [Safety modes and toolset tiers]({{< relref "../reference/safety-modes" >}}), restart icuvisor after changing env vars, and start a new client conversation.                                                                                                            |
| An activity response says `strava_imported: true` with `unavailable.reason` set to `strava_tos` or `strava_blocked`, or `_meta.data_availability[].reason` is `restricted_source` | intervals.icu can see a historical Strava-imported shell, but Strava's restricted API does not expose the underlying activity data, streams, intervals, or complete max-heart-rate samples through the API. | Open the intervals.icu Connections page for the activity's original device provider and click Download old data so historical activities are re-imported directly from that provider instead of through Strava's restricted API. When icuvisor can identify the provider, it names it, for example: Open the intervals.icu Connections page, choose Wahoo, and click Download old data so historical activities are re-imported directly from Wahoo instead of through Strava's restricted API. For Garmin, Zwift, Wahoo, and similar providers, prefer a direct provider connection/import over a Strava-only history when you need streams. |
| A stream response has `_meta.data_availability[].reason: "missing_stream"` for `heart_rate`, `watts`, pace, or distance | The activity exists, but that stream channel is absent from the API response. This can happen when the original file lacked the sensor data or when source restrictions hide stream samples even though the Intervals.icu UI shows an activity summary. | Use summary fields such as `max_heart_rate_bpm` when present. If the activity came through Strava and you need stream-derived answers, re-import historical data directly from the native provider through intervals.icu Connections so icuvisor can read the samples through the API. |
| Fitness or training-summary output has `_meta.load_diagnostics` with `trimp_or_hr_load_available`, `missing_training_load`, or `fitness_fields_missing` | The account may be HR/TRIMP-only or upstream omitted some fitness fields for the requested dates. icuvisor preserves explicit load as neutral `training_load`; it does not call TRIMP or HR load `tss`. Missing CTL/ATL/TSB fields are omitted instead of reported as zero. | Treat `training_load` as the Intervals.icu load value for the athlete's configured model. For TSS/power/threshold-pace analysis, first check athlete sport settings, power/pace availability, and whether the requested dates have upstream fitness rows. |
