# Review R011 — Plan Step 4: Create empty section indexes

## Verdict: Changes requested

The Step 4 plan should be revised before implementation.

## Blocking issue

The plan explicitly says it will preserve existing section body content and record why it was preserved:

> Add the required section weights and Hextra display metadata without deleting preexisting content.
> Record why existing section body content was preserved instead of emptying files.

That conflicts with the task requirements for this step:

- Step 4: section `_index.md` files should have **only frontmatter**; body is intentionally empty.
- Do NOT: “section indexes stay empty.”
- File scope: the section index pages are scaffold-only; content migration is TP-052/TP-053.

Current files such as `web/content/install/_index.md`, `connect/_index.md`, `guides/_index.md`, `explain/_index.md`, and `tutorials/_index.md` contain end-user body content. Preserving that body content would fail the requested scaffold behavior even if the content already exists in the worktree.

## Required plan adjustment

Revise Step 4 to:

1. Update exactly these six section index files to be frontmatter-only:
   - `web/content/install/_index.md` — weight `10`
   - `web/content/connect/_index.md` — weight `20`
   - `web/content/tutorials/_index.md` — weight `30`
   - `web/content/guides/_index.md` — weight `40`
   - `web/content/reference/_index.md` — weight `50`
   - `web/content/explain/_index.md` — weight `60`
2. Keep `title`, `description`, `weight`, and any minimal Hextra display/cascade frontmatter needed for section behavior.
3. Remove body markdown from those six `_index.md` files.
4. Do not touch child content pages in this step unless the task owner explicitly decides to clean up out-of-scope preexisting migrated content separately.
5. Leave `web/content/_index.md` as the bespoke landing page index.
6. Run a Hugo build after the frontmatter changes to catch TOML/YAML/relref/layout issues.

## Minor note

The prompt text says “five Diataxis sections” in one place, but the required navigation and weights list six top-level sections including `install` and `connect`. The plan should state six files to avoid ambiguity.
