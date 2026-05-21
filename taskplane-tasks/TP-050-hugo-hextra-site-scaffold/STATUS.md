# TP-050-hugo-hextra-site-scaffold — Status

**Current Step:** Step 7: Verify
**Status:** ✅ Complete
**Last Updated:** 2026-05-20
**Review Level:** 2
**Review Counter:** 18
**Iteration:** 1
**Size:** M

---

### Step 1: Decide on landing-page strategy

**Status:** ✅ Complete

- [x] Review the current bespoke landing layout and stylesheet.
- [x] Decide whether to preserve the bespoke landing override or rebuild it with Hextra components.
- [x] Record the landing-page strategy choice and rationale in STATUS.md.

**Landing-page strategy:** choose **A, preserve the bespoke landing page**. Evidence: `web/layouts/index.html` is a complete custom Hugo home template with the current branded hero, download cards, documentation/tool links, release lookup, and footer; `web/static/css/style.css` has substantial project-specific design-system tokens and responsive card styling. Rebuilding this with Hextra components would be design work outside this scaffold. Hextra should own documentation pages while the project-level `web/layouts/index.html` continues to override only the home page through Hugo template precedence; avoid adding `type`/`layout` to `web/content/_index.md` unless build testing proves it necessary. Preserve landing dependencies: `web/static/css/style.css`, `web/layouts/partials/brand.html`, `web/layouts/partials/tool-list.html`, `web/data/tools.json`, `params.featuredTools`, `params.github`, `params.intervals`, and favicon/static assets. After Hextra installation, verify `/` uses the bespoke layout and section pages use Hextra.

### Step 2: Install Hextra as a Hugo Module

**Status:** ✅ Complete

- [x] Check whether the Hextra Hugo Module scaffold already exists in the current branch.
  Evidence: `web/hugo.toml` exists and already has local Hugo module mounts, but `web/go.mod` and `web/go.sum` are absent and grep found no Hextra module import or Pagefind/search config. The scaffold is therefore only partially present and needs minimum missing Hextra module work.
- [x] Initialize `web/go.mod` for the Hugo module if missing.
- [x] Add the Hextra module import and preserve required local mounts in `web/hugo.toml`.
- [x] Fetch and pin Hextra to a released tag, producing `web/go.sum`.
- [x] Record the Hextra pinning/version evidence in STATUS.md.
  Evidence: `web/go.mod` now requires `the pinned Hextra module v0.9.7`, a released tag rather than `latest`; `web/go.sum` is present after `hugo mod tidy`.
- [x] R004 revision: pin Hextra to a release compatible with the workflow-pinned Hugo 0.139.4, or update the workflow Hugo pin and record the decision.
  Evidence: downgraded to Hextra v0.9.7 to keep the Pages workflow on Hugo 0.139.4; verified with downloaded `hugo v0.139.4+extended` that `cd web && hugo --minify --gc` builds successfully.

### Step 3: Configure navigation, theme params, Pagefind

**Status:** ✅ Complete

- [x] Check existing navigation/search/theme/footer config before changing it.
  Evidence: `web/hugo.toml` had no `menu.main`, `params.search`, `params.navbar`, `params.theme`, or Hextra footer params; local `_default/list.html` and `_default/single.html` were still overriding docs pages; Hextra v0.9.7 provides FlexSearch only, so Pagefind needs explicit partial/indexing integration.
- [x] Route documentation pages through Hextra layouts while preserving only the bespoke home-page override.
- [x] Configure Hextra top-level menu entries and brand/theme params in `web/hugo.toml`.
- [x] Enable Pagefind search with Hextra v0.9.7 by adding small Pagefind UI/script partial overrides and an explicit Pagefind indexing step later in the Pages workflow.
  Evidence: `web/layouts/partials/search.html`, `web/layouts/partials/scripts/search.html`, and `web/layouts/partials/custom/head-end.html` render Pagefind UI hooks/assets on Hextra pages; `web/public/install/index.html` contains `.icuvisor-pagefind-search` and `/pagefind/pagefind-ui.js` after a Hugo build.
