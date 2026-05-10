# Security Policy

## Supported versions

icuvisor is pre-1.0. Until the first stable release, only the latest tagged release and `main` receive fixes.

| Version | Supported |
|---------|-----------|
| `main`  | yes       |
| latest tag | yes    |
| older tags | no     |

## Reporting a vulnerability

**Please do not open a public GitHub issue for security problems.**

Use GitHub's private vulnerability reporting:
https://github.com/ricardocabral/icuvisor/security/advisories/new

Or email the maintainer at **security@icuvisor.dev** (until provisioned, contact `@ricardocabral` on GitHub for the current channel).

Please include:

- A description of the issue and the impact you believe it has.
- Steps to reproduce, or a proof-of-concept.
- The version / commit SHA you tested against.
- Your name and affiliation (if any) for credit, or note that you prefer to remain anonymous.

## What to expect

- **Acknowledgement** within 3 business days.
- **Triage update** within 7 business days, including a severity assessment and a target fix window.
- **Coordinated disclosure**: we will agree a disclosure date with you, typically within 90 days of the initial report, sooner for actively exploited issues.
- **Credit** in the release notes and `CHANGELOG.md` once a fix is shipped, unless you ask to remain anonymous.

## Scope

In scope:

- The `icuvisor` binary and any code in this repository.
- Official installers, Homebrew tap, Scoop bucket, Winget manifest.
- The release signing and auto-update pipeline.

Out of scope:

- Vulnerabilities in intervals.icu itself — report those to the platform owner.
- Vulnerabilities in third-party MCP clients (Claude Desktop, ChatGPT, etc.).
- Issues that require physical access to the user's machine or a compromised OS account.
- Social engineering of maintainers.

## Hardening notes for users

- Your intervals.icu API key is stored in the OS keychain, not in plain text on disk.
- The MCP HTTP transport binds to `127.0.0.1` by default. Do not expose it to a public interface unless you understand the risks.
- icuvisor only contacts `intervals.icu` and (if auto-update is enabled) `releases.icuvisor.dev`. Verify network activity against this expectation.
