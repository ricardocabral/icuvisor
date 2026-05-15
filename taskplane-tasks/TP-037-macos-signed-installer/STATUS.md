# TP-037-macos-signed-installer: macOS signed installer + manual client config docs — Status

**Current Step:** Step 2: GoReleaser DMG + signing
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 3
**Review Counter:** 8
**Iteration:** 2
**Size:** M

---

### Step 1: Apple Developer setup + bundle identity

**Status:** ✅ Complete

**Operator-deferred release preflight (not locally satisfied):** the operator confirmed they do not have Apple metadata or keys now. Before any real signed/notarized macOS release is cut, the maintainer/release operator must enroll in the Apple Developer Program, create/export a Developer ID Application certificate as a password-protected `.p12` from Keychain Access, create an App Store Connect API key for `xcrun notarytool`, record non-secret Apple Team ID / Developer ID Application common name / certificate expiration date, add GitHub Actions secrets by name only, run a `v*` tag release, and verify with `security find-identity -v -p codesigning`, `codesign --verify --deep --strict /Applications/icuvisor.app`, `spctl -a -v /Applications/icuvisor.app`, and `xcrun stapler validate /path/to/icuvisor_*.dmg`. Local evidence remains `security find-identity -v -p codesigning` => `0 valid identities found`; this task only documents/scaffolds the gate and does not claim a live signed release was produced.

- [x] Maintainer preflight documented; live Developer ID certificate enrollment and `.p12` export validation deferred to release operator
- [x] Bundle identifier locked (proposed `dev.icuvisor.icuvisor`)
- [x] App-as-headless-in-.app, LSUIElement=true
- [x] R001 plan: record only non-secret Apple signing metadata requirements/TBDs and secret-handling boundaries
- [x] R001 plan: confirm final bundle identifier rationale and TP-036 keychain interaction
- [x] R001 plan: document headless `.app` launch behavior and Info.plist version-substitution plan
- [x] R003 code: make deferred Developer ID validation visible as an operator release gate and clean STATUS.md notes/newline

### Step 2: GoReleaser DMG + signing

**Status:** 🟨 In Progress

- [ ] Universal-2 binary configured for darwin amd64 + arm64 via GoReleaser `universal_binaries`
- [ ] Apple-native signing scaffold uses `codesign --options runtime --timestamp` with hard preflight gates
- [ ] DMG packaging scaffold uses `hdiutil` with minimal `.app` + Applications symlink layout
- [ ] Notarization/stapling scaffold uses `xcrun notarytool submit --wait` and `xcrun stapler`, with live execution operator-deferred until Apple credentials exist
- [ ] R005/R006 plan: remove the GoReleaser `brews` block and `HOMEBREW_TAP_GITHUB_TOKEN` from this tag-release path
- [ ] R005/R006 plan: suppress standalone darwin archives so the macOS release artifact is the signed/notarized DMG only, while Linux/Windows archives stay unchanged
- [ ] R005/R006/R007 plan: split GoReleaser builds so archives include only Linux/Windows build IDs; darwin builds feed `universal_binaries` for the DMG script but produce no standalone darwin archive
- [ ] R005/R006/R007 plan: add `build/macos/package_dmg.sh` for plist substitution, `.app` assembly, Developer ID app signing/verification, `hdiutil`, DMG container signing/verification, `notarytool`, stapling, and validation
- [ ] R005/R006/R007/R008 plan: keep GoReleaser publishing to a draft release only, run explicit GitHub Actions steps to package/upload the DMG and clobber the final `SHA256SUMS.txt`, then publish the draft only after all artifacts verify/upload successfully
- [ ] R005/R006 plan: use `ICUVISOR_MACOS_RELEASE=1` as the fail-closed tag-release switch requiring Developer ID identity plus `APPLE_TEAM_ID`, `APPLE_API_KEY_ID`, `APPLE_API_KEY_ISSUER`, and decoded API key path; snapshots omit the flag and build unsigned with warnings
- [ ] R005/R006 plan: validate with `goreleaser check`, local snapshot/dry-run app+DMG assembly, and a release-mode negative test proving missing Apple prerequisites fail closed on macOS-only tooling

### Step 3: Release workflow

**Status:** ⏳ Not started

- [ ] GH Actions job on tag push
- [ ] All Apple secrets via GH Actions secrets (named in SECURITY.md)
- [ ] DMG + SHA256SUMS uploaded to release

### Step 4: Manual client config docs

**Status:** ⏳ Not started

- [ ] `docs/clients/claude-desktop.md`
- [ ] `docs/clients/claude-code.md`
- [ ] Both: "API key in keychain, not in JSON" callout
- [ ] Verify-the-connection recipe

### Step 5: Verification

**Status:** ⏳ Not started

- [ ] Clean-account drag-install passes Gatekeeper
- [ ] `codesign --verify` + `spctl -a` pass
- [ ] Claude Desktop config-and-prompt round trip
- [ ] No plaintext key on disk after first run

---

## Decisions

