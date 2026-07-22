<p align="center">
  <img src="docs/brand/logo-wordmark.png" alt="icuvisor" width="540" />
</p>

[![CI](https://github.com/ricardocabral/icuvisor/actions/workflows/ci.yml/badge.svg)](https://github.com/ricardocabral/icuvisor/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/ricardocabral/icuvisor?sort=semver)](https://github.com/ricardocabral/icuvisor/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

> icuvisor is an open-source, local-first [Model Context Protocol](https://modelcontextprotocol.io) server for [intervals.icu](https://intervals.icu). Run the signed binary on your computer to connect an MCP-compatible AI client to your training data without putting your intervals.icu API key in chat. If your client needs a public HTTPS endpoint, use the optional hosted connector instead.

## Get started

1. [Install icuvisor](https://icuvisor.app/install/).
2. [Connect your AI client](https://icuvisor.app/connect/).
3. Try a prompt from the [Cookbook](https://icuvisor.app/cookbook/).

The documentation covers local and hosted modes, privacy, safety modes, supported clients, troubleshooting, and the complete tool reference.

## Direct CLI

Releases also include `icuvisor-cli`, a tools-only command interface for local scripts and agents. It uses the same registered handlers and safety gates as MCP, loads credentials from local configuration or the OS keychain, and does not support coach mode.

```bash
icuvisor-cli capabilities
icuvisor-cli doctor
icuvisor-cli tools list
icuvisor-cli tools describe get_today
icuvisor-cli tools call get_today --args '{}'
echo '{}' | icuvisor-cli tools call get_today --args-file -
```

Successful commands write JSON to stdout. Failures leave stdout empty and write one JSON error to stderr, using exit code `2` for usage errors and `1` for runtime errors. See the [CLI reference](https://icuvisor.app/reference/cli/) for the complete contract.

## For contributors

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and contribution guidelines. Product scope and planned work live in the [PRD](docs/prd/PRD-icuvisor.md) and [roadmap](ROADMAP.md).

## Security

See [SECURITY.md](SECURITY.md) for supported versions, vulnerability reporting, and installer integrity.

## License

[MIT](LICENSE).
