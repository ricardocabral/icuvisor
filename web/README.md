# icuvisor.app website

Hugo source for the public site at <https://icuvisor.app>.

## Prerequisites

- Hugo Extended 0.139.4 or newer. The GitHub Pages workflow currently pins 0.139.4.
- Go, for Hugo Modules.
- Node.js/npm, only when generating the Pagefind static search index with `npx --yes pagefind`.

## Local preview

Fetch the pinned Hugo Module dependency, then start Hugo's development server:

```bash
cd web
hugo mod get github.com/imfing/hextra@v0.9.7
hugo mod tidy
hugo server -D
```

`hugo server` is the fastest way to preview layout and content changes. Pagefind is a static post-build index, so the search UI is fully testable from a production-style build:

```bash
cd web
hugo --minify --gc
npx --yes pagefind --site public
python3 -m http.server --bind 127.0.0.1 --directory public 1313
```

Open <http://localhost:1313> after the static server starts.

## Theme dependency

Hextra is installed as a Hugo Module and pinned in `go.mod`:

```text
github.com/imfing/hextra v0.9.7
```

The pin is intentional because v0.9.7 is compatible with the workflow-pinned Hugo 0.139.4. To upgrade Hextra, choose a released tag and test it against the Pages Hugo version:

```bash
cd web
hugo mod get -u github.com/imfing/hextra@<tag>
hugo mod tidy
hugo --minify --gc
npx --yes pagefind --site public
```

If the new Hextra tag requires a newer Hugo version, update `.github/workflows/pages.yml` in the same change.

## Layout

- `hugo.toml` — site config, Hextra module import, menu, theme params, and Pagefind settings.
- `go.mod` / `go.sum` — Hugo Module manifests for the Hextra dependency.
- `data/tools.json` — generated from the Go tool registry by `make docs-tools` / `cmd/gendocs`; do not hand-edit.
- `layouts/index.html` — bespoke landing-page layout that overrides only the home page.
- `layouts/partials/` — small Hextra overrides for Pagefind, favicons, and footer text.
- `static/css/style.css` — bento-grid design tokens for the bespoke landing page.
- `static/CNAME` — published to the GitHub Pages root so the apex domain resolves.
- `content/_index.md` — placeholder so Hugo renders the home template.

## Deploy

`.github/workflows/pages.yml` builds the site on every push to `main` that touches `web/**` and publishes it to GitHub Pages. The workflow resolves Hugo Modules, runs `hugo --minify --gc`, generates the Pagefind index under `web/public/pagefind/`, smoke-tests the artifact, and uploads `web/public` to Pages. Configure the repo's Pages settings to source from "GitHub Actions" and add `icuvisor.app` as the custom domain.
