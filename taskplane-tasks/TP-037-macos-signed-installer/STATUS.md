# TP-037-macos-signed-installer: macOS signed installer + manual client config docs — Status

**Current Step:** Step 1: Apple Developer setup + bundle identity
**Status:** ⏳ Not started
**Last Updated:** 2026-05-15
**Review Level:** 3
**Review Counter:** 0
**Iteration:** 0
**Size:** M

---

### Step 1: Apple Developer setup + bundle identity

**Status:** ⏳ Not started

- [ ] Developer ID cert enrolled, `.p12` exportable for CI
- [ ] Bundle identifier locked (proposed `dev.icuvisor.icuvisor`)
- [ ] App-as-headless-in-.app, LSUIElement=true

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

- **Bundle identifier:** TBD (proposed `dev.icuvisor.icuvisor`); record in Step 1.
- **Cross-platform installers:** explicitly deferred to v1.0; v0.5 is macOS-only.

## Notes

_Add notes as work progresses._
