---
title: "Install icuvisor"
description: "Choose the right icuvisor install path for your computer."
---

icuvisor runs locally on your computer as a single binary. Install it first, then store your intervals.icu API key and connect an MCP client.

## Choose your platform

| Platform          | Status                                | Start here                                                                                 |
| ----------------- | ------------------------------------- | ------------------------------------------------------------------------------------------ |
| macOS             | Signed DMG for the beta release line. | [Install on macOS]({{< relref "macos" >}})                                                 |
| Windows           | Unsigned MSI from CI; signed at v1.0. | [Install on Windows]({{< relref "windows" >}})                                             |
| Linux             | Packages planned for v1.0.            | [Linux status]({{< relref "linux" >}})                                                     |
| Build from source | Developer/power-user path.            | [Build from source on GitHub](https://github.com/ricardocabral/icuvisor#build-from-source) |

## After installing

1. Run `icuvisor setup` to store your intervals.icu API key in the OS keychain and write non-secret configuration.
2. Run `icuvisor version` to confirm the binary starts.
3. [Connect your AI client]({{< relref "../connect" >}}) to the installed binary.

The exact command path depends on your platform. On macOS, MCP clients should execute:

```text
/Applications/icuvisor.app/Contents/MacOS/icuvisor
```

For CLI flags and environment variables, see the [CLI reference]({{< relref "../reference/cli" >}}).
