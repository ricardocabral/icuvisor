# R011 Plan Review — Step 5: Makefile + CI guard

Decision: APPROVE

The Step 5 plan matches the task prompt and fits the current repository structure. `cmd/gendocs` already supports `--out web/data/tools.json`, the Makefile uses documented `##` notes to populate `make help`, and `.github/workflows/ci.yml` already has an Ubuntu-only `tool-catalog-guards` job with Go set up, which is an appropriate place for the generated-catalog freshness check.

Implementation notes to keep the plan acceptance-ready:

- Add `docs-tools` to the Makefile `.PHONY` list.
- Implement the target as `$(GO) run ./cmd/gendocs --out web/data/tools.json` so it respects the existing configurable `GO` variable and remains idempotent.
- Add a `##` help note to the target; no separate help wiring is needed because the current `help` target discovers those notes automatically.
- In CI, prefer adding a step to the existing `tool-catalog-guards` job after Go setup / current catalog checks. No Hugo setup is required for this guard.
- Make the CI failure actionable and scoped, for example:
  - run `make docs-tools`
  - run `git diff --exit-code -- web/data/tools.json`, or wrap it with an `::error::web/data/tools.json is stale; run make docs-tools and commit the result` message plus `git diff -- web/data/tools.json` before exiting.
- Keep the diff path-scoped to `web/data/tools.json`; Step 5 should not fail because of unrelated generated or local changes.

No blockers found.
