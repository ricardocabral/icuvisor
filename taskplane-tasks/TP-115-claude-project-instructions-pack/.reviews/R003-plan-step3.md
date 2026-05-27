# Plan Review R003 — Step 3

Verdict: **Approved**

The Step 3 plan is appropriate for the task. Linking the new guide from the guides index and Claude connection docs, then adding short pointers from the weekly review/readiness/taper/prompt-library pages, is the right deduplication strategy and avoids copying the full Project-instruction blocks into recipe pages.

Implementation notes:

- Add a `claude-project-instructions` card/link in `web/content/guides/_index.md` so the new guide is discoverable from the Guides landing page.
- In Claude Desktop/Claude Code docs, keep the link near the “start a new chat / stale catalog / after setup” guidance rather than mixing Project instructions with credential configuration. Reiterate that Project instructions are not a place for API keys, athlete IDs, local paths, or config contents.
- In cookbook pages, use one concise callout or sentence linking to the guide for reusable standing instructions; do not duplicate the base/optional blocks.
- `web/content/reference/resources-prompts.md` only needs a small clarification if you see readers could confuse registered MCP Prompts with client-side Claude Project instructions. The new guide already explains this distinction, so a reference edit is optional rather than required.
- Add a docs-only `[Unreleased]` changelog bullet, since this publishes a new user-facing guide.
- After edits, `make web-build` should catch any broken `relref` targets introduced by the new links.