- **Apple signing metadata:** do not fabricate Apple Developer facts. The maintainer must provision a Developer ID Application certificate and record only non-secret metadata here before a real release: Apple Team ID, Developer ID Application common name, certificate expiration date, and whether the `.p12` has been stored as `APPLE_DEVELOPER_ID_P12_BASE64` with password in `APPLE_DEVELOPER_ID_P12_PASSWORD`. No `.p12`, `.p8`, app-specific password, API key, or secret value belongs in git or STATUS.md.
- **Bundle identifier:** locked as `dev.icuvisor.icuvisor`. Rationale: maintainers assert the project controls or is authorized to use the `icuvisor.dev` namespace, and this reverse-DNS identifier is treated as permanent for macOS trust and any future keychain access-control prompts.
- **TP-036 keychain interaction:** current credential namespace is service `icuvisor` and account `intervals-icu-api-key` (`internal/credstore`). The bundle identifier does not change those lookup strings, so no data migration is planned; users upgrading from an unsigned/manual binary may still see a macOS Keychain access prompt because the app's designated requirement changes to the signed Developer ID app.
- **macOS app launch model:** v0.5 ships a headless `.app` wrapper with `LSUIElement=true`; MCP clients execute `/Applications/icuvisor.app/Contents/MacOS/icuvisor` directly over stdio. Finder double-click/open is permitted for Gatekeeper/keychain trust but may exit or run without visible UI; no tray/menu-bar app is shipped. LaunchAgent support is optional documentation only and must not be auto-loaded by the installer.
- **Info.plist plan:** `build/macos/Info.plist` will carry `CFBundleIdentifier=dev.icuvisor.icuvisor`, `CFBundleExecutable=icuvisor`, `CFBundleName=icuvisor`, `CFBundlePackageType=APPL`, `LSUIElement=true`, and placeholder `CFBundleShortVersionString`/`CFBundleVersion` values that release packaging substitutes from GoReleaser instead of hard-coding per release.
- **Cross-platform installers:** explicitly deferred to v1.0; v0.5 is macOS-only.
- **Step 2 release-shaping decisions:** remove Homebrew auto-publishing from `.goreleaser.yaml`/`.github/workflows/release.yml` for TP-037; split GoReleaser builds so Linux/Windows archives remain unchanged but darwin builds only feed `universal_binaries`; do not publish standalone darwin archives because only the DMG receives the release signing/notarization gate; set `.goreleaser.yaml` checksum output to `SHA256SUMS.txt` and publish to a draft release first; implement macOS packaging in `build/macos/package_dmg.sh`; run DMG creation/upload/checksum as explicit GitHub Actions steps after GoReleaser so `gh release upload --clobber` attaches the DMG and replaces the final `SHA256SUMS.txt`; publish the draft only after signing/notarization/stapling, checksum regeneration, and uploads succeed, so failed DMG packaging never exposes an incomplete public release; sign the `.app` with `codesign --options runtime --timestamp`, sign the DMG container with `codesign --timestamp`, verify both before notarization, then submit/staple/validate; use `ICUVISOR_MACOS_RELEASE=1` in the tag release workflow to require Apple credentials and fail closed, while local snapshots without that flag may create an explicitly unsigned DMG scaffold for dry-run validation.

## Blockers

- Operator-deferred Apple Developer preflight (not a repository implementation blocker): operator confirmed they do not have Apple metadata or keys now. Local `security find-identity -v -p codesigning` => `0 valid identities found`. Required before a live signed/notarized release: enroll in Apple Developer Program, create/export Developer ID Application `.p12`, create App Store Connect API key for `notarytool`, record non-secret Team ID / Developer ID Application CN / cert expiry, add GitHub Actions secret presence by name only for `APPLE_TEAM_ID`, `APPLE_DEVELOPER_ID_P12_BASE64`, `APPLE_DEVELOPER_ID_P12_PASSWORD`, `APPLE_API_KEY_ID`, `APPLE_API_KEY_ISSUER`, and `APPLE_API_KEY_BASE64`, run a `v*` tag release, and verify DMG/app with `security find-identity -v -p codesigning`, `codesign --verify --deep --strict`, `spctl -a -v`, and `xcrun stapler validate`.

## Notes

_Add notes as work progresses._

- R001 plan review requested explicit Step 1 planning before implementation; suggested later Step 2/3 attention to existing Homebrew publishing configuration.
- R002 plan review approved implementation but required real Apple Team ID, Developer ID Application common name, certificate expiration date, and GitHub secret presence before checking the original Developer ID certificate item complete; supervisor steering later converted that live validation into an explicit operator release gate so repository scaffolding can continue without fabricated Apple Developer facts.
- Step 1 plist validation: `plutil -lint build/macos/Info.plist` passed.
- Steering accepted for Step 1: live Apple Developer certificate validation is maintainer-owned and deferred to release operator; SECURITY.md now carries the hard preflight gate/checklist and no secret or placeholder secret material is committed.
- Supervisor confirmed Apple metadata/keys are unavailable now; proceed with scaffoldable implementation and mark live cert/notarization/Gatekeeper checks as operator-deferred release preflight with exact commands.
- R003 suggestion: keep the authoritative non-secret release-operator record in STATUS.md while SECURITY.md documents the reusable gate.

| 2026-05-15 17:43 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 17:43 | Step 1 started | Apple Developer setup + bundle identity |
| 2026-05-15 17:46 | Review R001 | plan Step 1: REVISE |
| 2026-05-15 17:49 | Review R002 | plan Step 1: APPROVE |

| 2026-05-15 17:52 | Agent escalate | TP-037 Step 1 is blocked on maintainer-owned Apple Developer setup. The plan reviewer explicitly said not to mark `Developer ID cert enrolled, .p12 exportable for CI` complete until real non-secret metadata and GitHub secret-presence evidence are supplied. |
| 2026-05-15 17:52 | Worker iter 1 | done in 536s, tools: 47 |
| 2026-05-15 17:56 | Review R003 | code Step 1: REVISE |
| 2026-05-15 17:59 | Review R004 | code Step 1: APPROVE |
| 2026-05-15 18:03 | Review R005 | plan Step 2: REVISE |
| 2026-05-15 18:05 | Review R006 | plan Step 2: REVISE |
| 2026-05-15 18:10 | Review R007 | plan Step 2: REVISE |
| 2026-05-15 18:13 | Review R008 | plan Step 2: REVISE |
