---
title: "Connect Codex CLI"
description: "Configure Codex CLI to start icuvisor over MCP stdio."
---

Use this guide after installing `icuvisor` and storing your intervals.icu API key with `icuvisor setup`.

Codex supports MCP servers in the CLI and stores them in `config.toml`. The same config is shared with the Codex IDE extension. See OpenAI's [Codex MCP docs](https://developers.openai.com/codex/mcp) for the full command and config reference.

## Before you start

You need:

- Codex CLI installed and authenticated.
- `icuvisor` installed locally.
- Your intervals.icu athlete ID, written as `i12345` or `12345`.
- Your API key stored by `icuvisor setup`.

Do not put your intervals.icu API key in Codex config or prompts. The Codex MCP entry should contain only non-secret values such as athlete ID, timezone, and transport.

If needed, run setup first:

```bash
icuvisor setup
```

Find the absolute path to the binary:

```bash
command -v icuvisor
icuvisor version
```

## Add icuvisor with Codex CLI

Replace the path, athlete ID, and timezone:

```bash
codex mcp add icuvisor \
  --env INTERVALS_ICU_ATHLETE_ID=i12345 \
  --env ICUVISOR_TIMEZONE=America/Sao_Paulo \
  --env ICUVISOR_TRANSPORT=stdio \
  -- /absolute/path/to/icuvisor
```

For the macOS app install, the command path is usually:

```text
/Applications/icuvisor.app/Contents/MacOS/icuvisor
```

For the Windows installer, the command path is usually:

```text
C:\Users\<you>\AppData\Local\Programs\icuvisor\icuvisor.exe
```

Codex writes the server entry to `~/.codex/config.toml` by default. OpenAI also documents project-scoped `.codex/config.toml` for trusted projects.

## Manual config.toml option

If you prefer to edit config directly, add this to `~/.codex/config.toml`:

```toml
[mcp_servers.icuvisor]
command = "/absolute/path/to/icuvisor"

[mcp_servers.icuvisor.env]
INTERVALS_ICU_ATHLETE_ID = "i12345"
ICUVISOR_TIMEZONE = "America/Sao_Paulo"
ICUVISOR_TRANSPORT = "stdio"
```

`ICUVISOR_TRANSPORT=stdio` is optional because stdio is icuvisor's default transport, but keeping it explicit makes the config easier to audit.

## Verify the connection

Start a fresh Codex session and run:

```text
/mcp
```

You should see `icuvisor` listed as an active MCP server. Then ask:

```text
What's my FTP?
```

A working setup should call [`get_athlete_profile`]({{< relref "../reference/tools#get_athlete_profile" >}}) and answer from intervals.icu data.

If Codex cannot see icuvisor, run `codex mcp --help`, check the binary path, and start a new Codex session. If tools or answers look stale after an icuvisor upgrade, follow the [stale conversation troubleshooting guide]({{< relref "../guides/troubleshooting#stale-conversations-and-cached-tool-catalogs" >}}).

For maintainer-only one-off validation without persistent Codex MCP settings, see [`docs/clients/codex-local.md`](https://github.com/ricardocabral/icuvisor/blob/main/docs/clients/codex-local.md).
