# TP-037-macos-signed-installer: macOS signed installer + manual client config docs — Status

**Current Step:** Step 1: Apple Developer setup + bundle identity
**Status:** 🟡 In Progress
**Last Updated:** 2026-05-15
**Review Level:** 3
**Review Counter:** 2
**Iteration:** 1
**Size:** M

---

### Step 1: Apple Developer setup + bundle identity

**Status:** 🟨 In Progress

- [ ] Developer ID cert enrolled, `.p12` exportable for CI
- [x] Bundle identifier locked (proposed `dev.icuvisor.icuvisor`)
- [x] App-as-headless-in-.app, LSUIElement=true
- [x] R001 plan: record only non-secret Apple signing metadata requirements/TBDs and secret-handling boundaries
- [x] R001 plan: confirm final bundle identifier rationale and TP-036 keychain interaction
- [x] R001 plan: document headless `.app` launch behavior and Info.plist version-substitution plan

### Step 2: GoReleaser DMG + signing

**Status:** ⏳ Not started

- [ ] Universal-2 binary (amd64 + arm64 via lipo)
- [ ] `codesign --options runtime --timestamp`
- [ ] DMG packaging
- [ ] Notarize + staple

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

## Blockers

- Step 1 Developer ID certificate verification is blocked on maintainer-owned Apple Developer/GitHub setup. Local `security find-identity -v -p codesigning` returned `0 valid identities found`; no Apple Team ID, Developer ID Application common name, certificate expiration, or GitHub secret presence is available in the repository. Per R002, do not check the Developer ID certificate item complete until those non-secret facts are supplied.

## Notes

_Add notes as work progresses._

- R001 plan review requested explicit Step 1 planning before implementation; suggested later Step 2/3 attention to existing Homebrew publishing configuration.
- R002 plan review approved implementation but requires real Apple Team ID, Developer ID Application common name, certificate expiration date, and GitHub secret presence before checking the Developer ID certificate item complete.
- Step 1 plist validation: `plutil -lint build/macos/Info.plist` passed.

| 2026-05-15 17:43 | Task started | Runtime V2 lane-runner execution |
| 2026-05-15 17:43 | Step 1 started | Apple Developer setup + bundle identity |
| 2026-05-15 17:46 | Review R001 | plan Step 1: REVISE |
| 2026-05-15 17:49 | Review R002 | plan Step 1: APPROVE |
