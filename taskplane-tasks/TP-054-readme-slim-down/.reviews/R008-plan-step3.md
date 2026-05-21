# Plan Review R008 — Step 3: Rewrite README

## Verdict

Approved.

The Step 3 plan is sufficient for this focused README rewrite. It references the prompt's target structure, preserves the existing badges and project-layout block, and explicitly checks concision/no deleted-doc links. Step 2 has already narrowed the remaining live deleted-doc links to README entries that this rewrite is supposed to remove.

## Verification

I reviewed:

- `taskplane-tasks/TP-054-readme-slim-down/PROMPT.md`
- `taskplane-tasks/TP-054-readme-slim-down/STATUS.md`
- current `README.md`
- current `CHANGELOG.md`
- current `Makefile` targets

I also re-ran the scoped and broader product-link sweeps. The remaining hits are the expected README deferrals, `docs/install/macos.md` links that Step 4 deletion will remove, and the documented false positive for the kept `docs/threat-models/coach-mode.md` file.

## Implementation reminders

- Follow the prompt's target structure closely; the final README should be roughly 60-100 lines and developer-focused.
- Keep all nine existing badges.
- Preserve the current `Project layout` code block verbatim, as required by the task prompt.
- The development target list should include `make docs-tools`; this is a current Makefile target but not present in the old README block.
- Remove the old standalone end-user sections and deleted-doc links rather than moving them around:
  - Features
  - MCP tool catalog/resources/prompts
  - macOS install and quickstart/API-key prose
  - MCP transport
  - delete/write safety and toolset tiers
  - post-upgrade guidance
  - Claude Desktop/Claude Code user setup links
- Keep the Codex local doc only if referenced as developer validation; do not point at deleted Claude docs.
- Include developer pointers to `CONTRIBUTING.md`, `SECURITY.md`, `ROADMAP.md`, and `docs/prd/PRD-icuvisor.md` as shown in the target structure.
- After the rewrite, run the same scoped grep before moving on so Step 3 catches any accidental stale README link before Step 4 deletion.

No blocking changes are needed before implementation.
