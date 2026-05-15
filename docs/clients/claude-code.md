# Claude Code manual setup (macOS)

Use this guide after installing `icuvisor.app` from the signed macOS DMG. Claude Code starts icuvisor over MCP stdio by executing the binary inside the app bundle.

## Prerequisites

- macOS with Claude Code installed.
- `icuvisor.app` installed in `/Applications`.
- Your intervals.icu athlete ID, written as `i12345` or `12345`.
- Your intervals.icu API key stored in the macOS Keychain under service `icuvisor` and account `intervals-icu-api-key`.

> **Do not put your intervals.icu API key in `.mcp.json` or any Claude Code project config.** icuvisor reads it from the macOS Keychain. The MCP config should contain only non-secret values.

If needed, add the key to Keychain from Terminal:

```bash
security add-generic-password -U \
  -s icuvisor \
  -a intervals-icu-api-key \
  -w 'YOUR_INTERVALS_ICU_API_KEY'
```

## Project `.mcp.json` configuration

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

Restart Claude Code or reload MCP servers after editing the file.

## Verify the connection

1. Open Claude Code in the project directory containing `.mcp.json`.
2. Start a new session so the MCP catalog is refreshed.
3. Ask: `What's my FTP?`
4. Expected result: Claude Code calls icuvisor through MCP stdio and answers with FTP/threshold data from intervals.icu.

Quick local checks:

```bash
/Applications/icuvisor.app/Contents/MacOS/icuvisor version
security find-generic-password -s icuvisor -a intervals-icu-api-key >/dev/null
```

If Claude Code cannot see the tool, verify the JSON syntax, restart the session, and confirm the binary path is absolute and executable.
