---
title: "Other MCP clients"
description: "Use the standard icuvisor MCP JSON shape with Cursor, Continue, Zed, Pi, and other clients."
---

Most MCP clients need the same three pieces of information:

1. A command that starts the local icuvisor binary.
2. Non-secret environment variables such as athlete ID, timezone, and transport.
3. A restart or new conversation so the client reloads the MCP tool catalog.

Keep the intervals.icu API key out of client JSON. Store it with `icuvisor setup` or provide it through a process environment only for deliberate headless fallback.

## Standard stdio JSON

Use this shape for clients that accept a Claude-style `mcpServers` map, including Cursor and many local-agent clients:

macOS:

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

Windows:

```json
{
  "mcpServers": {
    "icuvisor": {
      "command": "C:\\Users\\<you>\\AppData\\Local\\Programs\\icuvisor\\icuvisor.exe",
      "env": {
        "INTERVALS_ICU_ATHLETE_ID": "i12345",
        "ICUVISOR_TIMEZONE": "Europe/Brussels",
        "ICUVISOR_TRANSPORT": "stdio"
      }
    }
  }
}
```

For clients that ask for only one server entry, copy the inner `icuvisor` object.

## Client notes

| Client                         | What to configure                                                                                                                                |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| Cursor                         | Add an MCP server named `icuvisor` with the command path and non-secret environment values above. Restart the relevant workspace or MCP session. |
| Continue                       | Add the server to Continue's MCP configuration using the same command/env shape. Restart Continue after editing config.                          |
| Zed                            | Add icuvisor as a local MCP server. Use an absolute binary path and non-secret environment variables.                                            |
| Pi or another MCP-aware client | If it supports local stdio servers, use the stdio JSON. If it requires an HTTP URL, use Streamable HTTP on loopback.                             |

## HTTP URL for clients that require it

Start icuvisor in HTTP mode:

macOS:

```bash
ICUVISOR_TRANSPORT=http /Applications/icuvisor.app/Contents/MacOS/icuvisor
```

Windows PowerShell:

```powershell
$env:ICUVISOR_TRANSPORT = "http"
& "$env:LOCALAPPDATA\Programs\icuvisor\icuvisor.exe"
```

Use this MCP endpoint:

```text
http://127.0.0.1:8765/mcp
```

Use loopback by default. A LAN bind exposes an unauthenticated MCP server to any host that can reach the address.

## Verify

After saving the configuration, start a new client conversation and ask a simple profile question such as `What's my FTP?` A working setup should call [`get_athlete_profile`]({{< relref "../reference/tools#get_athlete_profile" >}}). If the client keeps using stale tools or assumptions after a config change, see [stale conversation troubleshooting]({{< relref "../guides/troubleshooting#stale-conversations-and-cached-tool-catalogs" >}}).
