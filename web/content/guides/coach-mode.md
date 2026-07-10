---
title: "Set up coach mode"
description: "Configure icuvisor for a coach-held intervals.icu key and multiple athletes."
---

Coach mode lets one locally held, coach-scoped intervals.icu API key target multiple configured athletes. This guide covers setup. The API key is never accepted as an MCP tool argument and is never returned in tool output.

## 1. Decide the mode

Set `ICUVISOR_COACH_MODE` before starting the MCP server:

| Value          | Effect                                                                                    |
| -------------- | ----------------------------------------------------------------------------------------- |
| `off` or unset | Single-athlete mode. Coach tools are not registered.                                      |
| `auto`         | Enable coach mode only when the config file contains a non-empty `coach.athletes` roster. |
| `on`           | Require a non-empty `coach.athletes` roster. Startup fails if the roster is empty.        |

## 2. Add a coach roster to config

Create or edit the JSON config file you pass with `--config` or `ICUVISOR_CONFIG`:

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
        "allowed_tools": ["get_*"],
        "denied_tools": []
      }
    ]
  }
}
```

Athlete IDs may be written as `12345` or `i12345`. icuvisor trims surrounding whitespace, lowercases only an optional leading `I`, validates the remaining digits, and preserves the ID shape: `12345` remains `12345`, while `i12345` remains `i12345`.

`coach.default_athlete_id` must name an athlete in the roster. In enabled coach mode, it becomes the initial selected athlete and wins over legacy top-level `athlete_id` so startup and session defaults are unambiguous.

For field details, see the [config file reference]({{< relref "../reference/config-file" >}}).

## 3. Choose ACL patterns

`allowed_tools` is the positive allow list. `denied_tools` is an explicit veto list. Deny patterns override allow patterns.

Patterns may be:

- `*` for all athlete-scoped tools.
- An exact athlete-scoped tool name, such as [`get_athlete_profile`]({{< relref "../reference/tools#get_athlete_profile" >}}).
- A prefix wildcard ending in `*`, such as `get_*`.

Unknown tool names or patterns that match no athlete-scoped tools fail config loading so ACL typos are caught at startup.

## 4. Start icuvisor with coach mode

```bash
ICUVISOR_COACH_MODE=on \
/Applications/icuvisor.app/Contents/MacOS/icuvisor --config /path/to/coach-config.json
```

Use `auto` instead of `on` if you want the same config style to fall back to single-athlete mode when the roster is empty.

## 5. Use coach tools

Coach mode registers two coach-scoped tools:

- [`list_athletes`]({{< relref "../reference/tools#list_athletes" >}}) returns the configured roster and `_meta.source: "config"`.
- [`select_athlete`]({{< relref "../reference/tools#select_athlete" >}}) changes the default target for subsequent calls in the same MCP session.

Every athlete-scoped tool accepts an optional `athlete_id` argument in coach mode. If omitted, the tool targets the selected athlete. If supplied, the value is normalized and checked against the configured roster before any intervals.icu request.

## 6. Run a first-session onboarding check

After connecting a new athlete or team, start with the `coach_athlete_onboarding` MCP prompt or ask the assistant to run the same checklist manually:

```text
Use icuvisor's coach athlete onboarding flow for athlete i12345.
Confirm the selected athlete before summarizing data, then give me the
first-session checklist, missing data warnings, baseline context, races/goals
questions, and next coach actions. Do not modify any calendar or settings.
```

The first session should:

1. Call [`list_athletes`]({{< relref "../reference/tools#list_athletes" >}}) and, when needed, [`select_athlete`]({{< relref "../reference/tools#select_athlete" >}}) so the coach sees the canonical athlete ID/label being analyzed.
2. State that the coach must already have permission from that athlete to view and analyze their data. icuvisor enforces the local roster and ACLs; it does not collect consent in chat.
3. Read [`get_athlete_profile`]({{< relref "../reference/tools#get_athlete_profile" >}}) first for timezone, units, FTP/thresholds, zones, and profile warnings.
4. Check recent activity coverage, wellness/HRV freshness, fitness/load trend, upcoming races/events, training-plan context, and unavailable ACL/tool surfaces.
5. Return a pass/warn/missing checklist for thresholds/zones, activity data, wellness/HRV baseline, races/goals, devices or sync sources, and any missing-data warnings.
6. Ask follow-up questions for goals, race priority, constraints, communication preferences, and device/source fixes before giving plan advice.

Keep onboarding read-only. If the checklist implies calendar or settings changes, ask for a reviewed before/after proposal and explicit approval before using write tools.

## Catalog-cache caveat

MCP clients may cache the tool catalog for the current conversation. [`select_athlete`]({{< relref "../reference/tools#select_athlete" >}}) changes server-side routing immediately, and per-call `athlete_id` overrides are enforced immediately, but the model or client may not see a refreshed tools list until a new conversation or reconnect.

When [`select_athlete`]({{< relref "../reference/tools#select_athlete" >}}) returns `_meta.requires_new_conversation: true`, start a new conversation or reconnect the MCP client before relying on newly visible or hidden tools.

## Common setup errors

| Symptom                                                                                                                                                   | Fix                                                                                                                   |
| --------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| [`list_athletes`]({{< relref "../reference/tools#list_athletes" >}}) or [`select_athlete`]({{< relref "../reference/tools#select_athlete" >}}) is missing | Confirm `ICUVISOR_COACH_MODE=on` or `auto` with a non-empty `coach.athletes` roster, then restart the MCP server.     |
| A tool is missing after selecting an athlete                                                                                                              | Check all gates: delete mode, toolset tier, and the selected athlete ACL.                                             |
| Tool call returns `invalid athlete_id; use a configured target athlete`                                                                                   | Verify the athlete is in `coach.athletes`, the ID is valid, and the selected or per-call athlete ACL allows the tool. |
| Config fails with an unknown tool name                                                                                                                    | Fix the typo in `allowed_tools` or `denied_tools`; icuvisor validates ACL names against the shared catalog.           |
