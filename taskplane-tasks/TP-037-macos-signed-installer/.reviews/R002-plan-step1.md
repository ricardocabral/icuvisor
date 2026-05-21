# Plan review — TP-037 Step 1

Decision: **approved to proceed**

The updated `STATUS.md` now contains a concrete Step 1 plan instead of only the task checklist. It addresses the R001 blockers: Apple signing metadata is constrained to non-secret facts, the bundle identifier is treated as permanent, the TP-036 keychain namespace was audited, the headless `.app`/stdio/Finder/LaunchAgent behavior is documented, and the planned `Info.plist` keys plus release-time version substitution are called out.

## What looks good

- **Secret boundaries are explicit.** The plan correctly says not to fabricate or commit Apple Developer facts, and names only the non-secret metadata that should be recorded: Team ID, Developer ID Application common name, certificate expiration, and whether the `.p12`/password have been placed in GitHub Actions secrets.
- **Bundle identifier is locked deliberately.** `dev.icuvisor.icuvisor` is now recorded as final/permanent, which is important before any signed keychain prompts or user trust decisions are created.
- **TP-036 interaction is covered.** The current credential lookup strings match the code (`credstore.ServiceName = "icuvisor"`, `credstore.IntervalsAPIKeyAccount = "intervals-icu-api-key"`), and the plan correctly avoids a data migration while warning that macOS may prompt when the designated requirement changes.
- **Launch model is clear enough for v0.5.** MCP clients execute the binary inside the app bundle directly over stdio; Finder launch is only for Gatekeeper/keychain trust and may have no visible UI; LaunchAgent support remains optional documentation and is not auto-loaded.
- **Info.plist scope is appropriately narrow.** The required bundle metadata is listed, including `LSUIElement=true`, with version values intended to be substituted during release packaging rather than hand-edited each release.

## Required before marking Step 1 complete

These are execution checks, not plan blockers:

1. Replace the Apple Developer placeholders/TBDs with real non-secret metadata once the maintainer provisions the Developer ID Application certificate. Do not mark `Developer ID cert enrolled, .p12 exportable for CI` complete until Team ID, cert common name, expiration date, and GH secret presence are recorded in `STATUS.md`.
2. Strengthen the bundle-ID rationale from "`icuvisor.dev` is used in repository metadata" to an explicit maintainer assertion that the project controls or is authorized to use the `icuvisor.dev` namespace. Repository references alone are not proof of namespace control.

## Carry-forward note for later steps

The existing `.goreleaser.yaml` still has a `brews` section targeting `homebrew-icuvisor`. TP-037 says not to auto-publish Homebrew in this task, so Step 2/3 should either disable/gate that behavior for the DMG release path or document why it cannot be triggered by the new workflow.
