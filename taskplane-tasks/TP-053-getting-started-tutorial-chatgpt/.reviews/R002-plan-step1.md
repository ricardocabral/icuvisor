# Plan review — TP-053 Step 1

Decision: **approved**.

The revised Step 1 checklist in `STATUS.md` addresses the blocking issues from R001. It now treats the clean macOS profile as an explicit requirement or blocker, validates the current ChatGPT MCP connector UX as a go/no-go dependency, tests the same install path the tutorial will teach, protects the API key/keychain flow, and records timings/papercuts at a useful level of detail.

## Notes for execution

- When recording a clean-profile limitation, make it explicit enough for the maintainer to decide whether to continue or stop; do not silently treat a developer account as equivalent to a fresh user.
- Capture the exact connector JSON that ChatGPT accepts during the run. If it differs from the prompt’s `/Applications/icuvisor.app/Contents/MacOS/icuvisor` snippet because the fallback build-from-source path is used, update the tutorial plan before drafting.
- Keep any API key evidence out of terminal history, screenshots, ChatGPT connector JSON, and repo files. Record only that keychain storage was verified and whether macOS showed a permission prompt.
- For unexpected prompts, prefer recording the shortest tutorial-safe action, and defer failure-mode detail to the footer/troubleshooting link so the tutorial body stays Diataxis-compliant.

No further plan changes are required before executing Step 1.
