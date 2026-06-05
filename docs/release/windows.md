# Windows release: MSI, Scoop, and Winget publish setup

Companion to the macOS release runbook (kept maintainer-only at `docs/internal/release.md`). This doc covers the **Windows** half of the release pipeline: building an `.msi`, publishing to the Scoop bucket, and submitting Winget manifests.

The default Windows path does **not** require Authenticode signing. Winget community packages may use unsigned MSI/EXE installers; the trade-off is user trust and SmartScreen reputation, not package eligibility. The release workflow still has an optional Azure Trusted Signing step, but missing Azure secrets only logs a warning and ships an unsigned MSI.

## Pipeline overview

```
preflight (ubuntu)
   |
   v
release (macos)              <- builds linux/windows zips, dmg, runs goreleaser scoops
   |
   v
release-windows (windows)    <- builds .msi for amd64+arm64, optionally signs, uploads to draft
   |
   v
finalize (ubuntu)            <- regenerates SHA256SUMS, publishes draft

[release: published]
   |
   v
winget.yml                   <- submits PR to microsoft/winget-pkgs (skipped without WINGET_PAT)
```

Artifacts produced per release (added to the existing macOS/Linux set):

- `icuvisor_<version>_windows_amd64.msi` — unsigned by default; Authenticode-signed only when signing secrets are configured
- `icuvisor_<version>_windows_arm64.msi` — unsigned by default; Authenticode-signed only when signing secrets are configured
- Scoop manifest committed to `ricardocabral/scoop-icuvisor` at `bucket/icuvisor.json` (skipped without `SCOOP_BUCKET_PAT`)
- Winget PR opened against `microsoft/winget-pkgs` under `RicardoCabral.icuvisor` (skipped without `WINGET_PAT`)

## 1. GitHub Actions secrets

Only `WINGET_PAT` is required for Winget. MSI signing secrets are optional and can stay unset.

Add via the GitHub web UI (**Settings > Secrets and variables > Actions > New repository secret**) or `gh secret set`.

| Secret | Source | Gates |
| --- | --- | --- |
| `WINGET_PAT` | Classic PAT with `public_repo` scope | Winget update PRs |
| `SCOOP_BUCKET_PAT` | Fine-grained PAT with `contents: write` on `ricardocabral/scoop-icuvisor` | Scoop bucket commits |
| `AZURE_TENANT_ID` | Entra ID Directory (tenant) ID | Optional MSI signing |
| `AZURE_CLIENT_ID` | App registration Application (client) ID | Optional MSI signing |
| `AZURE_CLIENT_SECRET` | App registration client secret value | Optional MSI signing |
| `TRUSTED_SIGNING_ENDPOINT` | Trusted Signing account URI (e.g. `https://eus.codesigning.azure.net`) | Optional MSI signing |
| `TRUSTED_SIGNING_ACCOUNT` | Trusted Signing account name | Optional MSI signing |
| `TRUSTED_SIGNING_PROFILE` | Certificate profile name | Optional MSI signing |

For unsigned Winget update automation, set only:

```bash
gh secret set WINGET_PAT --repo ricardocabral/icuvisor
```

`WINGET_PAT` is not issued by Microsoft and does not require any Winget form. Create it yourself in GitHub (**Settings > Developer settings > Personal access tokens > Tokens (classic)**) with the `public_repo` scope. Use a dedicated token with an expiration date. Fine-grained PATs are not recommended here because `winget-releaser` / Komac document classic `public_repo` as the supported path.

If you also want Scoop publishing:

```bash
gh secret set SCOOP_BUCKET_PAT
```

## 2. Scoop bucket

The bucket repo `ricardocabral/scoop-icuvisor` is private at first and is **already created and seeded** with an empty `bucket/` directory. GoReleaser commits `bucket/icuvisor.json` on every tag push using `SCOOP_BUCKET_PAT`; without that secret, the scoops step is skipped silently.

When ready for public adoption:

- [ ] Flip the bucket repo to public: `gh repo edit ricardocabral/scoop-icuvisor --visibility public`.
- [ ] Smoke test: `scoop bucket add icuvisor https://github.com/ricardocabral/scoop-icuvisor && scoop install icuvisor`.

## 3. Winget submission

Package identifier: `RicardoCabral.icuvisor`.

