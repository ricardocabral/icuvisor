# Review R003 — Step 3 Plan

**Verdict:** APPROVE with minor follow-ups

The Step 3 scope is small and appropriate: link the new tutorial from discovery surfaces without changing runtime behavior or duplicating the walkthrough content. The existing site structure supports this via a tutorial-index card plus short cross-links from the Claude connection pages.

Follow-ups for implementation:

- Add a `garmin-to-claude` card to `web/content/tutorials/_index.md` so the new tutorial is discoverable from the tutorials landing page.
- Add short pointers from both `web/content/connect/claude-desktop.md` and `web/content/connect/claude-code.md`, preferably near the verification/smoke-check sections, using language that preserves the source chain: device provider syncs to intervals.icu; icuvisor reads intervals.icu; Claude calls icuvisor.
- Update `CHANGELOG.md` under `[Unreleased]` / `Added`; this is a user-facing documentation addition.
- Treat homepage and cookbook links as optional. The homepage already links to Tutorials, so avoid adding another homepage CTA unless it stays very focused. A one-line prompt-library pointer is acceptable if it helps users move from setup to prompts, but avoid copying tutorial prompts into the cookbook.
- Use Hugo `relref`/existing card patterns consistently and leave link/image validation to Step 4's `make web-build`.

No blocking issues found for proceeding to Step 3.
