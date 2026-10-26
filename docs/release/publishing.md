# Release publishing and recovery

Use this checklist with the tagged GitHub Actions release workflow. It is a maintainer guide; it does not publish anything by itself.

## Non-publishing preflight

Before tagging a release, run:

```bash
make release-preflight
```

`make release-preflight` validates `server.json`, the GoReleaser config, and the MCPB manifest.

## Publish gates

The primary `Release` workflow should be considered successful only after all of these have completed or explicitly skipped:

1. release preflight, tests, lint, GoReleaser config check, MCPB manifest validation, and distribution metadata validation;
2. GoReleaser draft release creation, including optional Homebrew and Scoop uploads when their PAT secrets are present;
3. signed MCPB, macOS DMG, Windows MSI, archives, regenerated checksums, and cosign signatures;
4. GitHub draft release publication; and
5. Winget submission only when the `publish-winget` job is enabled. It is currently intentionally disabled because the Winget fork token cannot execute `UpdateRef`.

The standalone Winget workflow is manual recovery only. The normal `release.published` event must not start an untracked asynchronous Winget publish after the primary release workflow has already reported success.

## Partial publish recovery

Do not retag an existing version and do not create a duplicate version to recover from a failed publish.

Release runs for the same tag share a concurrency group, including tag pushes and manual recovery runs. [GitHub concurrency queues](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency) let up to 100 runs wait rather than cancelling the active or pending releases. Preflight accepts version tags such as `v1.2.3` (including prerelease/build metadata), rejects branches and full Git refs, rejects already-published tags, and stops if GitHub release state cannot be verified.

The release-state guard and GoReleaser configuration come from the selected workflow revision. Application source and packaging scripts remain pinned to the requested tag. This lets manual recovery use updated release safeguards without changing an old tag.

For unpublished tags, normal GoReleaser retries reuse the existing draft, replace conflicting archive/checksum assets, and refresh notes from the tagged changelog. Failed runs retain the draft for recovery; successful runs publish that same draft after all installers and signatures are ready. Do not delete the draft to retry a failed release.

When `skip_package_managers=true`, preflight requires exactly one existing draft with all five original base archives (Linux amd64/arm64, Windows amd64/arm64, and macOS universal), each fully uploaded and nonempty. GoReleaser builds locally with `--skip=publish`, preserving those draft archives so existing Homebrew/Scoop checksums stay valid. The remaining steps upload installers and regenerate the final signed checksums. If original archives are missing or incomplete, reconcile the package-manager metadata before a normal retry; rebuilding them can change their hashes even for the same source tag.

First identify which publish surface completed:

- **Failure before GoReleaser uploads:** fix the issue and rerun the same tag normally.
- **Failure after Homebrew/Scoop updated but before GitHub release publication, checksums, MSI, DMG, MCPB, or Winget completed:** verify the tap and bucket already point at the intended tag and checksums, then rerun the `Release` workflow manually for the same tag with `skip_package_managers=true`. That recovery input withholds both `HOMEBREW_TAP_PAT` and `SCOOP_BUCKET_PAT` from GoReleaser while leaving normal tag-push releases unchanged.
- **Failure after the GitHub release is public but Winget submission failed:** the primary workflow currently skips Winget because of the fork permission issue. After that issue is fixed, run `Publish to Winget (manual recovery)` with the same tag. Do not rerun package-manager publishing unless the Homebrew/Scoop artifacts are wrong.
- **Duplicate-version errors:** treat them as a signal that a publish surface may already be complete. Compare the published formula, bucket entry, Winget PR/manifests, and GitHub release assets against the intended tag before deciding whether to skip that surface on the rerun.

If a published artifact is incorrect rather than merely incomplete, ship a new patch version instead of mutating tags.
