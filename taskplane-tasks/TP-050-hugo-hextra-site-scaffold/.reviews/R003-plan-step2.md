# Plan Review — TP-050 Step 2: Install Hextra as a Hugo Module

## Verdict: Approved

The Step 2 plan is focused and covers the important installation risks: detect any existing scaffold, initialize a nested Hugo module under `web/`, preserve the current local module mounts in `web/hugo.toml`, add Hextra as a module import, fetch it as a tagged dependency, and record the pinning evidence in `STATUS.md`.

## Implementation guardrails

- Use a tagged version explicitly, for example `hugo mod get the pinned Hextra module@<released-tag>`. Do not rely on an unqualified `-u`/`latest` result; the task explicitly requires a stable released tag.
- Keep the existing `[module]` mounts in `web/hugo.toml`, especially `content`, `static`, `layouts`, `data`, `assets`, and `../docs/brand -> assets/brand`. These preserve the bespoke landing page and brand assets chosen in Step 1.
- Initialize from `web/` with the module path from the prompt: `the local web module`. A nested `web/go.mod` is appropriate here and should not alter the root Go module.
- Commit both `web/go.mod` and `web/go.sum` if generated. If `go.sum` is unexpectedly not generated, note why in `STATUS.md` rather than fabricating it.
- After adding the import, run at least a quick `cd web && hugo mod graph` or `hugo mod vendor --help >/dev/null`-style Hugo module sanity check if a full site build is deferred to later steps. Full rendering/search validation can remain in Step 7.

## Notes

Step 3 will handle navigation, theme params, and Pagefind, so it is fine that this Step 2 plan does not attempt to configure Hextra behavior beyond the module import. The only blocker to avoid during execution is accidentally using a floating Hextra dependency instead of recording a concrete released version.
