---
title: "Config file reference"
description: "JSON fields accepted by icuvisor config files."
---

Most users can run [`icuvisor setup`]({{< relref "cli#commands" >}}) and let icuvisor write non-secret configuration. Use this reference when you maintain a JSON config by hand or when an MCP client points icuvisor at a config file with `--config` or `ICUVISOR_CONFIG`.

When neither flag is set, icuvisor also loads the platform default config path if the file exists: `$XDG_CONFIG_HOME/icuvisor/config.json` on Linux, `~/Library/Application Support/icuvisor/config.json` on macOS, and `%AppData%\icuvisor\config.json` on Windows. This is the same path `icuvisor setup` writes by default.

Keep API keys in the OS keychain whenever possible. `icuvisor setup` writes a non-secret `credential_ref` that documents the keychain service/account it uses; the actual key remains outside JSON. The legacy `api_key` JSON field still exists for compatibility, but plaintext config keys emit a warning and should not be committed, synced, or backed up.

## Example

```json
{
  "credential_ref": {
    "type": "keychain",
    "service": "icuvisor",
    "account": "intervals-icu-api-key"
  },
  "athlete_id": "i12345",
  "timezone": "America/Sao_Paulo",
  "api_base_url": "https://intervals.icu/api/v1",
  "http_timeout": "30s",
  "transport": "stdio",
  "http_bind": "127.0.0.1:8765"
}
```

## Top-level fields

| Field          | Type   | Default                                                        | Description                                                                                                                                                                                                        |
| -------------- | ------ | -------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `credential_ref` | object | omitted for hand-written configs                               | Non-secret metadata written by `icuvisor setup`: `{ "type": "keychain", "service": "icuvisor", "account": "intervals-icu-api-key" }`. It documents the OS keychain location; it is not a secret and does not override credential precedence. |
| `athlete_id`     | string | required unless coach mode supplies `coach.default_athlete_id` | intervals.icu athlete ID. Both `12345` and `i12345` are accepted; icuvisor normalizes to `i12345`.                                                                                                                 |
| `timezone`       | string | `UTC`                                                          | IANA timezone used for athlete-local dates and presentation, for example `UTC`, `Europe/London`, or `America/Sao_Paulo`. Invalid names fail config loading.                                                        |
| `api_base_url` | string | `https://intervals.icu/api/v1`                                 | Absolute `http` or `https` intervals.icu API base URL. Most users should not change this. Trailing slashes are trimmed.                                                                                            |
| `http_timeout` | string | `30s`                                                          | HTTP client timeout as a positive Go duration, such as `10s`, `30s`, or `1m`.                                                                                                                                      |
| `transport`    | string | `stdio`                                                        | MCP transport. Accepted values: `stdio` or `http`.                                                                                                                                                                 |
| `http_bind`    | string | `127.0.0.1:8765`                                               | IP address and port for Streamable HTTP. The host must be an explicit IP address, such as `127.0.0.1:8765` or `192.168.1.10:8765`.                                                                                 |
| `coach`        | object | omitted                                                        | Optional coach-mode roster and per-athlete ACLs. See [Set up coach mode]({{< relref "../guides/coach-mode" >}}) for setup and [Coach mode model]({{< relref "../explain/coach-mode" >}}) for the conceptual model. |
| `api_key`      | string | not recommended                                                | Legacy plaintext API-key fallback. Prefer the OS keychain or `INTERVALS_ICU_API_KEY` for emergency/headless use.                                                                                                   |

Implementation source of truth: `internal/config/config.go` and `internal/coach/config.go`.

## Coach config object

Coach mode is enabled with `ICUVISOR_COACH_MODE=on` or `ICUVISOR_COACH_MODE=auto`; see the [CLI reference]({{< relref "cli#environment-variables" >}}). The JSON config supplies the roster and ACLs.

```json
{
  "coach": {
    "default_athlete_id": "i12345",
    "athletes": [
      {
        "id": "i12345",
        "label": "Jane",
        "allowed_tools": ["*"],
        "denied_tools": ["delete_event", "delete_events_by_date_range"]
      },
      {
        "id": "i67890",
        "label": "Bob",
        "allowed_tools": ["get_*"],
        "denied_tools": []
      }
    ]
  }
}
```

| Field                            | Type             | Description                                                                                                                                                              |
| -------------------------------- | ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `coach.default_athlete_id`       | string           | Initial selected athlete when coach mode is enabled. It must name an athlete in `coach.athletes`. If there is exactly one athlete, icuvisor can fill it from that entry. |
| `coach.athletes`                 | array            | Configured roster. `ICUVISOR_COACH_MODE=on` requires this array to be non-empty; `auto` enables coach mode only when it is non-empty.                                    |
| `coach.athletes[].id`            | string           | Athlete selector, normalized the same way as top-level `athlete_id`.                                                                                                     |
| `coach.athletes[].label`         | string           | Human-readable label for the roster entry.                                                                                                                               |
| `coach.athletes[].allowed_tools` | array of strings | Positive allow list. Accepts exact athlete-scoped tool names, `*`, or a prefix wildcard such as `get_*`. Empty means deny all.                                           |
| `coach.athletes[].denied_tools`  | array of strings | Explicit veto list. Deny patterns override allow patterns.                                                                                                               |

Unknown tool names or ACL patterns that match no athlete-scoped tools fail config loading so typos are caught before the MCP server starts. [`icuvisor_list_advanced_capabilities`]({{< relref "tools#icuvisor_list_advanced_capabilities" >}}) is a meta/control tool, not an athlete-scoped ACL entry.

## Environment and flag overrides

Process environment values override JSON fields for the same setting, and CLI flags override transport-related settings for that process. See [CLI reference]({{< relref "cli" >}}) for the exact names.

For normal user setup:

1. Store the API key in the OS keychain with `icuvisor setup`.
2. Put non-secret values in the config file or MCP client environment. The generated `credential_ref` can stay in JSON because it contains only keychain metadata.
3. Point the MCP client at the config with `--config /path/to/config.json` or `ICUVISOR_CONFIG=/path/to/config.json` when you want JSON config loading.