- [x] Configure footer links/text to preserve the current legal/footer line.
- [x] Verify build warnings introduced or exposed by the theme configuration.
  Evidence: the workflow-pinned `hugo v0.139.4+extended` builds Step 3 with zero warnings. Newer local Hugo v0.161.1 emits upstream Hextra v0.9.7 deprecation warnings for `site.Data` and `.Language.LanguageDirection`; those cannot be removed without breaking Hugo 0.139 compatibility or forking larger theme templates.
- [x] R008 revision: add a local Hextra `favicons.html` partial so docs pages use icuvisor favicon assets instead of Hextra defaults.
- [x] R008 revision: prove Pagefind UI assets resolve by running the explicit Pagefind indexing step after Hugo and documenting that Step 5 will add it to the workflow.
  Evidence: `cd web && hugo --minify --gc && npx --yes pagefind --site public` generated `public/pagefind/pagefind-ui.js`, `public/pagefind/pagefind-ui.css`, and `public/pagefind/pagefind.js`; Step 5 will add the same explicit indexing command to CI before the smoke check.

**Step 3 plan after R006:** Keep Hextra pinned to v0.9.7 for Hugo 0.139.4 compatibility, but do not substitute FlexSearch for Pagefind. Use Hextra's `menu.main` search pseudo-entry (`params.type = "search"`) so the header reserves the search slot, override `layouts/partials/search.html` and `layouts/partials/scripts/search.html` to load Pagefind UI from `/pagefind/`, and add explicit Pagefind indexing in Step 5 so `web/public/pagefind/` exists. Remove the local generic `_default/list.html` and `_default/single.html` templates so documentation pages use Hextra layouts; keep `web/layouts/index.html` as the only bespoke landing override. Add a small `layouts/partials/custom/footer.html` and disable Hextra's default footer copyright/powered-by bits while preserving the MIT/not-affiliated line plus GitHub, Security, and Contribute links. Remove `sitemap` from `disableKinds` unless testing proves it must stay disabled; keep `enableRobotsTXT = true`. Verify a rendered docs page, not just `/`, includes Hextra nav, the Pagefind search container, and custom footer.

### Step 4: Create empty section indexes

**Status:** ✅ Complete

- [x] Verify the six section index files that already exist in the current branch.
  Evidence: all six required files existed (`install`, `connect`, `tutorials`, `guides`, `reference`, `explain`), but they lacked required weights and several contained body markdown from preexisting content migration work.
- [x] Convert the six section index files to frontmatter-only scaffold pages with required weights and minimal Hextra metadata.
- [x] Confirm child content pages are left untouched and the bespoke landing `web/content/_index.md` remains unchanged.
  Evidence: `git diff --name-only` for Step 4 lists only the six section `_index.md` files plus `STATUS.md`; no child content pages or root `web/content/_index.md` were modified.
- [x] Run a Hugo build after the frontmatter changes to catch frontmatter, relref, and layout issues.
  Evidence: downloaded workflow-pinned `hugo v0.139.4+extended` and ran `cd web && hugo --minify --gc`; build completed successfully with zero warnings.

**Step 4 plan after R011:** Update exactly `install`, `connect`, `tutorials`, `guides`, `reference`, and `explain` section `_index.md` files to frontmatter-only scaffold pages with weights 10/20/30/40/50/60. Remove body markdown from those six index files only; do not touch child content pages or the root landing index.

### Step 5: Pages workflow

**Status:** ✅ Complete

- [x] Inspect the existing Pages workflow and confirm the Hugo version strategy.
  Evidence: `.github/workflows/pages.yml` pins `HUGO_VERSION: 0.139.4`; Step 2 pinned Hextra v0.9.7 specifically to keep that workflow version unchanged.
- [x] Add Hugo Module dependency resolution before the Hugo build.
- [x] Add explicit Pagefind indexing after the Hugo build.
- [x] Add a smoke check for `web/public/index.html` and `web/public/pagefind` before upload.
- [x] Verify the updated workflow commands locally.
  Evidence: with downloaded workflow-pinned `hugo v0.139.4+extended`, ran `hugo mod get the pinned Hextra module@v0.9.7`, `hugo mod tidy`, `hugo --minify --gc`, `npx --yes pagefind --site public`, and smoke checks for `public/index.html` plus `public/pagefind`; all passed and `web/go.mod`/`web/go.sum` remained unchanged.

### Step 6: Local preview docs

**Status:** ✅ Complete

