---
title: "Connect Gemini Spark"
description: "Use Gemini Spark with the core icuvisor Streamable HTTP protocol where the client can reach a local MCP endpoint."
weight: 8
---

This page documents the **core icuvisor** connection boundary for Gemini Spark. It does not provide a hosted Gemini connector or OAuth service.

## Compatibility status

The core server has an in-process Streamable HTTP smoke test for the MCP protocol flow Gemini needs: `initialize`, `ping`, `tools/list`, and a representative read-only tool call. This verifies protocol envelopes and the local transport, not the Gemini Spark product experience.

No end-to-end Gemini Spark account or mobile test has been performed for this repository. Do not treat the protocol test as a claim that every Gemini surface, plan, web flow, or mobile app supports icuvisor.

## Local setup

Use this path only when the Gemini client can reach a process running on the same computer:

```bash
/Applications/icuvisor.app/Contents/MacOS/icuvisor --transport http --http-bind 127.0.0.1:8765
```

The core MCP endpoint is:

```text
http://127.0.0.1:8765/mcp
```

Keep the process running while the client connects. Core local HTTP is loopback-bound and has no OAuth layer. Keep the bind at `127.0.0.1`; a LAN bind exposes an unauthenticated MCP server to other reachable hosts. Store the intervals.icu API key with `icuvisor setup`, not in Gemini settings or a chat.

For the transport details and security warning, see [Use Streamable HTTP transport]({{< relref "../guides/http-transport" >}}).

## Reachability limits

A Gemini web or mobile surface cannot normally reach `127.0.0.1` on your computer. The loopback URL is not a public or mobile endpoint. Do not replace it with a generic public tunnel, and do not put an API key in a URL or client configuration.

If Gemini requires a public HTTPS MCP URL, OAuth, or Dynamic Client Registration (DCR), that is hosted integration work outside this core repository. The corresponding work belongs in `icuvisor-host`; this page does not promise that hosted flow.

After changing the local server configuration, reconnect the MCP server and start a fresh Gemini conversation so the client can refresh its tool catalog. If the client cannot launch or reach this local endpoint, use a client with local MCP support or wait for a separately documented hosted integration.
