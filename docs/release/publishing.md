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
5. Winget submission from the `publish-winget` job, or an explicit skip because `WINGET_PAT` is absent or the tag is not a stable `vX.Y.Z` release.

The standalone Winget workflow is manual recovery only. The normal `release.published` event must not start an untracked asynchronous Winget publish after the primary release workflow has already reported success.

## Partial publish recovery

Do not retag an existing version and do not create a duplicate version to recover from a failed publish. First identify which publish surface completed:

- **Failure before GoReleaser uploads:** fix the issue and rerun the same tag normally.
- **Failure after Homebrew/Scoop updated but before GitHub release publication, checksums, MSI, DMG, MCPB, or Winget completed:** verify the tap and bucket already point at the intended tag and checksums, then rerun the `Release` workflow manually for the same tag with `skip_package_managers=true`. That recovery input withholds both `HOMEBREW_TAP_PAT` and `SCOOP_BUCKET_PAT` from GoReleaser while leaving normal tag-push releases unchanged.
- **Failure after the GitHub release is public but Winget submission failed:** fix the Winget issue and run `Publish to Winget (manual recovery)` with the same tag. Do not rerun package-manager publishing unless the Homebrew/Scoop artifacts are wrong.
- **Duplicate-version errors:** treat them as a signal that a publish surface may already be complete. Compare the published formula, bucket entry, Winget PR/manifests, and GitHub release assets against the intended tag before deciding whether to skip that surface on the rerun.

If a published artifact is incorrect rather than merely incomplete, ship a new patch version instead of mutating tags.
