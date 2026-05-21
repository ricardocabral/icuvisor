---
title: "Connect ChatGPT"
description: "Minimal ChatGPT MCP connection notes for icuvisor."
---

ChatGPT MCP support is evolving across client and developer-mode surfaces. Use this page as the focused connection shape: point ChatGPT's MCP configuration at the local icuvisor binary and keep secrets out of the model conversation.

## Before you start

- Install icuvisor and run setup.
- Confirm the binary starts with `icuvisor version`.
- Know your non-secret athlete ID and timezone.

## Stdio configuration shape

Use the same stdio server definition as the Claude clients when ChatGPT asks for a local MCP server command:

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

If your ChatGPT MCP surface expects a single server object rather than a full `mcpServers` map, use the `icuvisor` object from the example as that server definition.

## HTTP alternative

If your ChatGPT MCP surface expects an HTTP URL, start icuvisor with Streamable HTTP on loopback:

```bash
ICUVISOR_TRANSPORT=http /Applications/icuvisor.app/Contents/MacOS/icuvisor
```

Then point the client at:

```text
http://127.0.0.1:8765/mcp
```

Do not bind icuvisor to a LAN address unless you intentionally want other machines to reach the unauthenticated local MCP server. The HTTP transport guide covers the security tradeoff in more detail.

## Remote connector UIs are not supported yet

ChatGPT-style remote custom connector UIs are different from local MCP client configuration. They run from the provider's infrastructure and require an HTTPS MCP endpoint that is reachable from that infrastructure. They cannot call `http://127.0.0.1:8765/mcp` on your laptop, and icuvisor intentionally does not add legacy SSE or a generic cloudflared/ngrok recipe for this.

This is a product safety boundary, not a bug in the Streamable HTTP transport. Tunneling the local server would make an unauthenticated MCP endpoint reachable from the public internet. Anyone who can reach that endpoint could invoke the tools registered in your icuvisor process using the intervals.icu credentials configured for that process.

Remote ChatGPT connectors are out of scope until the vNext hosted relay or a future explicit secure-tunnel design. Use local ChatGPT MCP surfaces that accept stdio or loopback Streamable HTTP for now.

## Verify

Start a fresh ChatGPT conversation after saving the MCP configuration, then ask a simple profile question such as `What's my FTP?` The expected tool call is [`get_athlete_profile`]({{< relref "../reference/tools#get_athlete_profile" >}}).
