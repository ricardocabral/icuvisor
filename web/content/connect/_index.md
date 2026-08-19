---
title: "Connect an AI client"
description: "Connect icuvisor to Claude, ChatGPT, Cursor, Continue, Zed, Pi, or another MCP client."
weight: 20
type: docs
cascade:
  type: docs
---

ICU Visor has two ways to operate:

- **Local mode**: install the `icuvisor` binary, run `icuvisor setup`, and connect a client to the local MCP server. Use this when the client can launch a local process or reach loopback HTTP.
- **Hosted mode**: add `https://connect.icuvisor.app/mcp` as a remote HTTPS MCP connector and sign in with Intervals.icu OAuth. Use this when the client needs a public endpoint.

Pick the path that matches your client. After the connection works, start with [What can I ask icuvisor?]({{< relref "../cookbook/what-can-i-ask" >}}) for beginner-friendly prompts before reading the full tool reference.

{{< cards >}}
  {{< card link="hosted" title="Hosted connector" subtitle="Use the public HTTPS MCP URL for clients that cannot run a local server." >}}
  {{< card link="claude-ai" title="Claude.ai" subtitle="Add hosted icuvisor from Claude's custom connector settings." >}}
  {{< card link="claude-desktop" title="Claude Desktop" subtitle="Install as a Desktop Extension or configure the MCP stdio server by hand." >}}
  {{< card link="claude-code" title="Claude Code" subtitle="Start icuvisor over MCP stdio from Claude Code." >}}
  {{< card link="codex-cli" title="Codex CLI" subtitle="Register icuvisor with Codex MCP config or the codex mcp command." >}}
  {{< card link="chatgpt" title="ChatGPT" subtitle="Minimal ChatGPT MCP connection notes." >}}
  {{< card link="gemini-spark" title="Gemini Spark" subtitle="Core Streamable HTTP setup and reachability limits." >}}
  {{< card link="other-clients" title="Other MCP clients" subtitle="The standard MCP JSON shape for Cursor, Continue, Zed, Pi, and more." >}}
{{< /cards >}}
