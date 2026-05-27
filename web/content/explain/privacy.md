---
title: "Privacy posture"
description: "What icuvisor keeps local, what still leaves your machine, and what it does not claim."
---

icuvisor is designed for athletes and coaches who want a small local trust boundary around their intervals.icu data. This page explains the privacy posture behind the local install; it is not legal advice and it is not a certification of GDPR or other regulatory compliance.

For vulnerability reporting, release integrity, and hardening details, see the repository [Security Policy](https://github.com/ricardocabral/icuvisor/blob/main/SECURITY.md).

## Local trust boundary

In the normal local setup, icuvisor runs as a binary on your computer. It reads intervals.icu through the public API using the credential you configured, returns the requested MCP tool response to the AI client you chose, and does not create an icuvisor-hosted athlete database or icuvisor SaaS account.

That keeps the icuvisor-operated trust boundary small: your machine, your configured MCP client, intervals.icu as the upstream training-data service, and optional icuvisor release-update hosts when you install or update the binary.

## Credential storage

`icuvisor setup` stores the intervals.icu API key in the OS credential store by default:

- macOS Keychain on macOS.
- Windows Credential Manager on Windows.
- libsecret/Secret Service on Linux desktops when available.

The local config file is intended for non-secret settings such as athlete ID, timezone, transport, and coach roster settings. Environment variables and legacy plaintext config keys remain available for headless or compatibility workflows, but plaintext credentials are discouraged because they can leak through backups, shell history, logs, or accidental commits.

## HTTP transport default

The default MCP transport is `stdio`. If you enable Streamable HTTP, icuvisor binds to loopback by default at `127.0.0.1:8765`, so same-machine clients can connect without exposing the server to the LAN.

Changing the bind address to a LAN IP is an explicit opt-in. A LAN bind exposes an unauthenticated MCP server to any machine that can reach that address, using the intervals.icu credential configured for this icuvisor process. See [Use Streamable HTTP transport]({{< relref "../guides/http-transport" >}}) before changing it.

## Coach mode

Coach mode still keeps the coach-scoped intervals.icu API key in the server's credential chain. The AI client receives athlete selectors and allowed tools, not the key. An `athlete_id` argument is only a target selector: icuvisor normalizes it, checks it against the configured coach roster, and applies the selected athlete's ACL before calling intervals.icu.

See [Coach mode model]({{< relref "coach-mode" >}}) for the full authorization model.

## AI client and upstream caveats

icuvisor keeps the intervals.icu API key out of tool arguments and responses, but it cannot control what your chosen AI client or model provider does with conversation text and tool-response content. Review that client's data-use, retention, enterprise, and opt-out settings before sending sensitive training context.

intervals.icu remains the upstream service that stores and processes the training data icuvisor reads or writes. Review intervals.icu's own privacy terms and account controls for the upstream data relationship.

## EU and GDPR due-diligence questions

Privacy-conscious and EU users may find this posture useful because the default local install avoids adding an icuvisor-hosted athlete-data service between them and intervals.icu. Treat that as a design choice, not a legal conclusion.

Questions to ask for your own situation include:

- Which machine and OS account can unlock the credential store?
- Which MCP client will receive tool responses, and what are that client's data-use settings?
- Which intervals.icu API key permissions and athlete/coach access are enabled?
- Is HTTP still loopback-only, or has a LAN bind been enabled?
- In coach mode, which athletes are configured and which tools do their ACLs allow?
