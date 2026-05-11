# Claude Desktop manual setup (macOS, v0.1)

This guide documents the v0.1 manual path for running icuvisor as a local MCP stdio server in Claude Desktop on macOS. v0.1 does not include an installer, keychain storage, or one-click client configuration yet.

Use placeholders in examples and replace them only on your own machine. Do not commit API keys, real athlete IDs, or local config files that contain secrets.

## Prerequisites

- macOS with Claude Desktop installed.
- Go 1.23 or newer.
- An intervals.icu account.
- A local clone of this repository.

## Build the local binary

From the repository root:

```bash
make build
./bin/icuvisor version
```

`make build` writes the v0.1 binary to `./bin/icuvisor`. For Claude Desktop, use an absolute path to that binary in `claude_desktop_config.json`; do not rely on shell aliases or a relative path.

If you prefer a user-local install, you can copy the built binary to a stable path such as:

```bash
mkdir -p "$HOME/bin"
cp ./bin/icuvisor "$HOME/bin/icuvisor"
"$HOME/bin/icuvisor" version
```

## Get an intervals.icu API key

1. Sign in to intervals.icu in your browser.
2. Open <https://intervals.icu/settings>.
3. Find the API key section and create or copy your personal API key.
4. Keep the key private. Treat it like a password: do not paste it into chat messages, screenshots, issues, commits, or shared config examples.

icuvisor v0.1 uses Basic Auth to call intervals.icu with the API key supplied by local config. The key is never a tool argument and should never be controlled by the LLM.

## v0.1 configuration inputs

icuvisor v0.1 can read configuration from a JSON file, a local untracked `.env` file, and process environment variables. Process environment variables win over `.env`, and `--config` or `ICUVISOR_CONFIG` chooses a JSON file to load before env overrides.

Required inputs:

| Purpose | Environment variable | JSON field | Example placeholder |
| --- | --- | --- | --- |
| intervals.icu API key | `INTERVALS_ICU_API_KEY` | `api_key` | `YOUR_INTERVALS_ICU_API_KEY` |
| intervals.icu athlete ID | `INTERVALS_ICU_ATHLETE_ID` | `athlete_id` | `i12345` |

Optional inputs:

| Purpose | Environment variable | JSON field | Default | Example placeholder |
| --- | --- | --- | --- | --- |
| Athlete timezone | `ICUVISOR_TIMEZONE` | `timezone` | `UTC` | `America/Sao_Paulo` |
| API base URL | `ICUVISOR_API_BASE_URL` | `api_base_url` | `https://intervals.icu/api/v1` | `https://intervals.icu/api/v1` |
| HTTP timeout | `ICUVISOR_HTTP_TIMEOUT` | `http_timeout` | `30s` | `30s` |
| Config path | `ICUVISOR_CONFIG` | n/a | unset | `/Users/YOU/.config/icuvisor/icuvisor.json` |

Athlete IDs may be written as `12345` or `i12345`; icuvisor normalizes them to the `i12345` form in responses. Timezones must be IANA timezone names such as `UTC`, `America/Sao_Paulo`, or `Europe/London`.

A JSON config file can look like this:

```json
{
  "api_key": "YOUR_INTERVALS_ICU_API_KEY",
  "athlete_id": "i12345",
  "timezone": "America/Sao_Paulo",
  "api_base_url": "https://intervals.icu/api/v1",
  "http_timeout": "30s"
}
```

Store this file outside the repository and restrict its permissions:

```bash
mkdir -p "$HOME/.config/icuvisor"
chmod 700 "$HOME/.config/icuvisor"
# Create the file with your editor, then:
chmod 600 "$HOME/.config/icuvisor/icuvisor.json"
```

## Optional local `.env` flow for maintainers

The config loader reads recognized keys from `.env` in the current working directory. This is useful for local maintainer smoke testing, but `.env` must remain untracked and must not be displayed in logs, issues, or screenshots.

Create a local `.env` only on your machine:

```bash
cat > .env <<'ENV'
INTERVALS_ICU_API_KEY=YOUR_INTERVALS_ICU_API_KEY
INTERVALS_ICU_ATHLETE_ID=i12345
ICUVISOR_TIMEZONE=America/Sao_Paulo
ENV
chmod 600 .env
```

Before committing, verify the file is ignored or untracked and do not print its values:

```bash
git status --short .env
```

For Claude Desktop, prefer the explicit `env` block or a JSON file referenced by `--config`, because Claude starts icuvisor with its own process environment and working directory.

## Configure Claude Desktop on macOS

Claude Desktop reads MCP server definitions from:

```text
~/Library/Application Support/Claude/claude_desktop_config.json
```

Create the file if it does not exist. The top-level shape is an `mcpServers` object. Use an absolute path for `command`.

### Option A: pass config with environment variables

```json
{
  "mcpServers": {
    "icuvisor": {
      "command": "/absolute/path/to/icuvisor",
      "env": {
        "INTERVALS_ICU_API_KEY": "YOUR_INTERVALS_ICU_API_KEY",
        "INTERVALS_ICU_ATHLETE_ID": "i12345",
        "ICUVISOR_TIMEZONE": "America/Sao_Paulo"
      }
    }
  }
}
```

### Option B: pass a JSON config file

```json
{
  "mcpServers": {
    "icuvisor": {
      "command": "/absolute/path/to/icuvisor",
      "args": [
        "--config",
        "/Users/YOU/.config/icuvisor/icuvisor.json"
      ]
    }
  }
}
```

After editing the config file, fully quit and reopen Claude Desktop so it starts the local MCP server again.

## Start a new chat after binary or tool changes

MCP clients, including Claude Desktop, can cache the tool catalog and JSON schemas for a conversation. If you rebuild icuvisor, change a tool, or edit configuration, restart Claude Desktop and start a new chat before testing. An existing conversation may continue using the old schema and can make it look like a fix did not land.

## Troubleshooting

### Claude Desktop cannot start icuvisor

- Confirm `command` is an absolute path to an executable file.
- Run the same binary from Terminal with `version`:
  ```bash
  /absolute/path/to/icuvisor version
  ```
- If you copied the binary, confirm macOS permissions allow execution:
  ```bash
  chmod +x /absolute/path/to/icuvisor
  ```

### `missing intervals.icu API key`

Set `INTERVALS_ICU_API_KEY` in the Claude Desktop `env` block, or put `api_key` in the JSON file passed with `--config`. Do not pass the API key as a chat message or tool argument.

### `missing athlete ID` or `invalid athlete ID`

Set `INTERVALS_ICU_ATHLETE_ID` or `athlete_id` to the intervals.icu athlete identifier. Both `12345` and `i12345` are accepted; other characters are rejected.

### `invalid timezone`

Use an IANA timezone name such as `UTC`, `America/Sao_Paulo`, or `Europe/London`.

### Authentication or profile fetch fails

Check that the API key is current, belongs to the expected intervals.icu account, and can access the configured athlete ID. The tool returns a short message such as `could not fetch athlete profile; check intervals.icu credentials and athlete ID` while internal details stay out of the LLM-visible response.

### Tool is missing or arguments look stale

Fully quit and reopen Claude Desktop, then start a new chat. The previous conversation may have cached an older MCP schema.

### JSON config is rejected

Check for trailing commas, comments, misspelled fields, or unknown fields. v0.1 accepts `api_key`, `athlete_id`, `timezone`, `api_base_url`, and `http_timeout` in JSON config.
