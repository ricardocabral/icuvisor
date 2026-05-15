# Plan review — TP-037 Step 4

Decision: **revise**

The Step 4 checklist is pointed in the right direction, but it is still too thin for the user-facing installer/client documentation that is part of the v0.5 acceptance criteria. It also inherits an inaccurate progress record for Step 3 that should be corrected before documenting the release flow as done.

## Blocking findings

1. **The plan must explicitly remove stale secret-bearing Claude Desktop examples, not just add a keychain callout.**
   - The current `docs/clients/claude-desktop.md` is a v0.1 build-from-source guide and still contains copy-pasteable examples with `INTERVALS_ICU_API_KEY`, `api_key`, and `YOUR_INTERVALS_ICU_API_KEY` in JSON/env snippets.
   - Step 4's checklist says both client docs will include an "API key in keychain, not in JSON" callout, but that is not sufficient if the old examples remain in the same page as alternate paths.
   - Revise the plan so the primary Claude Desktop and Claude Code snippets contain only non-secret values: `command` set to `/Applications/icuvisor.app/Contents/MacOS/icuvisor`, `INTERVALS_ICU_ATHLETE_ID`, optional `ICUVISOR_TIMEZONE`, optional `ICUVISOR_TRANSPORT=stdio`, and any write/toolset toggles only if clearly marked optional. The API key must be handled by the macOS Keychain instructions, not by client config snippets.
   - If legacy env/file API-key fallback is mentioned at all, keep it in a clearly discouraged troubleshooting/advanced section rather than as a normal copy-paste path for beta athletes.

2. **`docs/install/macos.md` needs a concrete creation plan, not a broad "updated for signed-DMG" bullet.**
   - This file does not currently exist, and the task prompt requires it to cover installer install path, Gatekeeper/first-launch behavior, signature verification, and uninstall.
   - Revise the Step 4 plan to specify the sections to create: download/open DMG, drag `icuvisor.app` to `/Applications` (and how snippets change if the user chooses `~/Applications`), store the intervals.icu API key in Keychain service `icuvisor` / account `intervals-icu-api-key`, verify with `codesign --verify --deep --strict`, `spctl -a -v`, and `xcrun stapler validate` where applicable, explain the headless `LSUIElement` launch behavior, and uninstall both the app and optional keychain item.
   - Also state explicitly that no LaunchAgent is auto-loaded by the installer. If optional LaunchAgent/power-user behavior is documented, it must be clearly optional and consistent with the current repo (do not imply a resident service exists unless the plist is actually added).

3. **The Claude Code plan needs exact scope/schema details and must not be conflated with the existing Codex guide.**
   - The repository already has `docs/clients/codex-local.md`; Step 4 must add a separate `docs/clients/claude-code.md` for Claude Code, not adapt Codex CLI instructions.
   - Revise the plan to include the exact `.mcp.json` shape and where the user should put it (project-local vs user/global, whichever this task chooses), plus a restart/new-session instruction. The snippet should mirror the installed `.app` path and keychain-only API-key stance from the Desktop doc.

4. **Add a validation plan for the documentation snippets.**
   - The acceptance criteria require copy-pasteable configs that work for non-developers. The current plan has no way to catch broken JSON or accidental secret placeholders.
   - Add validation steps such as: extract/parse the JSON snippets (or keep them standalone enough to run through `python -m json.tool`/`jq`), grep the client docs to ensure `INTERVALS_ICU_API_KEY`, `api_key`, and `YOUR_INTERVALS_ICU_API_KEY` do not appear in primary client-config examples, check that links to diagnostics/install docs resolve, and run a docs-oriented smoke check of the installed command path text.

5. **Reconcile the Step 3 status before Step 4 documents the release/operator flow.**
   - `STATUS.md` marks R011 and R012 as approved, but the checked-in `R011-plan-step3.md` says `Decision: revise`, and the current workflow still shows issues called out there/afterward: Apple secrets are scoped at job level and the macOS job uses `base64 --decode`.
   - Step 4 includes updating `SECURITY.md`/release-checklist docs for the signed-DMG/operator-preflight flow. Do not document Step 3 as accepted until the status/history and workflow state are corrected, or the docs will encode a release flow that still has known review defects.

## Non-blocking notes

- The verify recipe should use the prompt from the task: ask the client "What's my FTP?" and expect a populated answer from the athlete profile/sport settings. If it fails, link to a diagnostics/troubleshooting section in the same doc or `docs/install/macos.md`.
- Include the MCP client cache caveat: after changing config or upgrading icuvisor, fully restart the client and start a new conversation/session.
- Update `README.md` Quickstart to put "Download for macOS" above build-from-source, and add the Step 4 documentation entry under `CHANGELOG.md` `[Unreleased]`.
