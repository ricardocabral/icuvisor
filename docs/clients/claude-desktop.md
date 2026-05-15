# Claude Desktop manual setup (macOS)

Use this guide after installing `icuvisor.app` from the signed macOS DMG. Claude Desktop starts icuvisor over MCP stdio by executing the binary inside the app bundle.

## Prerequisites

- macOS with Claude Desktop installed.
- `icuvisor.app` installed in `/Applications`.
- Your intervals.icu athlete ID, written as `i12345` or `12345`.
- Your intervals.icu API key stored in the macOS Keychain under service `icuvisor` and account `intervals-icu-api-key`.

> **Do not put your intervals.icu API key in `claude_desktop_config.json`.** icuvisor reads it from the macOS Keychain. The Claude Desktop JSON should contain only non-secret configuration such as athlete ID, timezone, and transport.

If the key is not already in Keychain, add it from Terminal:

```bash
security add-generic-password -U \
  -s icuvisor \
  -a intervals-icu-api-key \
  -w 'YOUR_INTERVALS_ICU_API_KEY'
```

## Configure Claude Desktop

Claude Desktop reads MCP server definitions from:

```text
~/Library/Application Support/Claude/claude_desktop_config.json
```

Create the file if it does not exist. Add or merge this `mcpServers.icuvisor` block, replacing only the non-secret placeholders:

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

- `ICUVISOR_TRANSPORT=stdio` is optional because stdio is the default, but keeping it explicit makes the MCP client setup easier to audit.
- Use a real IANA timezone such as `UTC`, `America/Sao_Paulo`, or `Europe/London`.
- If you installed the app somewhere else, update `command` to the absolute path to `icuvisor.app/Contents/MacOS/icuvisor`.

After editing the config file, fully quit and reopen Claude Desktop.

## Verify the connection

1. Open Claude Desktop after restarting it.
2. Start a new chat so Claude refreshes the MCP tool catalog.
3. Ask: `What's my FTP?`
4. Expected result: Claude calls icuvisor, reads your athlete profile, and answers with the configured FTP/threshold data from intervals.icu.

If the answer says the tool is missing or cannot start:

```bash
/Applications/icuvisor.app/Contents/MacOS/icuvisor version
plutil -lint "$HOME/Library/Application Support/Claude/claude_desktop_config.json"
```

If the answer reports missing credentials, confirm the Keychain item exists and the athlete ID is set in the JSON:

```bash
security find-generic-password -s icuvisor -a intervals-icu-api-key >/dev/null
```

## Updating the app

Download the newer signed DMG from the GitHub release, replace `/Applications/icuvisor.app`, fully quit Claude Desktop, and start a new chat. Do not move the API key into the JSON during upgrades; it remains in Keychain.
