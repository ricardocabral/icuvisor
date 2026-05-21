# Plan review — TP-037 Step 2

Decision: **changes requested**

The Step 2 plan has incorporated the previous review themes into the STATUS checklist, but it still reads as intent rather than an implementation plan. For a Level 3 release/signing change, the plan needs to make the release-shaping decisions now, before code is written.

## Blocking plan gaps

1. **Choose the Homebrew behavior, not "remove or disable."**  
   The plan currently says to "remove or disable Homebrew auto-publishing." That leaves the key release-risk unresolved. Pick the exact approach for TP-037, e.g. remove the `brews` block from `.goreleaser.yaml` and remove `HOMEBREW_TAP_GITHUB_TOKEN` from the release workflow, or keep it but prove it is skipped by the exact GoReleaser command/config split. The task says no Homebrew auto-publish in this release path, so the plan must be unambiguous.

2. **Decide what happens to standalone darwin archives.**  
   The plan says to prevent unsigned standalone macOS archives from release promotion, but does not state whether Step 2 will suppress darwin archives entirely or sign/notarize every macOS artifact that remains attached. This is a release contract issue, not an implementation detail. State the exact GoReleaser/archive strategy.

3. **Name the files/scripts and their responsibilities.**  
   The plan mentions `.app` assembly order, signing, DMG creation, notarization, and stapling, but does not identify where this logic will live. Please specify the concrete files to add/change, such as `build/macos/package_app.sh`, `build/macos/sign_app.sh`, `build/macos/make_dmg.sh`, or a single script, and which GoReleaser hooks invoke them. This will make reviewable whether secrets are handled safely and whether macOS-only tooling is isolated.

4. **Define the snapshot vs tag-release switch precisely.**  
   The plan says snapshots are unsigned with warnings and real releases fail closed, but not what condition controls that behavior. State the exact environment variable/GoReleaser metadata/CLI mode the scripts will use, and which prerequisites are required in release mode: signing identity, imported keychain item, Team ID if used, and notary credentials. Avoid any path where a tag release can silently produce an unsigned DMG.

5. **Add an expected validation list for the implemented Step 2.**  
   Before implementation, the plan should say which local checks will be run and recorded after Step 2, for example `goreleaser check`, a snapshot/dry-run command that exercises unsigned `.app`/DMG assembly, and a release-mode negative test showing missing Apple credentials fail closed. Live notarization can remain operator-deferred per steering, but the fail-closed path should be testable now.

## Non-blocking suggestions

- Prefer one small orchestration script plus small helper functions over large inline GoReleaser YAML hooks.
- Use `SHA256SUMS.txt` in the plan if Step 2 touches checksums, so Step 3 does not need to rename the artifact later.
- If suppressing darwin archives, keep Linux/Windows behavior unchanged to avoid expanding the task beyond macOS packaging.

Once the plan makes the above choices explicitly, Step 2 should be ready to implement without another broad design pass.
