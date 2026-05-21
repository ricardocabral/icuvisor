# Code Review — TP-050 Step 2: Install Hextra as a Hugo Module

## Verdict: Request changes

## Findings

### Blocking: Hextra v0.11.0 is incompatible with the pinned deploy Hugo version

`web/go.mod:5` pins `the pinned Hextra module v0.11.0`, but the existing Pages workflow still installs Hugo `0.139.4` (`.github/workflows/pages.yml:28`). Hextra v0.11.0 declares a minimum Hugo version of `0.146.0`, and the site does not build with the workflow-pinned version.

I verified this with the same Hugo version used by CI:

```text
hugo v0.139.4+extended
WARN  Module "the pinned Hextra module" is not compatible with this Hugo version: Min 0.146.0
Error: error building site: .../the pinned Hextra module@v0.11.0/layouts/_markup/render-passthrough.html:4:1: parse failed: template: _markup/render-passthrough.html:4: function "try" not defined
```

The task prompt says the workflow is pinned to `0.139.4` and to keep that version unless Hextra needs newer. This step either needs to pin a Hextra release compatible with Hugo `0.139.4` (for example, Hextra `v0.9.x` declares `min_version = "0.134.0"`) or explicitly update the workflow Hugo pin to at least `0.146.0` in the same change set and record that decision in `STATUS.md`. As-is, landing Step 2 would leave the Pages build broken.

## Checks run

- `cd web && hugo mod graph`
- `cd web && hugo --minify --gc` with local Hugo `0.161.1` — passes
- `cd web && /tmp/hugo-0.139.4-test/hugo --minify --gc` — fails as shown above
