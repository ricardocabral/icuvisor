---
title: "CLI reference"
description: "icuvisor commands, flags, environment variables, and exit codes."
---

The icuvisor binary is both the MCP server and the setup/diagnostics CLI. Running `icuvisor` with no command starts the MCP server over stdio by default.

Use this page when you need the exact command-line surface. The full output below is rendered from `internal/app/testdata/help.golden`, which is the CLI golden fixture used by tests.

## Commands

| Command                                           | What it does                                                                                                                                      |
| ------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| `icuvisor`                                        | Starts the MCP server. The default transport is stdio.                                                                                            |
| `icuvisor setup`                                  | Stores the intervals.icu API key in the OS keychain and writes non-secret config. Setup has its own `--config`, `--offline`, and `--force` flags. |
| `icuvisor diagnostics`                            | Prints redacted local diagnostics and exits. It does not start the MCP server.                                                                    |
| `icuvisor version`                                | Prints the version and exits.                                                                                                                     |
| `icuvisor help`, `icuvisor --help`, `icuvisor -h` | Prints help and exits.                                                                                                                            |

## Flags

| Flag                 | Description                                                                                                                                       |
| -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--config <path>`    | JSON config file path. Equivalent environment variable: `ICUVISOR_CONFIG`. When both are omitted, icuvisor also loads the default user config (`~/Library/Application Support/icuvisor/config.json` on macOS, `$XDG_CONFIG_HOME/icuvisor/config.json` on Linux, `%AppData%\icuvisor\config.json` on Windows) if that file exists. |
| `--env-file <path>`  | Env-file path loaded before process environment. Equivalent environment variable: `ICUVISOR_ENV_FILE`. If omitted, `.env` is loaded when present. |
| `--transport <name>` | MCP transport: `stdio` or `http`. Default: `stdio`.                                                                                               |
| `--http-bind <addr>` | HTTP bind address for `--transport http`. Default: `127.0.0.1:8765`.                                                                              |
| `-h`, `--help`       | Print help and exit.                                                                                                                              |

`icuvisor setup` has a separate flag set: `--config <path>`, `--offline`, `--force`, and `--help`. There is intentionally no `--api-key` flag; setup asks for the key interactively with masked terminal input.

## Environment variables

| Variable                   | Default                        | Description                                                                                                                  |
| -------------------------- | ------------------------------ | ---------------------------------------------------------------------------------------------------------------------------- |
| `INTERVALS_ICU_API_KEY`    | none                           | intervals.icu API key. This overrides keychain lookup and legacy plaintext file keys. Prefer the OS keychain for normal use. |
| `INTERVALS_ICU_ATHLETE_ID` | none                           | Athlete ID with or without the leading `i`, for example `12345` or `i12345`.                                                 |
| `ICUVISOR_CONFIG`          | user config dir if present     | JSON config file path used when `--config` is omitted. When both are unset, icuvisor falls back to the default user config path described under `--config` above. |
| `ICUVISOR_ENV_FILE`        | `.env` when present            | Env-file path used when `--env-file` is omitted. Explicit env-file paths must exist.                                         |
| `ICUVISOR_TIMEZONE`        | `UTC`                          | Athlete timezone as an IANA name, such as `Europe/London` or `America/Sao_Paulo`.                                            |
| `ICUVISOR_API_BASE_URL`    | `https://intervals.icu/api/v1` | intervals.icu API base URL. Most users should not change this.                                                               |
| `ICUVISOR_HTTP_TIMEOUT`    | `30s`                          | HTTP client timeout as a Go duration string.                                                                                 |
| `ICUVISOR_TRANSPORT`       | `stdio`                        | MCP transport: `stdio` or `http`.                                                                                            |
| `ICUVISOR_HTTP_BIND`       | `127.0.0.1:8765`               | HTTP bind address for Streamable HTTP. Use loopback unless you deliberately want LAN access.                                 |
| `ICUVISOR_DELETE_MODE`     | `safe`                         | Write/delete registration mode. See [safety modes]({{< relref "safety-modes" >}}).                                           |
| `ICUVISOR_TOOLSET`         | `core`                         | Tool catalog tier. See [safety modes]({{< relref "safety-modes" >}}).                                                        |
| `ICUVISOR_DEBUG_METADATA`  | `false`                        | Include debug metadata in MCP responses only when set to `true`.                                                             |
| `ICUVISOR_COACH_MODE`      | `off`                          | Coach-mode feature flag: `off`, `on`, or `auto`.                                                                             |

## Exit codes

| Code | Meaning                                                     |
| ---- | ----------------------------------------------------------- |
| `0`  | Success, including help and version output.                 |
| `2`  | Usage error, such as an unknown flag or missing flag value. |
| `1`  | Runtime error while loading config or running the server.   |

## Full `--help` output

```text
icuvisor connects intervals.icu training data to MCP-compatible AI clients.

Usage:
  icuvisor [flags]
  icuvisor <command> [flags]

Commands:
  (no command)  Run the MCP server (stdio transport by default).
  diagnostics  Print redacted local diagnostics and exit.
  setup         Store intervals.icu credentials and write non-secret config.
  version       Print the icuvisor version and exit.
  help          Print this help and exit.

Flags:
  --config <path>        JSON config file path. Can also be set with ICUVISOR_CONFIG.
  --env-file <path>      Env-file path to load before process env. Can also be set with ICUVISOR_ENV_FILE. Default: .env when present.
  --transport <name>     MCP transport: stdio or http. Default: stdio.
  --http-bind <addr>     HTTP bind address for --transport http. Default: 127.0.0.1:8765.
  -h, --help             Print this help and exit.

Environment variables:
  INTERVALS_ICU_API_KEY      intervals.icu API key. Required unless provided by config/keychain.
  INTERVALS_ICU_ATHLETE_ID   Athlete ID, with or without leading i. Required unless provided by config.
  ICUVISOR_CONFIG            JSON config file path used when --config is omitted.
  ICUVISOR_ENV_FILE          Env-file path used when --env-file is omitted.
  ICUVISOR_TIMEZONE          Athlete timezone. Default: UTC.
  ICUVISOR_API_BASE_URL      intervals.icu API base URL. Default: https://intervals.icu/api/v1.
  ICUVISOR_HTTP_TIMEOUT      HTTP client timeout. Default: 30s.
  ICUVISOR_TRANSPORT         MCP transport: stdio or http. Default: stdio.
  ICUVISOR_HTTP_BIND         HTTP bind address for Streamable HTTP. Default: 127.0.0.1:8765.
  ICUVISOR_DELETE_MODE       Write/delete registration mode: safe, full, or none. Default: safe.
  ICUVISOR_TOOLSET           Tool catalog tier: core or full. Default: core.
  ICUVISOR_DEBUG_METADATA    Include debug metadata in MCP responses when set to true. Default: false.
  ICUVISOR_COACH_MODE        Coach-mode feature flag: off, on, or auto. Default: off.

Examples:
  icuvisor
  icuvisor diagnostics
  icuvisor setup
  icuvisor setup --config /path/to/icuvisor.json
  ICUVISOR_TRANSPORT=http icuvisor
  icuvisor --transport http --http-bind 127.0.0.1:8765
  icuvisor --config /path/to/icuvisor.json

Exit codes:
  0  Success, including help and version output.
  2  Usage error, such as an unknown flag or missing flag value.
  1  Runtime error while loading config or running the server.

For deeper documentation, see README.md and docs/prd/PRD-icuvisor.md.
```
