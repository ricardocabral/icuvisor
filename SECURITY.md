# Security Policy

## Supported versions

icuvisor is pre-1.0. Until the first stable release, only the latest tagged release and `main` receive fixes.

| Version    | Supported |
| ---------- | --------- |
| `main`     | yes       |
| latest tag | yes       |
| older tags | no        |

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

## Release signing and notarization

Official macOS releases use a Developer ID Application certificate for the app bundle and Apple notarization for the DMG. Maintainers must provision the certificate in Apple Developer, export it as a password-protected `.p12`, and store only the following GitHub Actions secrets for release jobs:

- `APPLE_TEAM_ID` — Apple Developer Team ID used for code signing and notarization metadata.
- `APPLE_DEVELOPER_ID_P12_BASE64` — base64-encoded Developer ID Application `.p12` export.
- `APPLE_DEVELOPER_ID_P12_PASSWORD` — password for the `.p12` import.
- `APPLE_API_KEY_ID` — App Store Connect API key ID for `notarytool`.
- `APPLE_API_KEY_ISSUER` — App Store Connect issuer UUID for `notarytool`.
- `APPLE_API_KEY_BASE64` — base64-encoded App Store Connect API key (`.p8`).

Do not commit certificate exports, `.p8` files, app-specific passwords, API keys, or decoded secret material. Release logs must not echo these values.

Before cutting a signed macOS release, the release operator must complete this preflight gate and record the non-secret results in the release notes or task status:

- `security find-identity -v -p codesigning` on the release runner/keychain shows a valid `Developer ID Application` identity for the project.
- Apple Team ID, Developer ID Application common name, and certificate expiration date are known to the release operator.
- GitHub Actions secrets exist by name for `APPLE_TEAM_ID`, `APPLE_DEVELOPER_ID_P12_BASE64`, `APPLE_DEVELOPER_ID_P12_PASSWORD`, `APPLE_API_KEY_ID`, `APPLE_API_KEY_ISSUER`, and `APPLE_API_KEY_BASE64`.
- No secret values, decoded `.p12`, decoded `.p8`, or placeholder secret material are committed to git.

Users can verify an installed app with:

```sh
codesign --verify --deep --strict /Applications/icuvisor.app
spctl -a -v /Applications/icuvisor.app
```

## Hardening notes for users

- Your intervals.icu API key is stored in the OS keychain by default, not in plain text on disk. The OS account/session that can unlock your keychain is part of the trust boundary.
- `INTERVALS_ICU_API_KEY` remains available for headless servers and CI, and legacy `.env`/JSON `api_key` files remain available for compatibility, but plaintext file credentials are discouraged because they can leak through backups, shell history, or accidental commits. icuvisor warns when it uses a plaintext file-sourced API key.
- Startup diagnostics and `Config.String()` redact the API key value and report only the credential source (`env`, `keychain`, or `file`).
- The MCP HTTP transport binds to `127.0.0.1` by default. Do not expose it to a public interface unless you understand the risks.
- icuvisor only contacts `intervals.icu` and (if auto-update is enabled) `releases.icuvisor.dev`. Verify network activity against this expectation.
