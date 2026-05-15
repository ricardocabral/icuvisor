# Plan review — TP-037 Step 2

Decision: **changes requested**

The updated Step 2 plan now makes the major release decisions from R005/R006: remove Homebrew publishing, suppress standalone darwin archives, use a single macOS packaging script, fail closed in release mode via `ICUVISOR_MACOS_RELEASE=1`, and validate with `goreleaser check` plus local dry-run/negative tests. That is much closer.

There is still one release-pipeline integration gap that should be resolved before implementation, because it affects whether the signed/notarized DMG is actually a GoReleaser/GitHub release artifact and whether checksums cover the promoted file.

## Blocking plan gaps

1. **Name a valid GoReleaser integration point for `package_dmg.sh`.**  
   The plan says `build/macos/package_dmg.sh` will be invoked as a GoReleaser “after-hook”. With the current installed GoReleaser schema (`goreleaser 2.15.4`), there is no top-level `after` hook in `.goreleaser.yaml`; build hooks exist as `builds[].hooks.pre/post`, but those run around individual Go builds and are not a good fit for packaging the final `universal_binaries` output. A script that runs after the checksum/release phases also will not automatically become a release artifact or be included in `checksums.txt`/`SHA256SUMS.txt`.

   Please update the plan to state the exact supported mechanism, for example one of:

   - generate the DMG before GoReleaser checksum/release artifact collection and register it with `checksum.extra_files` and `release.extra_files`; or
   - make DMG creation/upload/checksum an explicit GitHub Actions step outside GoReleaser, with `gh release upload` and a separately generated `SHA256SUMS.txt`; or
   - use another GoReleaser-supported publisher/upload path that `goreleaser check` validates and that is proven to include the DMG in the tag-release output.

   The important part is that the plan must make it impossible to finish a tag release with only Linux/Windows artifacts and a locally-created DMG that was never uploaded/checksummed.

2. **Make the darwin archive suppression strategy concrete in YAML terms.**  
   The status now says standalone darwin archives will be suppressed, which is the right decision. The current config, however, produces `dist/icuvisor_<version>_darwin_all.tar.gz` from the universal binary. Before implementation, state how the config will prevent that artifact: e.g. split build IDs/archive entries so the archive pipe only sees Linux/Windows artifacts, or use an equivalent GoReleaser-supported exclusion strategy. “Suppress darwin archives” is a decision; the plan still needs the concrete config shape to avoid leaving the unsigned `darwin_all.tar.gz` attached.

3. **Clarify whether the DMG itself is code-signed before notarization.**  
   The task mission and acceptance criteria call for a signed, notarized, stapled `.dmg`. The plan clearly signs/verifies the `.app`, creates the DMG, notarizes, staples, and validates, but it does not say whether `codesign --options runtime --timestamp` is also applied to the DMG before `notarytool submit`. Please make the intended sequence explicit and include the verification command for the DMG artifact itself, not only the app bundle.

## Non-blocking notes

- I verified the current baseline with a snapshot: it creates a `darwin_all.tar.gz` archive and a Homebrew formula, matching the risks called out in prior reviews. The updated plan's intent to remove both is correct.
- Prefer naming the checksum file `SHA256SUMS.txt` in Step 2 if the checksum config is touched now; the prompt uses that name for the released DMG verification flow.
- Keep the snapshot behavior visibly unsafe-for-release: unsigned/ad-hoc DMGs are fine for dry-run assembly only, but the generated artifact name or log should make that clear.

Once the plan identifies a valid artifact/checksum/upload integration path and the concrete archive suppression approach, Step 2 should be ready to implement.
