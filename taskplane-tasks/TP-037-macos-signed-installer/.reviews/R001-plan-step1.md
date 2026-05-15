# Plan review — TP-037 Step 1

Decision: **changes requested**

I only found the task prompt and `STATUS.md`; there is not an explicit Step 1 implementation plan beyond the checklist. The checklist is directionally correct, but it is not yet specific enough for a Level 3 task that anchors signing identity and keychain trust.

## Blocking plan gaps

1. **Do not fabricate Apple Developer facts.** Step 1 depends on human/maintainer setup in Apple Developer and GitHub Actions. The plan should say exactly what will be verified and recorded in `STATUS.md` (Team ID, certificate common name/expiration, chosen bundle identifier), and explicitly that no `.p12`, `.p8`, passwords, or secret values will be committed or pasted into status.

2. **Bundle ID needs an ownership/permanence check.** `dev.icuvisor.icuvisor` is plausible, but the plan should include confirming control of the `icuvisor.dev` namespace and recording why this reverse-DNS ID is final. Since the prompt calls out keychain ACL breakage, Step 1 should also audit the current TP-036 keychain service/account usage and state whether the signed app's designated requirement/bundle ID affects existing saved credentials or requires a migration note.

3. **Clarify the `.app` launch model before committing `LSUIElement=true`.** A headless `.app` with `LSUIElement=true` is fine for code signing, but double-clicking it may start a stdio MCP server with no visible UI and no useful client connection. The plan should state the intended behavior for:
   - MCP clients executing `/Applications/icuvisor.app/Contents/MacOS/icuvisor` directly;
   - Finder double-click / `open /Applications/icuvisor.app`;
   - whether the optional LaunchAgent is documented only and not auto-loaded.

4. **Info.plist/versioning details are missing.** If Step 1 creates or locks `build/macos/Info.plist`, the plan should specify the required keys (`CFBundleIdentifier`, `CFBundleExecutable`, `CFBundleName`, `CFBundleVersion`, `CFBundleShortVersionString`, `LSUIElement`) and how release-time versions are substituted by GoReleaser rather than hard-coded.

## Non-blocking note for later steps

The existing `.goreleaser.yaml` currently has a `brews` section that publishes to `homebrew-icuvisor`, while TP-037 explicitly says not to auto-publish Homebrew in this task. That is a Step 2/3 issue, not Step 1, but the implementation plan should account for disabling or gating it before the release workflow is changed.

## Suggested Step 1 acceptance for the plan

Before implementation proceeds, update the plan/status so Step 1 ends with:

- Apple Team ID and Developer ID Application identity recorded in `STATUS.md` using non-secret metadata only.
- Final bundle identifier recorded as permanent, with namespace ownership rationale.
- Headless `.app`/stdio/Finder/LaunchAgent behavior documented.
- Planned `Info.plist` fields and version-substitution mechanism documented.
- Explicit note that Apple secrets are only created as GitHub Actions secrets and are never committed.
