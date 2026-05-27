---
title: "Connect Claude Code"
description: "Configure Claude Code to start icuvisor over MCP stdio on macOS."
---

Use this guide after installing `icuvisor.app` from the macOS DMG and storing your intervals.icu API key with `icuvisor setup`.

## Before you start

You need:

- Claude Code installed on macOS.
- `icuvisor.app` installed in `/Applications`.
- Your intervals.icu athlete ID, written as `i12345` or `12345`.
- Your API key stored in the macOS Keychain under service `icuvisor` and account `intervals-icu-api-key`.

Do not put your intervals.icu API key in `.mcp.json` or any Claude Code project config. The MCP config should contain only non-secret values. For reusable chat behavior, use [Claude Project instructions]({{< relref "../guides/claude-project-instructions" >}}) that enforce timezone/date discipline and data grounding without storing secrets.

If needed, run setup first:

```bash
/Applications/icuvisor.app/Contents/MacOS/icuvisor setup
```

## Add project `.mcp.json`

From the project directory where you run Claude Code, create or edit `.mcp.json`:

```json
{
  "mcpServers": {
    "icuvisor": {
      "command": "/Applications/icuvisor.app/Contents/MacOS/icuvisor",
      "env": {
        "INTERVALS_ICU_ATHLETE_ID": "i12345",
        "ICUVISOR_TIMEZONE": "America/Sao_Paulo",
        "ICUVISOR_TRANSPORT": "stdio"
      }
    }
  }
}
```

Notes:

- `ICUVISOR_TRANSPORT=stdio` is optional because stdio is the default.
- Keep `.mcp.json` out of commits if it contains personal athlete IDs or local-only paths.
- If your team commits a shared `.mcp.json`, use placeholders and document that each user must add their own non-secret athlete ID locally.
- If you installed the app somewhere else, update `command` to the absolute path to `icuvisor.app/Contents/MacOS/icuvisor`.

Restart Claude Code or reload MCP servers after editing the file. Start a new session after changing Project instructions or MCP config. If tools or answers still look stale, follow the [stale conversation troubleshooting guide]({{< relref "../guides/troubleshooting#stale-conversations-and-cached-tool-catalogs" >}}).

## Verify the connection

1. Open Claude Code in the project directory containing `.mcp.json`.
2. Start a new session so the MCP catalog is refreshed.
3. Ask: `What's my FTP?`
4. Expected result: Claude Code calls icuvisor through MCP stdio and uses [`get_athlete_profile`]({{< relref "../reference/tools#get_athlete_profile" >}}) to answer with FTP/threshold data from intervals.icu.

Quick local checks:

```bash
/Applications/icuvisor.app/Contents/MacOS/icuvisor version
security find-generic-password -s icuvisor -a intervals-icu-api-key >/dev/null
```

If Claude Code cannot see icuvisor, verify the JSON syntax, restart the session, and confirm the binary path is absolute and executable.
