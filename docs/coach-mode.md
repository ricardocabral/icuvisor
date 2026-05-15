# Coach mode

Coach mode lets one locally held, coach-scoped intervals.icu API key target multiple configured athletes. The key is loaded through the normal icuvisor credential chain (environment, OS keychain, or legacy local config fallback); it is never accepted as an MCP tool argument and is never returned in tool output.

Coach mode is off by default. When it is off, icuvisor keeps the single-athlete catalog and does not register `list_athletes` or `select_athlete`.

## Enable coach mode

Set `ICUVISOR_COACH_MODE` before starting the MCP server:

- `off` or unset: single-athlete mode.
- `auto`: enable coach mode only when the config file contains a non-empty `coach.athletes` roster.
- `on`: require a non-empty `coach.athletes` roster.

Coach mode composes with the existing catalog gates:

1. `ICUVISOR_DELETE_MODE` controls write/delete capability.
2. `ICUVISOR_TOOLSET` controls core versus full catalog tier.
3. The per-athlete coach ACL controls what the active athlete may use.

A tool is exposed only when all applicable gates allow it. Any deny wins.

## Config schema

Add a `coach` stanza to the JSON config file written by setup or supplied with `ICUVISOR_CONFIG`:

```json
{
  "athlete_id": "i12345",
  "timezone": "UTC",
  "coach": {
    "default_athlete_id": "i12345",
    "athletes": [
      {
        "id": "i12345",
        "label": "Jane (active client)",
        "allowed_tools": ["*"],
        "denied_tools": ["delete_event", "delete_events_by_date_range"]
      },
      {
        "id": "i67890",
        "label": "Bob (prospect, read-only)",
        "allowed_tools": ["get_*", "list_athletes", "select_athlete"],
        "denied_tools": []
      }
    ]
  }
}
```

Athlete IDs may be written as `12345` or `i12345`; icuvisor normalizes them to `i12345`. `coach.default_athlete_id` must name an athlete in the roster and becomes the initial selected athlete when coach mode is enabled. In enabled coach mode, it also wins over a legacy top-level `athlete_id` so startup and session defaults are unambiguous.

`allowed_tools` is the positive allow list. `denied_tools` is an explicit veto list. Patterns may be exact tool names or a trailing-prefix wildcard such as `get_*`; `*` matches every tool. Deny patterns override allow patterns. Unknown tool names or patterns that do not match the shared tool catalog fail config loading so ACL typos are caught at startup.

## ACL examples

Full access for an active client, except destructive event deletion:

```json
{
  "id": "i12345",
  "label": "Jane (active client)",
  "allowed_tools": ["*"],
  "denied_tools": ["delete_event", "delete_events_by_date_range"]
}
```

Read-only access for a prospect:

```json
{
  "id": "i67890",
  "label": "Bob (prospect)",
  "allowed_tools": ["get_*", "icuvisor_list_advanced_capabilities"],
  "denied_tools": []
}
```

Deny-all placeholder for an athlete who should remain in the roster but not be usable yet:

```json
{
  "id": "i99999",
  "label": "Paused client",
  "allowed_tools": [],
  "denied_tools": []
}
```

## Tools and routing

Coach mode registers two coach-scoped tools:

- `list_athletes` returns the configured roster and `_meta.source: "config"`. Upstream roster discovery is intentionally deferred until a real coach-scoped intervals.icu key validates the endpoint.
- `select_athlete` changes the default target for subsequent calls in the same MCP session. It returns `previous_athlete_id`, `new_athlete_id`, the newly visible `allowed_tools`, `_meta.scope`, and `_meta.requires_new_conversation`.

Every athlete-scoped tool accepts an optional `athlete_id` argument in coach mode. If omitted, the tool targets the selected athlete. If supplied, the value is normalized and checked against the configured roster before any intervals.icu request. A value outside the roster, malformed value, or ACL-denied target returns the same short error so roster membership cannot be enumerated through error text.

## Catalog-cache caveat

MCP clients may cache the tool catalog for the current conversation. `select_athlete` changes server-side routing immediately, and per-call `athlete_id` overrides are enforced immediately, but the model or client may not see a refreshed `tools/list` until a new conversation or reconnect.

When `select_athlete` returns `_meta.requires_new_conversation: true`, start a new conversation or reconnect the MCP client before relying on newly visible or hidden tools. This avoids stale catalog entries after switching between athletes with different ACLs. TP-040 is expected to add notifications so clients can refresh catalogs more gracefully.

## Threat-model summary

`athlete_id` is a target selector, not a credential. The coach API key remains in the local icuvisor process and is not accepted from or returned to the LLM. Request routing normalizes and authorizes the target athlete before calling intervals.icu, and the per-athlete ACL is evaluated in addition to delete-mode and toolset gates.

This means an LLM-controlled `athlete_id` cannot:

- read or exfiltrate the coach API key;
- select an athlete absent from the configured roster;
- bypass a per-athlete ACL by naming a different athlete; or
- re-enable a tool hidden by `ICUVISOR_DELETE_MODE` or `ICUVISOR_TOOLSET`.

The full writeup lives in [`docs/threat-models/coach-mode.md`](threat-models/coach-mode.md).

## Troubleshooting

- `list_athletes` or `select_athlete` is missing: confirm `ICUVISOR_COACH_MODE=on` or `auto` with a non-empty `coach.athletes` roster, then restart the MCP server.
- A tool is missing after selecting an athlete: check all three gates. Delete tools require `ICUVISOR_DELETE_MODE=full`, full-tier tools require `ICUVISOR_TOOLSET=full`, and the athlete ACL must allow the tool.
- A tool call returns `invalid athlete_id; use a configured target athlete`: verify the athlete is in `coach.athletes`, the ID is valid, and the selected/per-call athlete ACL allows the tool. The message is intentionally generic to avoid roster enumeration.
- The client still shows stale tools after `select_athlete`: start a new conversation or reconnect the MCP client when `_meta.requires_new_conversation` is `true`.
- Config fails at startup with an unknown tool name: fix the typo in `allowed_tools` or `denied_tools`; icuvisor validates ACL names against the shared catalog.
