# TP-037 — macOS signed installer + manual Claude Desktop / Claude Code config docs

## Mission

Ship a signed, notarized macOS `.dmg` for `icuvisor` so the v0.5 internal-beta cohort (5–10 forum-recruited athletes, ROADMAP v0.5) can install without a terminal. The installer drops `icuvisor` into `/Applications/` (or `~/Applications/`) and registers a launchd plist that runs it under the user's login session for stdio clients. Pair the installer with copy-pasteable manual config snippets for Claude Desktop and Claude Code so a non-developer can wire it up. Full multi-platform installers, the Onboarding UI, and DXT bundles are out of scope here — those land in v1.0 (TP track to be created later).

The v0.5 milestone explicitly says "macOS signed installer; manual Claude Desktop / Claude Code config documentation" — not full cross-platform parity, not auto-update. Keep this task tight.

PRD anchors: §7.2.A Distribution; §7.4 #1 ("auto-update via signed releases is acceptable to athletes and to the macOS/Windows platforms"); KR1 (install success). PRD positions DMG + notarization as v1.0-grade; for v0.5 we accept a lower polish bar (no fancy DMG background, no welcome PDF) but the *signing/notarization chain* must work end-to-end because that's the risky bit.

ROADMAP positioning: v0.5 — Internal beta. Depends on TP-036 (keychain) so the installed binary uses the right credential path on first launch. Independent of TP-038 (onboarding) — the v0.5 installer ships without an onboarding UI; the README walks the athlete through `secret-tool` / Keychain Access manually. Onboarding UI is v1.0.

Complexity: Blast radius 2 (release pipeline only — no runtime code paths change), Pattern novelty 3 (Apple notarization is a known sharp edge), Security 3 (signing keys handled in CI), Reversibility 2 = 10 → Review Level 3. Size: M.

## Dependencies

- **TP-036** — keychain credential storage. Installer-installed athletes are the cohort that most benefits from "API key not on disk"; shipping the installer before keychain support would lock the cohort into the legacy plaintext path.

## Context to Read First

- `CLAUDE.md` "Build, test, release" — current `make snapshot` GoReleaser flow.
- `docs/prd/PRD-icuvisor.md` §7.2.A, §7.4 #1, §4 KR1.
- `ROADMAP.md` v0.5 and v1.0 — note what is *deferred* to v1.0 (DXT, Homebrew tap, cross-platform installers, auto-update); v0.5 ships *only* macOS DMG + manual config.
- Any existing `.goreleaser.yml` or release workflow under `.github/workflows/`.
- `README.md` Quickstart — the installer flow has to dovetail with the existing manual-install path, not replace it.

## File Scope

Expected files:

- `.goreleaser.yml` — add a `dmg` artifact and the macOS signing/notarization configuration. Universal binary (`arm64` + `amd64`) per PRD §7.2.A.
- `.github/workflows/release.yml` (or extend the existing one) — wire Apple Developer ID cert, app-specific password, notarytool, and `gon` (or equivalent). Secrets stored in GitHub Actions secrets; never echoed.
- `build/macos/Info.plist` and any LSUIElement / bundle-identifier metadata. Bundle identifier: `dev.icuvisor.icuvisor` (placeholder — confirm in Step 1).
- `build/macos/launchd/dev.icuvisor.plist` — optional LaunchAgent for power users who want the server resident; document but do not auto-load.
- `docs/install/macos.md` — installer install path, Gatekeeper / first-launch right-click flow, how to verify the signature with `codesign --verify --deep --strict`, and how to uninstall.
- `docs/clients/claude-desktop.md` and `docs/clients/claude-code.md` — copy-pasteable JSON snippets for the `mcpServers` block, with absolute paths for `/Applications/icuvisor.app/Contents/MacOS/icuvisor` and the env-var lines for `INTERVALS_ICU_ATHLETE_ID` etc. The API key is NOT in the JSON — it's in the keychain. Make that explicit.
- `README.md` — Quickstart adds a "Download for macOS" section above the existing build-from-source instructions.
- `CHANGELOG.md`.
- `SECURITY.md` — note the signing identity and how to verify a release.
- `taskplane-tasks/TP-037-macos-signed-installer/STATUS.md`.

Out of scope: Windows `.msi`, Linux `.deb`/`.rpm`, Homebrew tap, Scoop, Winget, DXT, auto-update, onboarding UI, custom DMG background art. Open follow-up tasks for those at the end of v0.5 if they look like blockers for v1.0.

## Steps

### Step 1: Apple Developer setup + bundle identity