The `.github/workflows/winget.yml` workflow uses `vedantmgoyal9/winget-releaser` to:

- detect the `.msi` artifacts attached to the published release,
- generate manifest YAMLs for the next version,
- fork `microsoft/winget-pkgs`, push a branch, and open a PR.

### 3.1 Bootstrap the first Winget version manually

`winget-releaser` expects at least one version of the package to already exist in `microsoft/winget-pkgs`. Use `wingetcreate` for the first stable version, then let the workflow handle future versions. No separate issue or Microsoft form is required; the submitted pull request is the package request.

From a Windows machine:

```powershell
winget install Microsoft.WingetCreate
wingetcreate new `
  "https://github.com/ricardocabral/icuvisor/releases/download/vX.Y.Z/icuvisor_X.Y.Z_windows_amd64.msi" `
  "https://github.com/ricardocabral/icuvisor/releases/download/vX.Y.Z/icuvisor_X.Y.Z_windows_arm64.msi"
```

Use these manifest values when prompted:

| Field | Value |
| --- | --- |
| PackageIdentifier | `RicardoCabral.icuvisor` |
| PackageName | `icuvisor` |
| Publisher | `Ricardo Niederberger Cabral` |
| PackageVersion | `X.Y.Z` (no leading `v`) |
| Moniker | `icuvisor` |
| License | `MIT` |
| PackageUrl | `https://icuvisor.app` |
| LicenseUrl | `https://github.com/ricardocabral/icuvisor/blob/main/LICENSE` |
| ShortDescription | `MCP server connecting intervals.icu training data to AI assistants.` |

Before submitting, verify the generated installer entries point at the `.msi` assets and have the correct architectures (`x64` / `arm64`) and SHA-256 hashes.

If you are not on Windows, Komac is a cross-platform alternative that can bootstrap the same first PR:

```bash
brew install komac
komac new RicardoCabral.icuvisor \
  --version X.Y.Z \
  --urls \
    "https://github.com/ricardocabral/icuvisor/releases/download/vX.Y.Z/icuvisor_X.Y.Z_windows_amd64.msi" \
    "https://github.com/ricardocabral/icuvisor/releases/download/vX.Y.Z/icuvisor_X.Y.Z_windows_arm64.msi" \
  --package-locale en-US \
  --publisher "Ricardo Niederberger Cabral" \
  --package-name icuvisor \
  --package-url https://icuvisor.app \
  --moniker icuvisor \
  --license MIT \
  --license-url https://github.com/ricardocabral/icuvisor/blob/main/LICENSE \
  --short-description "MCP server connecting intervals.icu training data to AI assistants." \
  --release-notes-url "https://github.com/ricardocabral/icuvisor/releases/tag/vX.Y.Z" \
  --submit
```

Komac also requires a classic GitHub token with `public_repo`; pass it with `--token`, set `GITHUB_TOKEN`, or run `komac token add`.

### 3.2 Winget PR review and merge

After submission, `wingetbot` runs validation and opens/updates status comments. The PR is reviewed and merged by the `microsoft/winget-pkgs` automation and package moderators/maintainers. First package submissions usually take longer than version updates.

You generally only need to monitor the PR:

- If validation fails, fix what the bot or moderator asks for.
- If a moderator asks about the unsigned MSI, reply that unsigned MSI/EXE installers are accepted in `winget-pkgs`; the manifest pins immutable GitHub release asset URLs and SHA-256 hashes.
- Once merged, users can install with `winget install RicardoCabral.icuvisor`.

### 3.3 Automate future Winget versions

After the first Winget PR is merged:

- [ ] Create a classic PAT with `public_repo` scope for the account that will open Winget PRs.
- [ ] Make sure that account can fork `microsoft/winget-pkgs` (pre-create the fork if needed).
- [ ] Add `WINGET_PAT` to this repository's Actions secrets.
- [ ] Publish a stable release such as `v1.0.1`; prereleases are intentionally skipped.
- [ ] Confirm `.github/workflows/winget.yml` opens an update PR.

The workflow strips the leading `v` from tags before passing `PackageVersion`, so `v1.0.1` becomes Winget version `1.0.1`.

## 4. WiX MSI internals

The MSI source is `build/windows/icuvisor.wxs` (WiX v4). Highlights:

