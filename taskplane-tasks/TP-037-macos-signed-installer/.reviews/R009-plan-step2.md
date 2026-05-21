# Plan review — TP-037 Step 2

Decision: **approved**

The Step 2 plan in `STATUS.md` now addresses the release-signing risks called out in the prior reviews and is concrete enough to implement.

## What is now sufficiently specified

- Homebrew auto-publishing will be removed from this tag-release path, including the existing `brews` block and `HOMEBREW_TAP_GITHUB_TOKEN` usage.
- Standalone darwin archives will be suppressed by splitting GoReleaser build IDs so Linux/Windows archives remain unchanged while darwin builds only feed the universal binary used by the DMG script.
- `build/macos/package_dmg.sh` is named as the packaging point and has a clear responsibility set: Info.plist substitution, `.app` assembly, app signing/verification, DMG creation, DMG signing/verification, notarization, stapling, and validation.
- Release-mode behavior is fail-closed behind `ICUVISOR_MACOS_RELEASE=1` and requires the Developer ID identity plus Apple notary prerequisites; local snapshot behavior is explicitly unsigned/scaffold-only.
- The publication gate is now acceptable: GoReleaser publishes only to a draft release, the workflow packages/verifies/uploads the DMG and replaces the final `SHA256SUMS.txt`, and the draft is published only after all artifact checks and uploads succeed.

## Implementation notes to carry forward

- When changing `.goreleaser.yaml`, ensure `release.draft: true` and `checksum.name_template: SHA256SUMS.txt` are both applied so there is no stale `checksums.txt` asset path.
- Keep the DMG checksum regeneration/upload in one final path; avoid leaving both GoReleaser's pre-DMG checksum and a post-DMG checksum attached.
- Keep command output secret-safe around decoded App Store Connect key paths and imported signing identities. Logging secret names/presence is fine; values are not.
- Step 2 code review should include the planned validation evidence: `goreleaser check`, a local snapshot/dry-run app+DMG assembly path, and a release-mode negative test showing missing Apple prerequisites fail closed.

No further plan changes are required before implementation.
