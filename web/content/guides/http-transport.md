---
title: "Use Streamable HTTP transport"
description: "Enable icuvisor's Streamable HTTP MCP transport and understand the LAN warning."
---

`stdio` is the default MCP transport and is the right choice for most local desktop clients. Streamable HTTP is available for clients that need a local HTTP URL.

## Start HTTP on loopback

Pass the explicit loopback bind when starting HTTP:

macOS:

```bash
/Applications/icuvisor.app/Contents/MacOS/icuvisor --transport http --http-bind 127.0.0.1:8765
```

Windows PowerShell:

```powershell
& "$env:LOCALAPPDATA\Programs\icuvisor\icuvisor.exe" --transport http --http-bind 127.0.0.1:8765
```

By default, icuvisor listens on loopback only:

```text
http://127.0.0.1:8765/mcp
```

Config files can also set:

```json
{
  "transport": "http",
  "http_bind": "127.0.0.1:8765"
}
```

See the [CLI reference]({{< relref "../reference/cli" >}}) and [config file reference]({{< relref "../reference/config-file" >}}) for exact field names and defaults. To keep this loopback service running without an open terminal, follow [Keep local HTTP running]({{< relref "persistent-http-service" >}}).

## Stateless HTTP and JSON responses

The default HTTP server retains sessions and returns SSE responses. For a gateway or container deployment that needs independent requests and ordinary JSON responses, enable both options:

```bash
icuvisor --transport http --http-stateless=true --http-json-response=true
```

The equivalent environment configuration is:

```bash
ICUVISOR_TRANSPORT=http ICUVISOR_HTTP_STATELESS=true ICUVISOR_HTTP_JSON_RESPONSE=true icuvisor
```

Or add these fields to your existing JSON config:

```json
{
  "transport": "http",
  "http_stateless": true,
  "http_json_response": true
}
```

Both options default to `false` and can be enabled independently. CLI flags accept an explicit `true` or `false`, either with `=` or a separate value. Precedence is CLI flags, process environment, JSON config, `.env`, then defaults. An explicit `false` overrides a lower-priority `true`. These options affect HTTP only; stdio is unchanged.

- `http_stateless` disables MCP session IDs and per-session state. Each request is independent; standalone GET streams and session DELETE requests return `405 Method Not Allowed`. Server-to-client requests are unavailable. Coach mode over HTTP requires sessions for athlete selection, so startup rejects stateless HTTP when coach mode is enabled, including `auto` with a roster.
- `http_json_response` returns `application/json` instead of `text/event-stream` for POST request responses. It does not disable sessions on its own. Keep it `false` if your client needs streamed progress notifications.

These settings do not add authentication, HTTPS, or a read-only policy. For a public gateway deployment, authenticate clients and terminate HTTPS at the gateway, and keep the icuvisor backend inaccessible directly. The loopback bind and existing tool/write policy remain unchanged.

## Configure the client URL

Use this endpoint in MCP clients that require HTTP:

```text
http://127.0.0.1:8765/mcp
```

If the client runs on the same computer, do not change the bind address.

## LAN binding warning

Do not change `ICUVISOR_HTTP_BIND` or `--http-bind` to make a remote client work. A LAN bind exposes an unauthenticated MCP server: anyone who can connect to that address can call registered tools using the intervals.icu credentials configured for this icuvisor process. Keep `127.0.0.1:8765` unless you understand and explicitly accept that risk.

This guide deliberately provides no LAN-start command. icuvisor logs a warning when HTTP starts on a non-loopback bind. For how the loopback default fits into icuvisor's local privacy posture, see [Privacy posture]({{< relref "../explain/privacy" >}}).

## Troubleshooting

| Symptom                                                        | Fix                                                                                                                                                                                                                                                                                                                      |
| -------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Client cannot connect to `127.0.0.1:8765`                      | Confirm icuvisor is running with HTTP transport, not stdio.                                                                                                                                                                                                                                                              |
| Startup fails with an invalid bind error                       | Use an explicit IP address and numeric port, such as `127.0.0.1:8765`; hostnames are not accepted.                                                                                                                                                                                                                       |
| Another machine cannot connect                                 | Confirm you intentionally bound a LAN IP, that the OS firewall allows the port, and that the risk is acceptable.                                                                                                                                                                                                         |
| ChatGPT-style remote connector UI rejects or cannot reach icuvisor | This is expected. Remote connector UIs run from provider infrastructure and cannot reach `http://127.0.0.1:8765/mcp` on your laptop. Use the [hosted ICU Visor connector]({{< relref "../connect/hosted" >}}) at `https://connect.icuvisor.app/mcp` for a provider-reachable HTTPS MCP endpoint. Do not expose local loopback HTTP through a generic public tunnel; a tunnel URL is not authentication. |
