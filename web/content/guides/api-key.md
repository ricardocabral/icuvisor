---
title: "Get and store an API key"
description: "Create an intervals.icu API key and store it safely for icuvisor."
---

icuvisor needs an intervals.icu API key to read your training data. The key is a credential, not a chat prompt: do not paste it into Claude, ChatGPT, or any MCP client JSON.

## 1. Create the key in intervals.icu

1. Open <https://intervals.icu/settings>.
2. Create or copy an API key.
3. Keep the browser tab open until setup has stored the key.

## 2. Store it with `icuvisor setup`

Run setup from your installed binary:

```bash
/Applications/icuvisor.app/Contents/MacOS/icuvisor setup
```

If you built from source, run:

```bash
./bin/icuvisor setup
```

Setup asks for the API key with masked input and your athlete ID, verifies the pair against intervals.icu, stores the key in the OS keychain, autodetects your timezone, and writes only non-secret config fields. The generated config may include a `credential_ref` that names the keychain service/account (`icuvisor` / `intervals-icu-api-key`), but it never contains the API key itself. Athlete IDs are usually the letter `i` followed by digits (for example `i12345`), but accounts created by linking Strava have a bare-numeric ID with no `i` — both forms are accepted. intervals.icu does not expose the ID via the API, so it cannot be autodetected.

Useful setup flags are documented in the [CLI reference]({{< relref "../reference/cli" >}}):

- `--config /path/to/config.json` writes a specific non-secret config file.
- `--force` overwrites an existing config file without the overwrite prompt.
- `--offline` skips intervals.icu verification when you are offline and accepts a manual timezone prompt.

## 3. Configure your MCP client with non-secrets only

MCP client JSON should contain only values such as `INTERVALS_ICU_ATHLETE_ID`, `ICUVISOR_TIMEZONE`, `ICUVISOR_TRANSPORT`, or a `--config` path. It is also safe to point clients at the generated config file: its `credential_ref` is only keychain metadata. The API key stays in the OS keychain.

## Manual keychain storage

Manual storage is useful for advanced or headless setup, but `icuvisor setup` is the recommended path.

### macOS Keychain

Use Keychain Access, or run:

```bash
security add-generic-password -U \
  -s icuvisor \
  -a intervals-icu-api-key \
  -w 'YOUR_INTERVALS_ICU_API_KEY'
```

### Windows Credential Manager

Open Credential Manager and add a Windows credential with:

- Internet/network address: `icuvisor:intervals-icu-api-key`
- User name: `intervals-icu-api-key`
- Password: your intervals.icu API key

CLI equivalent:

```powershell
cmdkey /add:icuvisor:intervals-icu-api-key /user:intervals-icu-api-key /pass:YOUR_INTERVALS_ICU_API_KEY
```

### Linux libsecret

On desktop Linux with Secret Service support, use Passwords and Keys, KWallet, or:

```bash
secret-tool store --label='icuvisor intervals.icu API key' service icuvisor username intervals-icu-api-key
# Paste the intervals.icu API key when prompted.
```

If `secret-tool` is missing, install your distribution's libsecret package. Headless systems without a keychain can use `INTERVALS_ICU_API_KEY` as a deliberate fallback; remember that process environments are easier to leak through shell history, process listings, logs, and service files.

## Credential precedence

When icuvisor starts, the API key comes from the highest available source:

1. Process environment `INTERVALS_ICU_API_KEY`.
2. OS keychain service `icuvisor`, account `intervals-icu-api-key`.
3. Legacy plaintext `.env` or JSON config fallback.

Plaintext keys emit a warning and should not be committed or backed up.
