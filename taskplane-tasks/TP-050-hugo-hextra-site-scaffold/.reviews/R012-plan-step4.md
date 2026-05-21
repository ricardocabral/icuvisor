# Review R012 — Plan Step 4: Create empty section indexes

## Verdict: Approved

The revised Step 4 plan now matches the task requirements and addresses the blocking issue from R011.

## What looks good

- It scopes the change to exactly the six top-level section indexes:
  - `web/content/install/_index.md`
  - `web/content/connect/_index.md`
  - `web/content/tutorials/_index.md`
  - `web/content/guides/_index.md`
  - `web/content/reference/_index.md`
  - `web/content/explain/_index.md`
- It uses the required section ordering: `10/20/30/40/50/60` for Install, Connect, Tutorials, Guides, Reference, Explain.
- It explicitly removes body markdown from those section indexes, which satisfies the scaffold-only requirement and keeps content migration out of TP-050.
- It leaves child pages and the root landing `web/content/_index.md` untouched, preserving the Step 1 landing-page decision and avoiding out-of-scope cleanup.
- It includes a Hugo build after the frontmatter edits, which is the right verification for malformed frontmatter or layout regressions.

## Non-blocking implementation note

When converting the files, keep the frontmatter useful but minimal: `title`, `description`, `weight`, and only any Hextra-specific display/cascade fields that are actually needed. Avoid adding placeholder body text, summaries, tables, or migration content in this step.
