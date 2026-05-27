---
title: "Local-first design"
description: "Why icuvisor keeps credentials and data flow on your computer."
---

icuvisor is local-first: the server binary runs on your computer, reads your intervals.icu data with your credentials, and returns only the requested tool response to the AI client you chose.

The goal is to keep the connector simple and inspectable. You install one signed binary, connect it to intervals.icu with your own API key, and point an MCP-compatible client at that local process. There is no icuvisor-hosted account, relay, onboarding credit, OAuth consent screen, or SaaS quota in the normal local setup.

## Why the API key lives in the OS keychain

An intervals.icu API key can read and change training data according to the permissions intervals.icu gives that key. It should not be pasted into chat or saved in MCP client JSON.

`icuvisor setup` stores the key in the OS keychain instead:

- macOS Keychain on macOS.
- Windows Credential Manager on Windows.
- libsecret/Secret Service on Linux desktops.

The non-secret config file can hold athlete ID, timezone, transport, and coach roster settings. The key stays separate.

## Local binary versus hosted connector flows

Some MCP integrations are hosted services: you create an account with the connector, authorize access through an OAuth-style flow, and route tool calls through that provider's infrastructure. That can be convenient, but it adds another account, another availability dependency, and another place where credentials or delegated access may be held.

icuvisor takes the smaller local path:

| Concern | Local icuvisor binary | Hosted connector or OAuth-style flow |
| --- | --- | --- |
| Connector account | No icuvisor account to create. | Usually requires a connector/provider account. |
| intervals.icu credential | Stored in the OS keychain on your machine. | Often stored or delegated to the hosted provider. |
| Tool-call path | MCP client ↔ local icuvisor ↔ intervals.icu. | MCP client ↔ hosted connector ↔ upstream service. |
| Operational dependency | Your machine, your client, intervals.icu, and your model provider. | Those services plus the hosted connector's availability and quota model. |

This is not a promise that nothing ever leaves your machine. icuvisor calls intervals.icu's public API, and your chosen AI client may send conversation content and tool results to its model provider according to that client's own terms and settings. The local-first promise is narrower and concrete: icuvisor does not custody your intervals.icu API key or run an icuvisor SaaS middle layer for the normal setup.

## What leaves your machine

The local icuvisor process calls intervals.icu's public API and sends tool responses to the local MCP client process. icuvisor does not host athlete data on an icuvisor server, and there is no icuvisor SaaS account in the normal local setup.

Your chosen AI client may send conversation content to its model provider according to that client's own terms and settings. icuvisor's job is to keep the intervals.icu credential out of the conversation and provide the smallest useful data response for each tool call.

## Why local-first matters

Local-first keeps the trust boundary small:

- You can inspect the open-source binary's behavior.
- You do not hand an API key to an icuvisor-hosted service.
- You can revoke or replace the intervals.icu key yourself.
- You can choose which MCP clients to connect.
- You have fewer moving parts to check when a client caches old tools or a local config changes.

For setup steps, see [API key setup]({{< relref "../guides/api-key" >}}). For the broader privacy model, see [Privacy posture]({{< relref "privacy" >}}).
