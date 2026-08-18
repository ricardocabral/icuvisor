---
title: "Other MCP clients"
description: "Use the standard icuvisor MCP JSON shape with Cursor, Continue, Zed, Pi, and other clients."
---

Most MCP clients need the same three pieces of information:

1. A command that starts the local icuvisor binary.
2. Non-secret environment variables such as athlete ID, timezone, and transport.
3. A restart or new conversation so the client reloads the MCP tool catalog.

Keep the intervals.icu API key out of client JSON. Store it with `icuvisor setup` or provide it through a process environment only for deliberate headless fallback.

If the client cannot run a local process and requires a public HTTPS connector URL, use [hosted mode]({{< relref "hosted" >}}) instead.

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
        "ICUVISOR_TRANSPORT": "stdio",
        "ICUVISOR_TOOLSET": "core"
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
        "ICUVISOR_TRANSPORT": "stdio",
        "ICUVISOR_TOOLSET": "core"
      }
    }
  }
}
```

For clients that ask for only one server entry, copy the inner `icuvisor` object. `ICUVISOR_TOOLSET=core` is the default and is shown explicitly so you can change profiles deliberately.

## Choose a tool profile

`ICUVISOR_TOOLSET` controls how many tools the model sees:

| Profile | Use it for | What changes |
| --- | --- | --- |
| `core` | Claude, ChatGPT local surfaces, Cursor, Cline, Continue, and most daily training chats. | Default balanced catalog with common reads and safe write workflows. |
| `compact` | Local/Ollama/OpenRouter-style clients, smaller models, or any client that starts choosing nonexistent tools or arguments from a large catalog. | Smaller read-focused catalog for status, activities, streams, wellness, calendar reads, and safe planning context. Write/delete and specialist analyzer tools stay hidden. |
| `full` | Expert workflows on capable clients when you need the complete icuvisor surface. | Broadest catalog, including specialist analyzers and advanced management tools subject to write/delete safety settings. |

If the assistant says a specialist capability is missing, ask it to call `icuvisor_list_advanced_capabilities` first. That discovery tool is visible in `compact` and `core` and can explain whether you should stay on the current profile, move from `compact` to `core`, or opt in to `full`.

If you switch between `compact`, `core`, and `full`, restart the MCP server and open a new client conversation so the client reloads the catalog.

## Client notes

| Client                         | What to configure                                                                                                                                |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| Cursor                         | Add an MCP server named `icuvisor` with the command path and non-secret environment values above. Use `core` unless you see tool-selection confusion, then try `compact`. Restart the relevant workspace or MCP session. |
| Gemini Spark                   | See [Gemini Spark setup]({{< relref "gemini-spark" >}}). Core verification covers the MCP protocol over local Streamable HTTP only; a Gemini web/mobile surface needs a separately reachable hosted integration, which core does not provide. |
| Cline                          | Add the same stdio server object to Cline's MCP settings. Start with `core`; use `compact` for smaller/local model providers.                    |
| Continue                       | Add the server to Continue's MCP configuration using the same command/env shape. Use `compact` when pairing Continue with local/Ollama-style models that struggle with larger catalogs. Restart Continue after editing config. |
| Zed                            | Add icuvisor as a local MCP server. Use an absolute binary path and non-secret environment variables.                                            |
| Local/Ollama/OpenRouter-style clients | Prefer `ICUVISOR_TOOLSET=compact` first; move to `core` only after basic profile, activity, wellness, and calendar reads route reliably. |
| Pi or another MCP-aware client | If it supports local stdio servers, use the stdio JSON. If it requires local HTTP, use Streamable HTTP on loopback. If it requires public HTTPS, use hosted mode. |

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

After saving the configuration, start a new client conversation and ask a simple profile question such as `What's my FTP?` A working setup should call [`get_athlete_profile`]({{< relref "../reference/tools#get_athlete_profile" >}}). Then try [What can I ask icuvisor?]({{< relref "../cookbook/what-can-i-ask" >}}) for beginner prompts. If the client keeps using stale tools or assumptions after a config change, see [stale conversation troubleshooting]({{< relref "../guides/troubleshooting#stale-conversations-and-cached-tool-catalogs" >}}).
