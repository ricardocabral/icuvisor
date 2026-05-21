# Plan Review: Step 4 — Docs and verification

## Verdict: Request changes

The Step 4 plan is still the original high-level checklist. For a docs/verification step that updates generated tool reference content and runs release-quality checks, it needs a little more precision before implementation.

## Findings

1. **The documentation target is ambiguous.**
   - `web/content/reference/tools.md` is only a generated-reference wrapper using `{{< tool-catalog >}}`; the visible tool wording comes from the registry-generated data (`web/data/tools.json`) and ultimately the Go tool description/schema in `internal/tools/get_wellness_data.go`.
   - The plan should state whether Step 4 will update the `get_wellness_data` registry wording/output schema to mention source-specific `_meta.provenance.<field>.native_scale` labels, then run `make docs-tools` to refresh `web/data/tools.json`.
   - Avoid manually editing generated/reference output in a way that will be overwritten or leave the catalog stale.

2. **The changelog update needs an explicit user-facing entry.**
   - Add a planned `[Unreleased]` entry, likely under `Changed`, saying that `get_wellness_data` provenance now reports provider-native sleep/readiness scale labels for Garmin, WHOOP, Oura, and Polar, with unknown sources reported as `unknown`.
   - This is a response-shape/user-facing semantics change, so the changelog item should be concrete enough for users to understand the compatibility impact.

3. **“Run full quality gate” should be spelled out and reconciled with Step 5.**
   - The task completion criteria require `make test`, `make build`, and `make lint`; Step 5 repeats the same gates. The Step 4 plan should either list the exact commands to run now and record their results in `STATUS.md`, or explicitly defer the full gate to Step 5 while running only docs/catalog checks in Step 4.
   - If tool catalog wording changes, include `make docs-tools` and a stale-docs check such as reviewing `git diff`/generated output before the full gate.

## Suggested Step 4 acceptance criteria

Before implementation, expand Step 4 in `STATUS.md` with something like:

- Update `get_wellness_data` docs/source wording so the generated reference explains that `_meta.provenance` carries provider-native `native_scale` labels for supported providers and `unknown` when unresolved, without changing canonical `_meta.scales.sleepScore` wording.
- Run `make docs-tools` if registry wording changes, and include the generated `web/data/tools.json` diff.
- Add a `[Unreleased]` changelog entry under `Changed` for the provenance native-scale behavior.
- Run and record exact verification commands (`make test`, `make build`, `make lint`, or explicitly defer these to Step 5 and state what Step 4 verifies instead).