- **Per-user install** (no UAC prompt). Files land in `%LOCALAPPDATA%\Programs\icuvisor`.
- **PATH update** is per-user (`System="no"`) so the binary is callable from any new shell. New shells only — existing shells need to be reopened.
- **UpgradeCode** is `8F557483-04AB-41DF-B138-3F979AF6496F` and **must not change**. Changing it would break MSI upgrade behavior for already-installed users.
- **Version** must be N.N.N (max each component 65535). The CI strips any `-SNAPSHOT`/`-rc` suffix before passing `Version=` to `wix build`. Pre-release tags like `v0.5.0-beta.1` therefore ship `0.5.0` as the MSI ProductVersion. The full semver still appears in the filename and release notes.

The `wix build` invocation passes the binary, license, and icon paths in via WiX preprocessor variables (`$(var.BinarySource)`, etc.) so the same `.wxs` works for amd64 and arm64.

## 5. Pre-flight before first unsigned Winget release

- [ ] `goreleaser check` passes locally.
- [ ] Tag a release candidate (`v1.0.1-rc.1`) and watch all jobs go green. Expect a warning in `release-windows` saying signing is disabled; the unsigned `.msi` is still uploaded to the draft. `winget.yml` will not fire for prereleases.
- [ ] Smoke test both MSIs from a Windows VM:
  - install,
  - open a new shell,
  - run `icuvisor version`,
  - uninstall via Apps & Features,
  - confirm clean removal.
- [ ] Promote to a stable tag (`v1.0.0`).
- [ ] Submit the first Winget version with `wingetcreate` or Komac (section 3.1).
- [ ] Wait for `wingetbot` validation and `microsoft/winget-pkgs` moderator review. No separate issue or form is required.
- [ ] After the first Winget PR is merged, set `WINGET_PAT` and let `winget.yml` handle future stable releases.

Windows SmartScreen may warn "Unknown publisher" on unsigned builds. That is expected and does not prevent Winget submission.

## 6. Optional Authenticode signing

Signing is useful for SmartScreen reputation, enterprise environments, and verified publisher identity, but it is not a Winget prerequisite.

The current CI signing hook is Azure Trusted Signing. If Azure Trusted Signing is not available to the maintainer, leave the Azure secrets unset. If a different signing provider is adopted later (for example, a traditional OV/EV code-signing certificate or an OSS-friendly signing service), replace the `Sign MSI with Azure Trusted Signing` step in `.github/workflows/release.yml`.

## 7. Failure modes and recovery

- **Winget first submission fails in `winget-releaser`**: the package probably is not present in `microsoft/winget-pkgs` yet. Bootstrap the first version with `wingetcreate` or Komac, wait for it to merge, then retry automation on the next release.
- **Komac cannot find `ricardocabral/winget-pkgs`**: fork `microsoft/winget-pkgs` under the token user's account, then retry the submission.
- **Winget PR never opens**: `WINGET_PAT` is missing `public_repo`, the token user cannot fork `microsoft/winget-pkgs`, or the action is rate-limited. Re-run `winget.yml` via `workflow_dispatch` with the tag.
- **Winget moderation flags unsigned installer**: note that unsigned MSI/EXE installers are allowed, then wait for static analysis/reputation if requested. If moderation asks for another fix, update the manifest or release asset as directed.
- **`wix build` fails with version error**: the tag included pre-release metadata that survived the strip step. Inspect the `Compute MSI version` step output.
- **Scoop manifest never appears in bucket repo**: `SCOOP_BUCKET_PAT` is missing `contents: write` on `scoop-icuvisor`, or expired. Regenerate.
- **MSI installs but `icuvisor` is not on PATH**: existing shells were opened before install. Open a new shell. If still missing, `setx PATH "%PATH%;%LOCALAPPDATA%\Programs\icuvisor"` and re-test.

## 8. What is _not_ wired up here

- **No NSIS installer.** Roadmap calls for `.msi` only; NSIS is left for future iteration if user demand appears.
- **No GoReleaser-Pro MSI/winget/scoops-with-msi.** Everything here is OSS GoReleaser plus a community action. If we adopt GoReleaser Pro later, the `release-windows` job becomes a thin wrapper around `goreleaser release --split`.
- **No code signing of the raw `icuvisor.exe`** inside the MSI. If SmartScreen reputation requires per-binary signing later, sign the `.exe` in the same `release-windows` job before `wix build`.
