# icuvisor.app website

Hugo source for the public site at <https://icuvisor.app>.

## Local preview

```bash
cd web
hugo server -D
```

Hugo 0.128+ recommended.

## Layout

- `hugo.toml` — site config (baseURL is `https://icuvisor.app/`).
- `layouts/index.html` — single-page custom layout, no theme dependency.
- `static/css/style.css` — bento-grid design tokens (see `.claude/skills/design-system/`).
- `static/CNAME` — published to the GitHub Pages root so the apex domain resolves.
- `content/_index.md` — placeholder so Hugo renders the home template.

## Deploy

`.github/workflows/pages.yml` builds the site on every push to `main` that touches `web/**` and publishes it to GitHub Pages. Configure the repo's Pages settings to source from "GitHub Actions" and add `icuvisor.app` as the custom domain.
