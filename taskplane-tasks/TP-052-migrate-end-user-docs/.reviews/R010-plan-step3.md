# R010 Plan Review — Step 3: Install section

Verdict: APPROVE

The Step 3 plan is scoped correctly for the prompt: it creates the install chooser plus macOS, Windows, and Linux pages; adapts `docs/install/macos.md`; keeps the user-facing install/verification/uninstall material; and leaves unsupported platforms as placeholders rather than inventing installers.

Non-blocking implementation notes:

1. On `install/macos.md`, strip the entire maintainer/release-operator preflight section. Keep only user-facing verification commands (`codesign`, `spctl`, `stapler validate`, `icuvisor version`), DMG drag-to-Applications flow, first-run `icuvisor setup`, uninstall, and the MCP executable path.
2. Avoid moving broader API-key guide material into Step 3. It is fine for the macOS page to mention `icuvisor setup` and the Keychain service/account, but the cross-platform manual keychain instructions belong in `guides/api-key.md` in Step 5.
3. Windows and Linux pages should clearly say official installers are not available yet / coming with v1.0. Do not imply current support beyond building from source, and link to a top-level GitHub artifact such as the repository or releases, not to `docs/` developer pages.
4. Use Hugo `relref` for internal website links and avoid bare links to migrated `docs/` pages. If linking to future Step 4 client pages from the macOS page, either link to the existing `connect/_index.md` for now or be aware that a per-step Hugo build will fail until those target pages are authored.
5. Do not edit `README.md`, delete source files under `docs/`, or add developer-only release-signing details to the website.

With those guardrails, the plan is ready to implement.
