# Code review — TP-037 Step 3

Decision: **revise**

## Blocking findings

1. **macOS release job uses a GNU-only `base64 --decode` flag on a macOS runner.**
   - Location: `.github/workflows/release.yml:85` and `.github/workflows/release.yml:108`.
   - The `release` job runs on `macos-latest`, where `/usr/bin/base64` is BSD/macOS `base64` and does not reliably support the GNU long option `--decode`. These steps are responsible for decoding the Developer ID `.p12` and App Store Connect `.p8`; if they fail, the tag workflow cannot import the signing certificate or notarize the DMG, so the Step 3 acceptance path is broken before packaging starts.
   - Use a macOS-compatible decode command, e.g. `base64 -D`, or a portable decoder such as `python3`/`openssl base64 -d` if this snippet must remain cross-platform.

2. **Apple signing/notary secrets are still scoped at the release job level.**
   - Location: `.github/workflows/release.yml:51-57`.
   - This exposes `APPLE_DEVELOPER_ID_P12_BASE64`, `APPLE_DEVELOPER_ID_P12_PASSWORD`, and `APPLE_API_KEY_BASE64` to every release-job step, including unrelated third-party actions such as `actions/setup-go` and `goreleaser/goreleaser-action`. That directly conflicts with the Step 3 security requirement from the plan review to minimize secret exposure.
   - Remove the job-level `env` block. Scope the `.p12` secret and password only to the certificate-import step, scope `APPLE_API_KEY_BASE64` only to the API-key decode step, and pass only the metadata/path values needed by `build/macos/package_dmg.sh` to the packaging step. The GoReleaser action should only receive `GITHUB_TOKEN`.

3. **Task status records Step 3 as complete despite an unresolved revise review.**
   - Location: `taskplane-tasks/TP-037-macos-signed-installer/STATUS.md:47-49` and `:113`; conflicting source at `taskplane-tasks/TP-037-macos-signed-installer/.reviews/R011-plan-step3.md:3`.
   - The checked Step 3 items and history entry say the plan was approved, but the checked-in R011 review says `Decision: revise` and specifically called out the job-level Apple secret exposure that remains in the workflow. This makes the task state misleading and can cause the release workflow to be treated as accepted before the security requirement is fixed.
   - Update STATUS.md to reflect the actual R011 decision and leave the Step 3 checklist incomplete until the workflow issues above are addressed.

## Non-blocking notes

- Adding the `golangci-lint` gate before GoReleaser satisfies the missing lint preflight called out in R011.
- The release trigger remains `v*.*.*` rather than the prompt's broader `v*`. That may be intentional semver-only behavior, but it should be documented in STATUS.md if retained.

## Validation

- Ran `git diff ad05cae..HEAD --name-only`.
- Ran `git diff ad05cae..HEAD` and reviewed the full diff.
- Read `.github/workflows/release.yml`, `.goreleaser.yaml`, `build/macos/package_dmg.sh`, `SECURITY.md`, `.github/workflows/ci.yml`, `PROMPT.md`, `STATUS.md`, and the checked-in R011 plan review for context.
- Parsed `.github/workflows/release.yml` as YAML with Ruby successfully.
- Grepped for Apple secret placement, release upload/publish commands, and Homebrew-token paths. `actionlint` is not installed in this environment, so I could not run it locally.
