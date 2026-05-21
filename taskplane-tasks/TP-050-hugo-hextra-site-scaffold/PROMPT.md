# TP-050 — Hugo + Hextra site scaffold with Pagefind search

## Mission

Replace the current single-page Hugo site at `web/` with a structured documentation site built on the Hextra theme, organized along Diataxis lines (tutorials / how-to / reference / explanation), with Pagefind static search. This is the foundation for moving end-user documentation onto **icuvisor.app**, so that the source tree can host developer-/contributor-facing material only.

This task lands the **scaffold + theme + navigation + search + deploy pipeline only**. Content migration is TP-052; the tool-catalog generator is TP-051; the getting-started tutorial is TP-053; the README rewrite is TP-054.

Project context (icuvisor is currently private, no external links to break, no migration notices needed):

- Today `web/` is a hand-rolled single page (`layouts/index.html` + `static/css/style.css`) plus a stub `content/_index.md`. Deploy is via `.github/workflows/pages.yml` to GitHub Pages with custom domain `icuvisor.app` (CNAME in `static/`).
- Hugo version pinned in workflow: `0.139.4` extended. Hextra requires Hugo extended ≥ 0.134; keep the pinned version unless Hextra needs newer.

PRD anchors: §7.4 KR1 (install success / discoverability), §6 KR5 (token efficiency — long-form content lives off-binary, on the website).

ROADMAP positioning: docs polish supporting v0.5 internal beta and v1.0 public launch.

Complexity: Blast radius 2 (web/ only, no Go code), Pattern novelty 3 (first time using Hextra in this repo), Security 1, Reversibility 2 = 8 → Review Level 2. Size: M.

## Dependencies

None blocking. The other doc-refactor tasks (TP-051, TP-052, TP-053, TP-054) depend on this landing first.

## Context to Read First

- `web/hugo.toml`, `web/layouts/index.html`, `web/static/css/style.css`, `web/content/_index.md`, `web/README.md`
- `.github/workflows/pages.yml`
- `CLAUDE.md` (esp. "no emojis in commits/PRs" and Conventional Commits rules)
- Hextra docs — installation as a Hugo Module is the recommended path.
- Pagefind: <https://pagefind.app/docs/> — Hextra has built-in Pagefind integration; verify which knob enables it.
- Diataxis overview already captured in this taskplane PR's source plan (tutorials / how-to / reference / explanation).

## File Scope

Expected files (new unless noted):

- `web/hugo.toml` — extend with Hextra module import, menu config, params for theme, Pagefind enabled.
- `web/go.mod`, `web/go.sum` — Hugo Modules manifests for the Hextra dependency.
- `web/content/_index.md` — keep as landing page; either retain the bespoke `layouts/index.html` for the landing only (Hextra supports per-page layout overrides) **or** rebuild the landing with Hextra's hero/feature shortcodes. Pick one and justify in `STATUS.md`. The bespoke landing has design value (see `static/css/style.css` — design tokens from `.claude/skills/design-system/`); preserving it is preferred if Hextra allows a clean override.
- `web/content/install/_index.md` — section landing with frontmatter only (empty body; content is TP-052).
- `web/content/connect/_index.md`
- `web/content/guides/_index.md`
- `web/content/reference/_index.md`
- `web/content/explain/_index.md`
- `web/content/tutorials/_index.md`
- `web/layouts/index.html` — keep or remove depending on landing decision above.
- `web/layouts/partials/` — any small overrides Hextra needs (e.g., custom footer matching current "MIT-licensed · not affiliated with intervals.icu" line).
- `web/static/CNAME` — keep as-is.
- `web/static/favicon.svg` — keep if present.
- `.github/workflows/pages.yml` — add a `hugo mod get` / `hugo mod tidy` step before `hugo --minify --gc`. Confirm Pagefind index is built (Hextra runs Pagefind at build time when enabled).
- `web/README.md` — update local-preview instructions for Hugo Modules (e.g., `hugo mod get -u`, then `hugo server`).
- `taskplane-tasks/TP-050-hugo-hextra-site-scaffold/STATUS.md`.

Out of scope:

- Tool-catalog data file (`web/data/tools.json`) — TP-051.
- Any content under `install/`, `connect/`, `guides/`, `reference/`, `explain/`, `tutorials/` beyond the empty section indexes — TP-052 and TP-053.
- README rewrite — TP-054.
- Deleting `docs/*.md` files migrated elsewhere — TP-052 and TP-054.

## Steps

### Step 1: Decide on landing-page strategy

- [ ] Read `web/layouts/index.html` and `web/static/css/style.css`. Decide:
  - **(A)** Keep the bespoke landing page exactly as-is by setting `layout` / `type` on `content/_index.md` so Hextra renders the home page through the existing layout. Hextra still owns every other page.
  - **(B)** Rebuild the landing using Hextra's hero/feature shortcodes for consistency.
- [ ] Default recommendation: **(A)**. The bespoke landing has explicit design-system styling that Hextra's defaults will not match, and rebuilding it is design work, not docs work. Only choose (B) if (A) requires fighting Hextra.
- [ ] Record the choice and rationale in `STATUS.md`.

### Step 2: Install Hextra as a Hugo Module

