---
title: "Use Streamable HTTP transport"
description: "Enable icuvisor's Streamable HTTP MCP transport and understand the LAN warning."
---

`stdio` is the default MCP transport and is the right choice for most local desktop clients. Streamable HTTP is available for clients that need a local HTTP URL.

## Start HTTP on loopback

Run icuvisor with `ICUVISOR_TRANSPORT=http`:

```bash
ICUVISOR_TRANSPORT=http /Applications/icuvisor.app/Contents/MacOS/icuvisor
```

By default, icuvisor listens on loopback only:

```text
http://127.0.0.1:8765/mcp
```

Equivalent flags:

```bash
/Applications/icuvisor.app/Contents/MacOS/icuvisor --transport http --http-bind 127.0.0.1:8765
```

Config files can also set:

```json
{
  "transport": "http",
  "http_bind": "127.0.0.1:8765"
}
```

See the [CLI reference]({{< relref "../reference/cli" >}}) and [config file reference]({{< relref "../reference/config-file" >}}) for exact field names and defaults.

## Configure the client URL

Use this endpoint in MCP clients that require HTTP:

```text
http://127.0.0.1:8765/mcp
```

If the client runs on the same computer, do not change the bind address.

## LAN binding warning

Only set `ICUVISOR_HTTP_BIND` or `--http-bind` to a LAN address when you deliberately want another machine to reach the server.

A LAN bind exposes an unauthenticated MCP server: anyone who can connect to that address can call registered tools using the intervals.icu credentials configured for this icuvisor process. Keep `127.0.0.1:8765` unless you understand that risk.

If you do opt in, use an explicit IP address and port:

```bash
ICUVISOR_TRANSPORT=http \
ICUVISOR_HTTP_BIND=192.168.1.10:8765 \
/Applications/icuvisor.app/Contents/MacOS/icuvisor
```

icuvisor logs a warning when HTTP starts on a non-loopback bind. For how the loopback default fits into icuvisor's local privacy posture, see [Privacy posture]({{< relref "../explain/privacy" >}}).

## Troubleshooting

| Symptom                                                        | Fix                                                                                                                                                                                                                                                                                                                      |
| -------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Client cannot connect to `127.0.0.1:8765`                      | Confirm icuvisor is running with HTTP transport, not stdio.                                                                                                                                                                                                                                                              |
| Startup fails with an invalid bind error                       | Use an explicit IP address and numeric port, such as `127.0.0.1:8765`; hostnames are not accepted.                                                                                                                                                                                                                       |
| Another machine cannot connect                                 | Confirm you intentionally bound a LAN IP, that the OS firewall allows the port, and that the risk is acceptable.                                                                                                                                                                                                         |
| ChatGPT-style remote connector UI rejects or cannot reach icuvisor | This is expected. Remote connector UIs run from provider infrastructure, cannot reach `http://127.0.0.1:8765/mcp` on your laptop, and require a provider-reachable HTTPS MCP endpoint. icuvisor intentionally does not support that mode until the hosted relay or a future explicit secure-tunnel design. Do not treat a generic public tunnel URL as authentication. |
