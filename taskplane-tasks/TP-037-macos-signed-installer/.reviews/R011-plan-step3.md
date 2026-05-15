# Plan review — TP-037 Step 3

Decision: **revise**

The Step 3 direction is aligned with the approved Step 2 release-shaping decision: draft-first publishing, explicit DMG/checksum upload, and final publication only after signing/notarization/stapling succeeds. However, the plan is still too thin for a release-pipeline/security step and misses one explicit task requirement.

## Blocking findings

1. **The plan does not explicitly include the required lint preflight.**
   - `PROMPT.md` Step 3 requires the tag workflow to “reuse the existing test/lint preflight.”
   - The current Step 3 checklist says “reuses release preflight,” but the existing `.github/workflows/release.yml` preflight only runs tidy, `go test -race`, and `goreleaser check`; it does not run the existing `golangci-lint` preflight from `ci.yml`.
   - Revise the plan to add a release preflight lint gate before any draft release/artifact publication. Duplicating the `golangci/golangci-lint-action` step from CI is acceptable; a reusable workflow is also fine if kept focused.

2. **The plan should constrain Apple secret exposure to the minimum necessary workflow steps.**
   - “All Apple secrets via GitHub Actions secrets” is necessary but not sufficient for this security-sensitive step.
   - Avoid putting Apple signing/notary secrets in job-level `env`, because that exposes them to unrelated steps and third-party actions such as the GoReleaser action.
   - Revise the plan so:
     - `APPLE_DEVELOPER_ID_P12_BASE64` and `APPLE_DEVELOPER_ID_P12_PASSWORD` are scoped only to the certificate-import step.
     - `APPLE_API_KEY_BASE64` is scoped only to the notary-key decode step.
     - Only non-secret metadata and the decoded key file path needed by `build/macos/package_dmg.sh` are made available to the packaging step.
     - No placeholder secret values, decoded `.p12`, or decoded `.p8` material are committed or echoed.

3. **Add explicit validation for the workflow-only changes.**
   - Because live Apple credentials are operator-deferred, the Step 3 plan needs a concrete validation path that can run without publishing a real release.
   - Include at least workflow syntax/static validation where available (`actionlint` or an equivalent YAML/GitHub Actions check), `goreleaser check`, and the existing local negative release-mode packaging test proving missing Apple prerequisites fail closed.
   - Also verify the final workflow no longer contains Homebrew publishing or `HOMEBREW_TAP_GITHUB_TOKEN` usage in the tag-release path.

## Non-blocking notes

- Keeping GoReleaser publication as a draft and publishing only after the DMG and regenerated `SHA256SUMS.txt` are uploaded remains the right shape.
- `APPLE_TEAM_ID` is already documented in `SECURITY.md`/`STATUS.md`; keep it in the plan if `package_dmg.sh` requires it, even though the original Step 3 prompt listed only the five Apple cert/notary secrets.
- If the workflow trigger remains stricter than `v*` (for example `v*.*.*`), note the intentional semver-only choice in the plan/status so it is not mistaken for missing the prompt requirement.
