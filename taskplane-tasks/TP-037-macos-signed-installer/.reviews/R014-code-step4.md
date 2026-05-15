# Code review — TP-037 Step 4

Decision: **revise**

## Blocking findings

1. **Task status still records revise reviews as approvals and marks the release workflow complete.**
   - Locations: `taskplane-tasks/TP-037-macos-signed-installer/STATUS.md:43-49`, `:115-117`; conflicting records at `taskplane-tasks/TP-037-macos-signed-installer/.reviews/R012-code-step3.md:3` and `taskplane-tasks/TP-037-macos-signed-installer/.reviews/R013-plan-step4.md:3`.
   - The current diff adds checked-in review files whose decisions are `revise`, but `STATUS.md` records R012 and R013 as `APPROVE` and keeps Step 3 as complete. R012 also documents concrete unresolved release-job blockers that still exist in `.github/workflows/release.yml` (`base64 --decode` on the macOS runner and job-level Apple secret scope). Step 4 adds user/operator release docs, so carrying a false “approved/complete” task state into those docs makes the signed-DMG flow look accepted when the committed review evidence says it is not.
   - Reconcile the task state before proceeding: either fix the R012 workflow blockers and add a superseding approving review/history entry, or mark the affected reviews/steps as revise/incomplete in `STATUS.md` and avoid documenting the release path as accepted.

2. **The macOS install doc offers `~/Applications` but leaves copy-paste commands hard-coded to `/Applications`.**
   - Location: `docs/install/macos.md:19-23`, `:38-42`, `:69-77`.
   - The install step says users may drag `icuvisor.app` to `/Applications` or `~/Applications`, but the first-launch command, verification commands, and uninstall command all target `/Applications/icuvisor.app`. A non-developer who chooses the user-local install path (explicitly allowed by the task prompt) will immediately get failing commands and then client snippets that point at the wrong binary unless they infer every substitution themselves.
   - Make `/Applications` the sole recommended copy-paste path, or add explicit `~/Applications` equivalents for first launch, verification, MCP `command`, and uninstall (for example `$HOME/Applications/icuvisor.app/Contents/MacOS/icuvisor` and `rm -rf "$HOME/Applications/icuvisor.app"`).

3. **Release-operator docs disagree on the tag pattern that triggers the workflow.**
   - Locations: `.github/workflows/release.yml:4-5`, `docs/install/macos.md:108`, `SECURITY.md:78`.
   - The workflow and new macOS checklist use `v*.*.*`, but `SECURITY.md` still tells the operator to run the workflow with a `v*` tag. An operator following the security checklist with a tag such as `v0.5` would not start the release workflow, which is exactly the kind of sharp edge this Step 4 operator documentation is meant to remove.
   - Use one pattern consistently across the workflow, install checklist, security checklist, and STATUS decision notes. If semver-only `vX.Y.Z` tags are intentional, document that explicitly.

## Non-blocking notes

- The Claude Desktop and Claude Code JSON snippets are valid JSON and no longer include the API key in the MCP config blocks.
- The client docs include the requested `What's my FTP?` verification prompt and new-session/restart guidance.

## Validation

- Ran `git diff c361fcd..HEAD --name-only`.
- Ran `git diff c361fcd..HEAD` and reviewed the full diff.
- Read `PROMPT.md`, `STATUS.md`, `README.md`, `SECURITY.md`, `docs/install/macos.md`, `docs/clients/claude-desktop.md`, `docs/clients/claude-code.md`, and the checked-in R012/R013 review files.
- Parsed the JSON snippets in both Claude client docs with Python `json.loads` successfully.
- Grepped for API-key placeholders in client docs and inspected the release workflow tag/secret/base64 lines for context.