- [ ] `cd web && hugo mod init the local web module`.
- [ ] Add the Hextra module import to `hugo.toml`:
  ```toml
  [module]
    [[module.imports]]
      path = "the pinned Hextra module"
  ```
- [ ] `hugo mod get -u the pinned Hextra module`.
- [ ] Verify `go.mod` and `go.sum` are committed.
- [ ] Confirm Hextra version pinning strategy — pin to a released tag, not `latest`, so the site does not silently break. Record the chosen version in `STATUS.md`.

### Step 3: Configure navigation, theme params, Pagefind

- [ ] `hugo.toml`: declare top-level menu (Install · Connect · Guides · Tutorials · Reference · Explain · Source). The Source entry links to the project source page.
- [ ] Enable Pagefind. Hextra exposes a `search.enable` (or equivalent — confirm against Hextra docs) param. Verify search UI appears in the rendered header.
- [ ] Set theme params consistent with the current brand: site title (`icuvisor — Talk to your intervals.icu data`), description (current value in `hugo.toml`), favicon, primary colour roughly matching `--primary` from `static/css/style.css`.
- [ ] Keep `enableRobotsTXT = true`. Keep `disableKinds` minimal; Hextra may want `taxonomy`/`term` enabled — verify before disabling.
- [ ] Configure footer to keep the line "MIT-licensed · not affiliated with intervals.icu" and links to GitHub, SECURITY.md, CONTRIBUTING.md.

### Step 4: Create empty section indexes

- [ ] Add `_index.md` files for the five Diataxis sections under `web/content/`. Each has only frontmatter (title, weight, optional `cascade` to set Hextra display options for the section). Body is intentionally empty; populated by TP-052/053.
- [ ] Section weights: `install=10`, `connect=20`, `tutorials=30`, `guides=40`, `reference=50`, `explain=60`. So navigation reads left-to-right: Install → Connect → Tutorials → Guides → Reference → Explain.

### Step 5: Pages workflow

- [ ] Update `.github/workflows/pages.yml`:
  - Before `hugo --minify --gc`, run `hugo mod get` (and optionally `hugo mod tidy`).
  - Confirm Pagefind output ends up under `web/public/` so the deploy artifact includes it. Hextra runs Pagefind automatically when enabled; if not, add an explicit `npx pagefind --site public` step after the Hugo build.
  - Keep the deploy job unchanged.
- [ ] Add a smoke step: `test -f web/public/index.html && test -d web/public/pagefind` (or equivalent) so a broken search index fails the build.

### Step 6: Local preview docs

- [ ] Rewrite `web/README.md` for the Hugo-Modules workflow:
  ```bash
  cd web
  hugo mod get -u
  hugo server -D
  ```
- [ ] Note the minimum Hugo extended version, where the Hextra version is pinned, and how to upgrade Hextra (`hugo mod get -u the pinned Hextra module@<tag>`).

### Step 7: Verify

- [ ] `cd web && hugo --minify --gc` builds cleanly with zero warnings.
- [ ] `hugo server -D` renders the landing, all five section indexes (even empty), and the search box.
- [ ] Pagefind index is generated under `web/public/pagefind/`.
- [ ] `.github/workflows/pages.yml` passes locally via `act` (optional) or via a PR preview if available.
- [ ] No broken internal links (`hugo` build will warn on these; promote warnings to errors via `--printPathWarnings` or equivalent if Hextra supports it).

## Acceptance Criteria

- `web/` builds via Hugo Modules with Hextra as the theme and Pagefind enabled.
- Landing page renders with current branding (either via preserved bespoke layout or via Hextra shortcodes — whichever was chosen).
- Top-level menu shows: Install, Connect, Tutorials, Guides, Reference, Explain, GitHub.
- Five empty section index pages exist with correct frontmatter; navigation lists each section.
- Pagefind search box is present in the site header and produces a static index under `public/pagefind/` at build time.
- `.github/workflows/pages.yml` builds the new site successfully, including the Hextra module fetch and Pagefind index.
- `web/README.md` documents the new local-preview workflow.
- No regression in deploy: `https://icuvisor.app/` still serves a valid landing page.

## Do NOT

- Do not add any end-user documentation content in this task — section indexes stay empty. Content lands in TP-052 / TP-053.
- Do not delete `docs/*.md` files in this task — handled by TP-052 / TP-054.
- Do not commit a Hugo theme as a vendored copy under `web/themes/`. Use Hugo Modules.
- Do not pin Hextra to `latest`; pin to a tagged release.
- Do not introduce a JavaScript build step (npm install, webpack, etc.) unless Pagefind requires it. Prefer the Hugo-Modules-only path.
- Do not change the GitHub Pages deploy target or the CNAME.
- Do not add emojis to commits, PR descriptions, or rendered site content (per CLAUDE.md).

## Documentation

Must update:

- `STATUS.md` — landing-page strategy decision, Hextra version pinned, Pagefind verification notes.
- `web/README.md` — local-preview instructions.
- `CHANGELOG.md` under `[Unreleased]` if the live site URL changes user-visible behaviour (unlikely for scaffolding; if added, mention "Site now built with Hextra theme and full-text search.").

## Git Commit Convention

Commit at step boundaries with messages prefixed by `TP-050`, for example: `TP-050 add Hextra Hugo module and section indexes`. Conventional Commits, lowercase imperative subject.

---

## Amendments

_Add amendments below this line only._
