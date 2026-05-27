---
title: "Install icuvisor"
description: "Choose the right icuvisor install path for your computer."
weight: 10
type: docs
cascade:
  type: docs
---

Install the icuvisor binary for your operating system. Pick your platform below, then head to [Connect]({{< relref "/connect" >}}) to wire it into an AI client.

icuvisor itself is [MIT-licensed and open source](https://github.com/ricardocabral/icuvisor/blob/main/LICENSE), free to install and use: there is no icuvisor-hosted account, onboarding credit, subscription, or icuvisor SaaS quota. Your AI client, model provider, download source, and intervals.icu account may still have their own terms or limits.

## Quick install

The fastest path on Linux, macOS (without Homebrew), WSL, and CI is the shell installer:

```bash
curl -fsSL https://icuvisor.app/install.sh | sh
```

On native Windows / PowerShell:

```powershell
iwr -useb https://icuvisor.app/install.ps1 | iex
```

The installer detects your platform, downloads the matching release asset, verifies the SHA256 checksum (and the cosign signature when `cosign` is on your `PATH`), and installs the binary. Re-run the same command to upgrade in place. See [installer integrity](https://github.com/ricardocabral/icuvisor/blob/main/SECURITY.md#installer-integrity) for signature verification details.

## Platform guides

Prefer a package manager or a signed installer package? Pick your platform:

{{< cards >}}
  {{< card link="macos" title="macOS" subtitle="Install the signed macOS app and verify Gatekeeper status." >}}
  {{< card link="windows" title="Windows" subtitle="Install from the MSI package." >}}
  {{< card link="linux" title="Linux" subtitle="Current Linux packaging status." >}}
{{< /cards >}}
