---
title: "Hosted connector"
description: "Use hosted ICU Visor when your MCP client needs a public HTTPS endpoint."
weight: 1
---

Use hosted mode when your AI client cannot run a local `icuvisor` process and needs a public HTTPS MCP endpoint.

If your client can run a local process or connect to loopback HTTP, use [local mode]({{< relref "/connect" >}}) instead. Local mode keeps your Intervals.icu API key on your machine.

## Connector URL

Paste this URL wherever your AI client asks for a remote MCP server, connector URL, or MCP endpoint:

```text
https://connect.icuvisor.app/mcp
```

Use the provider-specific pages when you want exact UI labels:

- [Claude.ai]({{< relref "claude-ai" >}}): **Customize > Connectors > + > Add custom connector**.
- [ChatGPT]({{< relref "chatgpt" >}}): open **Plugins > + > Add custom MCP server**, then create and install the plugin. Account and workspace policies apply.

## Connect in hosted mode

1. Add this MCP server URL to your client as a custom connector:

   ```text
   https://connect.icuvisor.app/mcp
   ```

2. When the authorization page opens, choose the hosted preferences:
   - **Toolset**: `core` for Claude.ai and ChatGPT daily use, `compact` for smaller/local-compatible clients that need a reduced tool catalog, or `full` for the broader expert catalog.
   - **Write safety**: read-only by default; enable writes or deletes only when you need them.
   - **Calendar and wellness writes**: opt in only when you want Intervals write scopes.
   - **Timezone**: used for hosted date handling.
3. Continue to Intervals.icu and approve the requested OAuth scopes.
4. Return to your MCP client and finish its connector flow.

Hosted mode uses Intervals OAuth. Do not paste an Intervals API key into the hosted connector or into chat.

Once the connector is available in your client, open a fresh chat and try [What can I ask icuvisor?]({{< relref "../cookbook/what-can-i-ask" >}}) for beginner-friendly prompts.

## Manage hosted access

Open hosted settings at:

```text
https://connect.icuvisor.app/settings
```

From settings you can change hosted preferences, start re-consent when broader Intervals scopes are needed, revoke client grants, log out of the settings session, or delete the hosted account.

For hosted privacy details, see <https://connect.icuvisor.app/privacy>.
