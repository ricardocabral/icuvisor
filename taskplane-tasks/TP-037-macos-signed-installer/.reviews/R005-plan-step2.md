# Plan review — TP-037 Step 2

Decision: **changes requested**

The Step 2 checklist in `STATUS.md` is directionally aligned with the task, but it is still too high-level for a Level 3 signing/notarization change. Before implementation, please make the plan concrete enough that the release path cannot accidentally publish unsigned macOS artifacts or an out-of-scope Homebrew update.

## Blocking plan gaps

1. **Account for the existing Homebrew publishing path before touching GoReleaser.**  
   `.goreleaser.yaml` still has a `brews` publisher, and `.github/workflows/release.yml` still exports `HOMEBREW_TAP_GITHUB_TOKEN`. TP-037 explicitly says this task must not auto-publish to Homebrew or any other channel. The Step 2/3 plan needs to state whether the `brews` section will be removed, temporarily gated, or skipped in the release command. Do not leave a tag release able to publish the DMG and also update `homebrew-icuvisor` as a side effect.

2. **Define what happens to existing macOS tar/zip archives.**  
   The current GoReleaser config archives every build, including darwin artifacts. If Step 2 only signs/notarizes the DMG but leaves unsigned darwin `.tar.gz` artifacts attached to the same release, it weakens the acceptance criterion that an unsigned build cannot be promoted. The plan should explicitly choose one of:
   - remove/suppress standalone darwin archives and publish the signed/notarized DMG as the macOS artifact; or
   - ensure every macOS artifact that remains attached is signed and verification instructions cover it.

3. **Specify the `.app` assembly and signing order.**  
   The plan mentions `codesign`, but not the concrete packaging sequence. It should say that Step 2 will create `icuvisor.app/Contents/MacOS/icuvisor`, copy/substitute `build/macos/Info.plist` version placeholders from GoReleaser metadata, sign the Mach-O/app bundle with `--options runtime --timestamp`, verify with `codesign --verify --deep --strict`, then build the DMG, submit it to `notarytool --wait`, staple the ticket, and validate stapling. This avoids signing the wrong path or notarizing an artifact that is later modified.

4. **Make credential/preflight behavior explicit for snapshots vs releases.**  
   Local Apple credentials are currently absent by design. The plan should state how scripts/hooks behave in each mode:
   - snapshot/local builds may create an unsigned DMG scaffold or skip live signing/notarization with an explicit warning; and
   - real non-snapshot/tag releases must fail hard if the signing identity, imported keychain item, Team ID, or notary credentials are missing.  
   Avoid any ad-hoc signing fallback that could produce a superficially valid but non-Developer-ID artifact.

5. **Constrain macOS-only tooling.**  
   `hdiutil`, `codesign`, `xcrun notarytool`, and `stapler` are macOS-only. The plan should ensure any GoReleaser hooks/scripts either run only on the macOS release job or fail with a clear message, without breaking ordinary non-macOS development commands unexpectedly.

## Non-blocking implementation suggestions

- Prefer small `build/macos/*.sh` scripts called by GoReleaser over large inline YAML hooks; they are easier to lint/review and can centralize secret-safe logging.
- Keep command tracing off around secret-derived paths/credentials and log only secret names or presence checks.
- Consider naming the checksum artifact to match the prompt (`SHA256SUMS.txt`) when the release workflow is updated; current config uses `checksums.txt`.

## Validation expected after implementation

For Step 2 code review, I will expect to see at least:

- `goreleaser check` passing.
- A snapshot/dry-run path that exercises app/DMG assembly without requiring Apple secrets, or a documented reason it cannot be exercised locally.
- Release-mode scripts that fail closed when Developer ID/notary prerequisites are missing.
- No remaining automatic Homebrew publication path for the TP-037 release.
