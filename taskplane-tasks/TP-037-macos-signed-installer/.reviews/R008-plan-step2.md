# Plan review — TP-037 Step 2

Decision: **changes requested**

The revised Step 2 plan resolves the prior concrete gaps around Homebrew removal, standalone darwin archive suppression, script location, release-mode preflight, DMG signing, and explicit upload/checksum handling. The remaining issue is the ordering of publication vs. DMG packaging/upload. As currently written, the plan can still leave a public tag release in an incomplete state if GoReleaser publishes Linux/Windows assets first and the later DMG signing/notarization/upload step fails.

## Blocking plan gap

1. **Gate public release publication on the signed/notarized DMG being uploaded and checksummed.**  
   The plan says GitHub Actions will run `package_dmg.sh` after GoReleaser, append the DMG to `SHA256SUMS.txt`, and upload both with `gh release upload`. With the current `.goreleaser.yaml`, GoReleaser publishes directly to a non-draft GitHub release (`release.draft: false`). If signing, notarization, stapling, checksum append, or `gh release upload` fails afterward, the workflow fails but the public GitHub release may already exist with only the non-macOS artifacts/checksum. That is not fail-closed enough for this task's acceptance criteria.

   Please update the plan to name the exact publication gate, for example one of:

   - run GoReleaser in a build/checksum-only mode (`--skip=publish` or equivalent), package/notarize/staple the DMG, generate a single `SHA256SUMS.txt`, then create/upload all release assets in one final `gh release create/upload` step; or
   - publish GoReleaser output only to a draft release, upload the DMG and final `SHA256SUMS.txt`, then mark the release non-draft only after every artifact verification/upload succeeds; or
   - use a GoReleaser-supported `release.extra_files` / checksum integration that includes the already-created DMG before the release is published.

   The important contract is: a tag workflow must not expose a public release until the signed/notarized/stapled DMG and its checksum are both attached.

## Minor clarification to include in the updated plan

- If the checksum file is renamed to `SHA256SUMS.txt`, state that `.goreleaser.yaml` will change `checksum.name_template` from `checksums.txt` and that the workflow will avoid leaving a stale `checksums.txt` alongside the final checksum file.

Once the publication gate is made explicit, the Step 2 plan should be ready to implement.
