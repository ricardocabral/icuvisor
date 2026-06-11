---
title: "Install on Windows"
description: "Install icuvisor on Windows from the unsigned MSI."
---

icuvisor for Windows ships as an MSI built by CI on every release. The MSI is currently **unsigned** — Windows SmartScreen may warn about an unknown publisher on first launch. The installer is functional; once past the warning, install, upgrade, and uninstall behave normally.

Winget distribution does not require Authenticode signing; it verifies the installer URL and SHA-256 hash from the package manifest. Signing may be added later if a suitable code-signing path is available.

## Install from the MSI

1. Download `icuvisor_<version>_windows_<arch>.msi` and `SHA256SUMS.txt` from the latest GitHub release. Pick `amd64` on Intel/AMD machines and `arm64` on ARM (Surface Pro X, Snapdragon laptops).
2. Optional: verify the checksum in PowerShell from the folder where both files were downloaded:

   ```powershell
   (Get-FileHash .\icuvisor_*_windows_*.msi -Algorithm SHA256).Hash.ToLower()
   Select-String -Path .\SHA256SUMS.txt -Pattern 'windows'
   ```

   The two hashes must match.

3. Double-click the MSI. When SmartScreen shows **"Windows protected your PC"**, click **More info → Run anyway**. This is expected for unsigned pre-v1 builds.
4. The installer is per-user — no UAC prompt. Files land in `%LOCALAPPDATA%\Programs\icuvisor`, which expands to `C:\Users\<you>\AppData\Local\Programs\icuvisor`, and that directory is added to your user `PATH`. `AppData` is hidden in File Explorer by default, so a normal C: drive search will not find it — paste the path into the address bar, or enable **View → Show → Hidden items**.
5. Open a **new** PowerShell or Command Prompt window (existing ones will not have the updated `PATH`) and confirm:

   ```powershell
   icuvisor version
   ```

## Before you start: API key and athlete ID

You need two things from intervals.icu:

- **API key.** Log in to [intervals.icu](https://intervals.icu), open **Settings → Developer Settings**, and copy your API key. Setup will ask for this with masked input.
- **Athlete ID.** It's displayed near the API key on the page above, or open any page on intervals.icu while logged in and look at the URL — it contains `/athlete/i12345/...`, where `i12345` is your athlete ID. Most IDs are the letter `i` followed by digits; accounts created by linking Strava have a bare-numeric ID with no `i`. Either form works.

## First-run setup

In a new shell, run:

```powershell
icuvisor setup
```

Setup asks for the intervals.icu API key (masked) and your athlete ID, verifies them against intervals.icu, stores the key in **Windows Credential Manager** under service `icuvisor` and account `intervals-icu-api-key`, autodetects your timezone, and writes only non-secret fields to the icuvisor config file. The config may include a `credential_ref` naming that Credential Manager location, but not the API key.

Useful setup flags:

| Flag                            | Use it when                                                                 |
| ------------------------------- | --------------------------------------------------------------------------- |
| `--config C:\path\to\config.json` | You want setup to write a non-default config path.                        |
| `--force`                       | You want to overwrite an existing config file without the overwrite prompt. |
| `--offline`                     | intervals.icu cannot be reached and you accept skipping verification.       |

There is no `--api-key` flag. Setup always asks for the key interactively so the key is not exposed through shell history or MCP client JSON.

Do not put the API key in Claude Desktop, Claude Code, or any committed config file.

## Configure an MCP client

Use this command path in the MCP client configuration:

```text
%LOCALAPPDATA%\Programs\icuvisor\icuvisor.exe
```

Keep the API key out of client JSON. Put only non-secret values in the client configuration, such as `INTERVALS_ICU_ATHLETE_ID`, `ICUVISOR_TIMEZONE`, `ICUVISOR_TRANSPORT`, or a `--config` path. For JSON config fields and default config paths, see the [config file reference]({{< relref "../reference/config-file" >}}); for environment variables, see the [CLI environment variable reference]({{< relref "../reference/cli#environment-variables" >}}).

## Uninstall

1. Quit any MCP clients using icuvisor.
2. **Settings → Apps → Installed apps**, find `icuvisor`, click **Uninstall**. The MSI cleans up files and the per-user `PATH` entry.
3. Optional: remove the Credential Manager API key:

   ```powershell
   cmdkey /delete:icuvisor
   ```

4. Remove any MCP client config blocks that launch icuvisor.

## Build from source

If you would rather not bypass SmartScreen, build the binary yourself with the Go toolchain. See [Build from source](https://github.com/ricardocabral/icuvisor#build-from-source) on the repository README.
