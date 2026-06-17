---
title: "Connect Claude Desktop"
description: "Install icuvisor as a Claude Desktop Extension or configure the MCP stdio server manually."
---

The recommended Claude Desktop setup is the signed icuvisor Desktop Extension (`.mcpb`). It installs the local MCP server without editing `claude_desktop_config.json` and lets Claude Desktop store the intervals.icu API key as a sensitive extension setting.

The manual JSON/keychain setup remains available as a fallback for users who cannot install the extension.

## Option 1: Install the Desktop Extension

You need:

- Claude Desktop on macOS.
- The current macOS universal `icuvisor_<version>_darwin_universal.mcpb` artifact from the GitHub release.
- Your intervals.icu athlete ID, written as `i12345` or `12345`.
- Your intervals.icu API key.

Install:

1. Download `icuvisor_<version>_darwin_universal.mcpb` from the release assets.
2. Open the `.mcpb` file with Claude Desktop, or drag it into Claude Desktop's Extensions UI.
3. When Claude Desktop asks for extension configuration, enter:
   - `api_key`: your intervals.icu API key. The manifest marks this field as sensitive, so Claude Desktop stores it in its secure extension configuration instead of writing it to `claude_desktop_config.json`.
   - `athlete_id`: your intervals.icu athlete ID, such as `i12345`.
   - `timezone`: an IANA timezone such as `UTC`, `America/Sao_Paulo`, or `Europe/London`.
   - `toolset`: keep `core` unless you intentionally want the larger `full` catalog.
4. Enable the extension and restart Claude Desktop if it asks you to.
5. Start a new chat so Claude refreshes the MCP tool catalog.

Do not paste the API key into chat messages or into manual JSON config. In the extension path, Claude Desktop passes it to icuvisor as `INTERVALS_ICU_API_KEY` only when launching the local stdio server.

After the connection works, you can add reusable [Claude Project instructions]({{< relref "../guides/claude-project-instructions" >}}) so new training chats consistently use athlete-local dates, cite icuvisor tools, and flag missing or stale data without storing secrets in the Project. If your training data starts on a Garmin or another device provider, the [Garmin to Claude walkthrough]({{< relref "../tutorials/garmin-to-claude" >}}) shows the full device-provider → intervals.icu → icuvisor → Claude path.

<details>
<summary>Extension smoke checklist</summary>

1. Open Claude Desktop after installing or enabling the extension.
2. Start a new chat.
3. Ask: `What's my FTP?`
4. Expected result: Claude calls icuvisor, uses [`get_athlete_profile`]({{< relref "../reference/tools#get_athlete_profile" >}}), and answers with your configured FTP/threshold data from intervals.icu.

If the extension installs but tools do not appear, open Claude Desktop settings, confirm the icuvisor extension is enabled, and verify the athlete ID/timezone fields.

</details>

## Option 2: Manual JSON and keychain fallback

Use this path if you installed `icuvisor.app` from the macOS DMG, installed `icuvisor.exe` on Windows, use a development binary, or cannot install `.mcpb` files in your Claude Desktop environment.

You need:

- macOS: `icuvisor.app` installed in `/Applications`, or another absolute path to an `icuvisor` binary.
- Windows: `icuvisor.exe` installed at `%LOCALAPPDATA%\Programs\icuvisor\icuvisor.exe`, or another absolute path to `icuvisor.exe`.
- Your intervals.icu athlete ID, written as `i12345` or `12345`.
- Your API key stored by `icuvisor setup` in the OS credential store. On macOS this is Keychain; on Windows this is Windows Credential Manager.

Store the API key first:

```bash
/Applications/icuvisor.app/Contents/MacOS/icuvisor setup
```

On Windows, run this in a new PowerShell window after installing:

```powershell
icuvisor setup
```

Claude Desktop reads MCP server definitions from:

```text
~/Library/Application Support/Claude/claude_desktop_config.json
```

On Windows, the file is:

```text
%APPDATA%\Claude\claude_desktop_config.json
```

That usually expands to `C:\Users\<you>\AppData\Roaming\Claude\claude_desktop_config.json`.

Create the file if it does not exist. Add or merge the right `mcpServers.icuvisor` block for your OS, replacing only the non-secret placeholders.

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

Notes:

- Do not put your intervals.icu API key in `claude_desktop_config.json`; the manual path reads it from the OS credential store.
- `ICUVISOR_TRANSPORT=stdio` is optional because stdio is the default, but keeping it explicit makes the config easier to audit.
- Use a real IANA timezone such as `UTC`, `America/Sao_Paulo`, or `Europe/London`.
- If you installed the app somewhere else, update `command` to the absolute path to `icuvisor.app/Contents/MacOS/icuvisor` or `icuvisor.exe`.

After editing the file, fully quit and reopen Claude Desktop.

<details>
<summary>Manual fallback smoke checklist</summary>

1. Open Claude Desktop after restarting it.
2. Start a new chat so Claude refreshes the MCP tool catalog.
3. Ask: `What's my FTP?`
4. Expected result: Claude calls icuvisor, uses [`get_athlete_profile`]({{< relref "../reference/tools#get_athlete_profile" >}}), and answers with your configured FTP/threshold data from intervals.icu.

If the answer says the tool is missing or cannot start, run:

```bash
/Applications/icuvisor.app/Contents/MacOS/icuvisor version
plutil -lint "$HOME/Library/Application Support/Claude/claude_desktop_config.json"
```

On Windows:

```powershell
& "$env:LOCALAPPDATA\Programs\icuvisor\icuvisor.exe" version
Get-Content "$env:APPDATA\Claude\claude_desktop_config.json" | ConvertFrom-Json | Out-Null
```

If the answer reports missing credentials, confirm the Keychain item exists and the athlete ID is set in the JSON:

```bash
security find-generic-password -s icuvisor -a intervals-icu-api-key >/dev/null
```

On Windows, confirm Credential Manager has an icuvisor entry:

```powershell
cmdkey /list | findstr /i icuvisor
```

</details>

## Updating

For the extension path, download and open the newer `.mcpb` release asset, then restart Claude Desktop and start a new chat.

For the manual app path:

- macOS: download the newer signed DMG, replace `/Applications/icuvisor.app`, fully quit Claude Desktop, and start a new chat.
- Windows: rerun `iwr -useb https://icuvisor.app/install.ps1 | iex` in PowerShell, or install the newer MSI, fully quit Claude Desktop, and start a new chat.

In both paths, keep the API key out of manual JSON config. Also start a new chat after changing Project instructions. If Claude still shows old tools or stale answers, use the [stale conversation troubleshooting guide]({{< relref "../guides/troubleshooting#stale-conversations-and-cached-tool-catalogs" >}}).