- [x] Review the current `web/README.md` preview/deploy instructions.
- [x] Rewrite local preview commands for Hugo Modules and Pagefind indexing.
- [x] Document the minimum Hugo extended version, current Hextra pin, and upgrade command.
- [x] R017 revision: bind the documented Python static preview server to `127.0.0.1` by default.

### Step 7: Verify

**Status:** ✅ Complete

- [x] Build the site with the workflow-pinned Hugo version and zero warnings.
  Evidence: downloaded `hugo v0.139.4+extended` and ran `cd web && hugo --minify --gc --printPathWarnings --panicOnWarning`; build completed successfully with no warnings.
- [x] Generate and verify the Pagefind index under `web/public/pagefind/`.
  Evidence: `cd web && npx --yes pagefind --site public` indexed 31 HTML files and produced `public/pagefind/pagefind-ui.js`, `public/pagefind/pagefind-ui.css`, and `public/pagefind/pagefind.js`.
- [x] Verify the rendered landing page, section indexes, Hextra navigation, footer, and search UI hooks.
  Evidence: checked `web/public/index.html` for the bespoke `site-header` and `css/style.css`; checked all six section `index.html` files exist; checked `web/public/install/index.html` for Hextra `nav-container`, Install/Connect/Tutorials/Guides/Reference/Explain/GitHub menu text, `.icuvisor-pagefind-search`, `pagefind-ui.js`, and the MIT/not-affiliated footer text.
- [x] Run a broken-link/path-warning check with Hugo.
  Evidence: the verification build used `--printPathWarnings --panicOnWarning`; Hugo reported no broken internal links/path warnings and exited successfully.
- [x] Record workflow verification limits (`act` or PR preview availability).
  Evidence: ran `act -W .github/workflows/pages.yml -j build -n`; dry-run planned and completed the build job steps including Hugo install, module resolution, Hugo build, Pagefind build, smoke test, and artifact upload. A full `act` execution was not run because the prompt marks it optional and the workflow commands were already executed locally with the pinned Hugo binary.

| 2026-05-20 10:18 | Task started | Runtime V2 lane-runner execution |
| 2026-05-20 10:18 | Step 1 started | Decide on landing-page strategy |
| 2026-05-20 10:20 | Review R001 | plan Step 1: UNKNOWN |
| 2026-05-20 10:22 | Review R002 | code Step 1: APPROVE |
| 2026-05-20 10:23 | Review R003 | plan Step 2: APPROVE |
| 2026-05-20 10:27 | Review R004 | code Step 2: UNKNOWN |
| 2026-05-20 10:32 | Review R005 | code Step 2: APPROVE |
| 2026-05-20 10:35 | Review R006 | plan Step 3: UNKNOWN |
| 2026-05-20 10:39 | Review R007 | plan Step 3: APPROVE |
| 2026-05-20 10:52 | Review R008 | code Step 3: UNKNOWN |
| 2026-05-20 11:17 | Review R010 | code Step 3: APPROVE |
| 2026-05-20 11:19 | Review R011 | plan Step 4: REVISE |
| 2026-05-20 11:21 | Review R012 | plan Step 4: APPROVE |
| 2026-05-20 11:24 | Review R013 | code Step 4: APPROVE |
| 2026-05-20 11:27 | Review R014 | plan Step 5: APPROVE |
| 2026-05-20 11:31 | Review R015 | code Step 5: APPROVE |
| 2026-05-20 11:32 | Review R016 | plan Step 6: APPROVE |
| 2026-05-20 11:35 | Review R017 | code Step 6: REVISE |
| 2026-05-20 11:37 | Review R018 | code Step 6: APPROVE |

| 2026-05-20 11:42 | Agent reply | Acknowledged. I will verify the existing Hugo/Hextra scaffold first, document evidence in STATUS.md, run verification, and only make minimum missing changes if needed. Per current worker instructions  |
| 2026-05-20 11:42 | ⚠️ Steering | Before making changes for TP-050, verify whether the Hugo/Hextra scaffold work already exists in the current branch. Supervisor found an older event where TP-050 completed in batch 20260516T174422, bu |
| 2026-05-20 11:42 | Worker iter 1 | done in 5020s, tools: 243 |
| 2026-05-20 11:42 | Task complete | .DONE created |