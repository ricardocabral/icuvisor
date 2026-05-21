# Plan review — TP-053 Step 2

Decision: **changes requested** before drafting the page.

The Step 2 checklist in `STATUS.md` covers the general Diataxis writing constraints, but it does not yet resolve two implementation details that will determine whether the draft can be correct on the first pass: the exact copy-pasteable ChatGPT command JSON for the build-from-source path, and the missing Tutorials section in the current website tree.

## Findings

1. **Blocking — the plan has not resolved the exact connector JSON after choosing the build-from-source fallback.**  
   Step 1 determined that no signed DMG/app artifact exists, so the tutorial must use the build-from-source path. The prompt's original ChatGPT JSON points at `/Applications/icuvisor.app/Contents/MacOS/icuvisor`, which will not exist for that fallback. The Step 2 plan still says to include the "exact ChatGPT local-stdio JSON" but does not say what command path the reader can paste without placeholders. This is especially important because the prompt forbids `<...>` placeholders and Step 1 review already called out that the connector JSON must be updated if the fallback path differs.

   Before drafting, update the plan with a deterministic install/build location and the exact JSON that matches it. For example, the tutorial could have the reader clone/build in a fixed path and then use that same absolute command path in the ChatGPT JSON, or it could add a simple copy/install command that places the binary at a stable path. Do not draft a JSON snippet that points at the developer worktree, uses `/Applications/icuvisor.app` without creating it, relies on shell expansion such as `~`, or asks the reader to mentally substitute their username.

2. **Blocking — the current site does not have a Tutorials section to host the page in navigation.**  
   `web/content/` currently has `connect/`, `explain/`, `guides/`, `install/`, and `reference/`, but no `tutorials/` section. The layouts' primary nav also has no Tutorials link. Step 4 requires the page to appear under "Tutorials" in nav, so Step 2 needs to either verify that the TP-050 scaffold is intentionally absent and add the minimum section/index/nav change now, or record the missing dependency as a blocker before writing an orphaned page. A standalone `web/content/tutorials/getting-started-chatgpt.md` may render by URL, but it will not satisfy the navigation acceptance check as the site is currently structured.

3. **Major — the draft plan should explicitly carry forward the Step 1 honesty constraints.**  
   `STATUS.md` records that live ChatGPT UI was unavailable, current public ChatGPT docs describe HTTPS `/mcp`, and Codex CLI was accepted by operator direction as the local-stdio simulator. The Step 2 plan should state how the page will stay honest: do not mention Codex, do not fabricate ChatGPT screenshots or UI states, and keep the HTTPS-docs caveat out of the tutorial body per steering. If the page includes a representative first answer, base it on the redacted aggregate-only validation output and avoid athlete IDs, activity names, locations, API keys, or real private URLs.

4. **Minor — include the footer-only troubleshooting link decision in the drafting checklist.**  
   The prompt says not to link troubleshooting from inside the tutorial body and to mention it only in a footer. Step 2 should make that explicit so Keychain/Gatekeeper/ChatGPT failure modes recorded in Step 1 do not creep into the happy path.

## Suggested revised Step 2 checklist

- Confirm whether TP-050's Tutorials scaffold exists in this worktree. If it does not, add the minimum `web/content/tutorials/` section/index/nav work needed for the page to appear under Tutorials, or mark the dependency as blocked.
- Choose a deterministic build-from-source install location and write the exact commands that create it.
- Write the exact ChatGPT local-stdio JSON that points at that location, with no API key, no Hugo shortcode placeholders, no username placeholders, and no reliance on `~` expansion.
- Draft `web/content/tutorials/getting-started-chatgpt.md` in the required Diataxis order, second person, present tense, short paragraphs, no marketing voice, no `Note:`/`Tip:` interruptions, no conceptual detours, and no troubleshooting branches in the body.
- Use the redacted Step 1 validation result only as a representative aggregate answer; do not expose athlete identifiers, API keys, activity titles, locations, or private URLs.
- Add screenshot references only for screenshots that will be captured/redacted in Step 3; do not fabricate unavailable ChatGPT UI evidence.
- Update `CHANGELOG.md` under `[Unreleased]` / `Added` with the new tutorial.
