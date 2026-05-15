# macOS install and release-operator checklist

icuvisor v0.5 ships a signed, notarized DMG for macOS. The app is a headless `.app` wrapper around the MCP stdio binary; MCP clients execute:

```text
/Applications/icuvisor.app/Contents/MacOS/icuvisor
```

The app does not contain credentials. Your intervals.icu API key stays in the macOS Keychain.

## Install from the DMG

1. Download `icuvisor_<version>_macos_universal.dmg` and `SHA256SUMS.txt` from the GitHub release.
2. Optional checksum verification:
   ```bash
   shasum -a 256 -c SHA256SUMS.txt --ignore-missing
   ```
3. Open the DMG.
4. Drag `icuvisor.app` to `/Applications` or `~/Applications`.
5. On first launch, open it from Finder or run the binary once:
   ```bash
   /Applications/icuvisor.app/Contents/MacOS/icuvisor version
   ```

A properly signed and notarized release should not show the macOS "unidentified developer" warning. If macOS blocks the app, stop and verify the signature before overriding Gatekeeper.

## First-run setup

Run the terminal setup flow after installing:

```bash
/Applications/icuvisor.app/Contents/MacOS/icuvisor setup
```

Setup asks for the intervals.icu API key with masked input, verifies it, stores it in Keychain under service `icuvisor` and account `intervals-icu-api-key`, autodetects your athlete ID/timezone, and writes only non-secret fields to the icuvisor config file. Use `--config /path/to/config.json` for a non-default config path, `--force` to overwrite an existing config file without the prompt, or `--offline` only when intervals.icu cannot be reached and you accept skipping verification.

Manual Keychain storage is still available for advanced/headless setups:

```bash
security add-generic-password -U \
  -s icuvisor \
  -a intervals-icu-api-key \
  -w 'YOUR_INTERVALS_ICU_API_KEY'
```

Do not put the API key in Claude Desktop, Claude Code, `Info.plist`, the DMG, or any committed config file.

## Verify an installed release

After dragging the app into Applications, run:

```bash
codesign --verify --deep --strict /Applications/icuvisor.app
spctl -a -v /Applications/icuvisor.app
xcrun stapler validate /path/to/icuvisor_<version>_macos_universal.dmg
/Applications/icuvisor.app/Contents/MacOS/icuvisor version
```

Expected:

- `codesign` exits 0.
- `spctl` reports the app is accepted and references the Developer ID authority.
- `stapler validate` reports the notarization ticket is valid.
- `icuvisor version` prints the release version without asking for an API key.

## Configure an MCP client

- Claude Desktop: [docs/clients/claude-desktop.md](../clients/claude-desktop.md)
- Claude Code: [docs/clients/claude-code.md](../clients/claude-code.md)

Both configs should use `/Applications/icuvisor.app/Contents/MacOS/icuvisor` as the command and should not contain the API key.

## Optional LaunchAgent for power users

v0.5 does not auto-load a LaunchAgent. Most MCP clients start icuvisor on demand over stdio. If you later add a LaunchAgent for a local workflow, keep it user-scoped, review it before loading, and do not store API keys in the plist.

## Uninstall

1. Quit any MCP clients using icuvisor.
2. Remove the app:
   ```bash
   rm -rf /Applications/icuvisor.app
   ```
3. Optional: remove the Keychain API key:
   ```bash
   security delete-generic-password -s icuvisor -a intervals-icu-api-key
   ```
4. Remove any Claude Desktop or Claude Code `mcpServers.icuvisor` config blocks.

## Release operator preflight

Live signing and notarization require Apple assets that are intentionally not stored in git. Before creating a real release, the operator must complete this checklist:

1. Enroll the maintainer account or organization in the Apple Developer Program.
2. Create a **Developer ID Application** certificate in Apple Developer Certificates, Identifiers & Profiles.
3. Install the certificate in Keychain Access on a trusted Mac, then export the certificate and private key as a password-protected `.p12`.
4. Base64-encode the `.p12` for GitHub Actions:
   ```bash
   base64 -i DeveloperIDApplication.p12 | pbcopy
   ```
5. Create an App Store Connect API key for `xcrun notarytool`, download the `.p8` once, and base64-encode it:
   ```bash
   base64 -i AuthKey_XXXXXXXXXX.p8 | pbcopy
   ```
6. Record only non-secret release metadata: Apple Team ID, Developer ID Application common name, and certificate expiration date.
7. Add these GitHub Actions secrets by name only:
   - `APPLE_TEAM_ID`
   - `APPLE_DEVELOPER_ID_P12_BASE64`
   - `APPLE_DEVELOPER_ID_P12_PASSWORD`
   - `APPLE_API_KEY_ID`
   - `APPLE_API_KEY_ISSUER`
   - `APPLE_API_KEY_BASE64`
8. Confirm the signing identity on the release runner/keychain:
   ```bash
   security find-identity -v -p codesigning
   ```
   The output must include a valid `Developer ID Application` identity.
9. Push a `v*.*.*` tag to run the release workflow.
10. Download the DMG from the draft/published release and verify it with the commands in [Verify an installed release](#verify-an-installed-release).

Do not commit the `.p12`, `.p8`, passwords, API keys, decoded secret material, or placeholder secret files.
