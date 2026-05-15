# Code review — TP-037 Step 2

Decision: **approved**

## Blocking findings

None.

## Non-blocking notes

- The implementation satisfies the Step 2 release-shaping requirements from the approved plan: Homebrew publishing is removed, standalone darwin archives are suppressed, GoReleaser publishes only a draft, and the DMG/checksum upload happens before the draft is made public.
- `build/macos/package_dmg.sh` is fail-closed in release mode and produces an explicitly unsigned `*_macos_universal_unsigned.dmg` only for local dry-run validation.
- `STATUS.md` still marks Step 3 as not started even though `.github/workflows/release.yml` now contains several Step 3 mechanics. That is not blocking for Step 2, but the next step should reconcile the status checklist with the workflow changes already made.

## Validation

- Ran `git diff e23515c..HEAD --name-only` and reviewed the full diff.
- Read the changed release configuration and packaging files: `.goreleaser.yaml`, `.github/workflows/release.yml`, `build/macos/package_dmg.sh`, `build/macos/Info.plist`, `go.mod`, `go.sum`, and `STATUS.md`.
- Ran `goreleaser check` successfully.
- Ran `goreleaser release --snapshot --clean --skip=publish`; it produced Linux/Windows archives and `dist/icuvisor-darwin-universal_darwin_all/icuvisor` without a standalone darwin archive.
- Ran `build/macos/package_dmg.sh`; it produced an unsigned dry-run DMG with the expected warning.
- Ran `ICUVISOR_MACOS_RELEASE=1 build/macos/package_dmg.sh`; it failed closed immediately because Apple release prerequisites were absent.
