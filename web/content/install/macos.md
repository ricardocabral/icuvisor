---
title: "Install on macOS"
description: "Install the signed icuvisor macOS app and verify Gatekeeper status."
---

icuvisor for macOS is distributed as a signed, notarized DMG. The app is a headless `.app` wrapper around the MCP server binary; AI clients run the binary inside the app bundle.

```text
/Applications/icuvisor.app/Contents/MacOS/icuvisor
```

The app does not contain credentials. Your intervals.icu API key stays in the macOS Keychain.

## Install from the DMG

1. Download `icuvisor_<version>_macos_universal.dmg` and `SHA256SUMS.txt` from the latest GitHub release.
2. Optional: verify the checksum from the folder where both files were downloaded:

   ```bash
   shasum -a 256 -c SHA256SUMS.txt --ignore-missing
   ```

3. Open the DMG.
4. Drag `icuvisor.app` to `/Applications` or `~/Applications`.
5. Run the binary once to confirm it starts:

   ```bash
   /Applications/icuvisor.app/Contents/MacOS/icuvisor version
   ```

A properly signed and notarized release should not show the macOS "unidentified developer" warning. If macOS blocks the app, verify the signature before overriding security warnings.

## Before you start: API key and athlete ID

You need two things from intervals.icu:

- **API key.** Log in to [intervals.icu](https://intervals.icu), open **Settings → Developer Settings**, and copy your API key. Setup will ask for this with masked input.
- **Athlete ID.** It's displayed near the API key on the page above, or open any page on intervals.icu while logged in and look at the URL — it contains `/athlete/i12345/...`, where `i12345` is your athlete ID. intervals.icu IDs always start with the letter `i` followed by digits.

## First-run setup

After installing, run:

```bash
/Applications/icuvisor.app/Contents/MacOS/icuvisor setup
```

Setup asks for the intervals.icu API key (masked) and your athlete ID, verifies them against intervals.icu, stores the key in Keychain under service `icuvisor` and account `intervals-icu-api-key`, autodetects your timezone, and writes only non-secret fields to the icuvisor config file. The config may include a `credential_ref` naming that Keychain location, but not the API key.

Useful setup flags:

| Flag                            | Use it when                                                                 |
| ------------------------------- | --------------------------------------------------------------------------- |
| `--config /path/to/config.json` | You want setup to write a non-default config path.                          |
| `--force`                       | You want to overwrite an existing config file without the overwrite prompt. |
| `--offline`                     | intervals.icu cannot be reached and you accept skipping verification.       |

There is no `--api-key` flag. Setup always asks for the key interactively so the key is not exposed through shell history or MCP client JSON.

Manual Keychain storage remains available for advanced or headless setups:

```bash
security add-generic-password -U \
  -s icuvisor \
  -a intervals-icu-api-key \
  -w 'YOUR_INTERVALS_ICU_API_KEY'
```

Do not put the API key in Claude Desktop, Claude Code, `Info.plist`, the DMG, or any committed config file.

## Verify Gatekeeper and notarization (optional)

These checks are optional. macOS already enforces Gatekeeper and notarization when you first open the app — if the app launched without an "unidentified developer" warning, the signature and notarization ticket are valid. Run the commands below only if you want to confirm the chain manually or you are auditing the release.

After dragging the app into Applications, run:

```bash
codesign --verify --deep --strict /Applications/icuvisor.app
spctl -a -v /Applications/icuvisor.app
xcrun stapler validate /path/to/icuvisor_<version>_macos_universal.dmg
/Applications/icuvisor.app/Contents/MacOS/icuvisor version
```

Expected results:

- `codesign` exits 0.
- `spctl` reports the app is accepted and references the Developer ID authority.
- `stapler validate` reports the notarization ticket is valid.
- `icuvisor version` prints the release version without asking for an API key.

## Configure an MCP client

Use this command path in the MCP client configuration:

```text
/Applications/icuvisor.app/Contents/MacOS/icuvisor
```

Keep the API key out of client JSON. Put only non-secret values in the client configuration, such as `INTERVALS_ICU_ATHLETE_ID`, `ICUVISOR_TIMEZONE`, `ICUVISOR_TRANSPORT`, or a `--config` path. The full list is in the [CLI reference]({{< relref "../reference/cli" >}}).

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

4. Remove any MCP client config blocks that launch icuvisor.

## Optional LaunchAgent for power users

icuvisor does not auto-load a LaunchAgent. Most MCP clients start icuvisor on demand over stdio. If you later add a LaunchAgent for a local workflow, keep it user-scoped, review it before loading, and do not store API keys in the plist.