- [ ] Confirm the Developer ID (Application) certificate is enrolled and exportable as a `.p12` for CI. Record the chosen team identifier in `STATUS.md` (not the value of the password — that lives in GH Actions secrets).
- [ ] Lock the bundle identifier (`dev.icuvisor.icuvisor` proposed). Once chosen it is effectively permanent — changing it later breaks user-saved keychain ACLs.
- [ ] Decide app-as-tray-app vs. headless binary in `.app`. v0.5 default: headless binary inside `.app` wrapper for code-signing; no tray icon (that's v1.0). LSUIElement=true so it does not show in the Dock if launched manually.

### Step 2: GoReleaser DMG + signing

- [ ] Universal-2 binary build (amd64 + arm64 via `lipo`).
- [ ] `codesign --options runtime --timestamp` with the Developer ID cert.
- [ ] DMG packaging with a minimal layout (Applications symlink + `.app`); no custom background art.
- [ ] Notarize via `notarytool submit --wait`; staple ticket to the DMG; fail the release if any step returns non-zero.

### Step 3: Release workflow

- [ ] GitHub Actions job runs on tag push (`v*`); reuses the existing test/lint preflight.
- [ ] Secrets handled via `secrets.APPLE_DEVELOPER_ID_P12_BASE64`, `APPLE_DEVELOPER_ID_P12_PASSWORD`, `APPLE_API_KEY_ID`, `APPLE_API_KEY_ISSUER`, `APPLE_API_KEY_BASE64`. Document each in `SECURITY.md` (names only).
- [ ] Artifact uploaded to the GitHub release. Include `SHA256SUMS.txt` (signed/notarized DMG sha) so power users can verify.
- [ ] Workflow does **not** auto-publish to a Homebrew tap or any other channel in this task.

### Step 4: Manual client config docs

- [ ] `docs/clients/claude-desktop.md` — exact `claude_desktop_config.json` `mcpServers.icuvisor` block, with absolute binary path, `INTERVALS_ICU_ATHLETE_ID` env var, optional `ICUVISOR_TRANSPORT=stdio`. Explicit note: "Your API key is read from the macOS Keychain. Do not put it in this file."
- [ ] `docs/clients/claude-code.md` — equivalent for Claude Code's `.mcp.json` / project config.
- [ ] Each doc ends with a "Verify the connection" recipe: open the client, ask "What's my FTP?", expect a populated answer; if not, link to the diagnostics section.

### Step 5: Verification

- [ ] On a fresh macOS user account (or test VM), download the DMG from a draft release, open, drag to Applications, double-click — Gatekeeper accepts without the "unidentified developer" dialog.
- [ ] `codesign --verify --deep --strict /Applications/icuvisor.app && spctl -a -v /Applications/icuvisor.app` both pass.
- [ ] Wire Claude Desktop manually using the new doc; ask the test athlete prompt; tool call succeeds.
- [ ] Confirm zero plaintext API key on disk after a successful first run (TP-036 keychain path).

## Acceptance Criteria

- A tagged release produces a signed, notarized, stapled `.dmg` containing a universal-2 `icuvisor.app`.
- A clean macOS user with no developer tooling can drag-install and launch without Gatekeeper warnings.
- Claude Desktop and Claude Code each have a copy-pasteable config doc that works against the installed `.app`.
- The installed binary reads its API key from the keychain (not from the config file) on the happy path.
- Release workflow runs on tag push; secrets never appear in logs; an unsigned build cannot be promoted.
- `README.md` Quickstart links to the macOS install doc above the build-from-source instructions.

## Do NOT

- Do not bake the API key into the `.app`, the Info.plist, or any defaults. The bundle ships *empty of credentials*.
- Do not ship Windows / Linux installers in this task. v0.5 is macOS-only; cross-platform parity is v1.0.
- Do not ship auto-update in this task — that's v1.0. Users update by re-downloading.
- Do not commit the `.p12`, the `.p8` App Store Connect key, the app-specific password, or any other Apple secret. Use GH Actions secrets exclusively.
- Do not change the bundle identifier between releases; it's effectively load-bearing for keychain ACLs and user trust.
- Do not retag a broken release. Cut a new patch tag (CLAUDE.md "Build, test, release").

## Documentation

Must update:

- `STATUS.md`
- `docs/install/macos.md` (new)
- `docs/clients/claude-desktop.md` (new)
- `docs/clients/claude-code.md` (new)
- `README.md` Quickstart
- `CHANGELOG.md`
- `SECURITY.md` (signing identity + verification recipe)

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-037`, for example: `TP-037 add notarized DMG packaging to goreleaser`.

---

## Amendments

_Add amendments below this line only._
