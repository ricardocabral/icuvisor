# OpenAPI endpoint-diff triage

icuvisor keeps a pinned intervals.icu OpenAPI path and schema-name baseline in `scripts/openapidiff/baseline/intervals-openapi.json`. Maintainers use it to spot newly documented or removed upstream endpoint paths and schema names, then decide whether the change deserves a Taskplane/backlog item. The scheduled workflow publishes the same Markdown classification in the GitHub step summary and as the `openapi-structural-drift-report` artifact.

## Classification contract

This check is intentionally key-level:

- **Structural drift:** added or removed OpenAPI `paths` keys, or added or removed `components.schemas` names. These changes require human triage before the pinned baseline is refreshed.
- **No structural key drift:** the path-key inventory and schema-name inventory are unchanged. `info`, descriptions, summaries, examples, ordering, and formatting may still differ between specs.
- **Deferred semantic drift:** method changes, operation IDs, parameters, request/response shapes, enum values, and fields inside existing schemas are not evaluated by this report. Do not describe a clean report as proof that those dimensions are unchanged.

Run an offline comparison with local specs:

```sh
go run ./scripts/openapidiff \
  -baseline scripts/openapidiff/baseline/intervals-openapi.json \
  -latest /path/to/latest-intervals-openapi.json \
  -output openapi-endpoint-diff.md
```

Run an opt-in live comparison by providing the upstream OpenAPI JSON URL explicitly:

```sh
go run ./scripts/openapidiff \
  -baseline scripts/openapidiff/baseline/intervals-openapi.json \
  -latest-url https://intervals.icu/api/v1/docs \
  -output openapi-endpoint-diff.md
```

Normal tests never fetch the network; the tool's tests use in-memory fixture specs. The scheduled/manual GitHub workflow is the only committed automation that performs a live fetch.

## Triage process

1. Read the generated Markdown report. It compares OpenAPI `paths` keys and `components.schemas` names. It is a human triage aid, not product approval or a product-scope decision; a "no structural key drift" classification only means this narrow inventory check found no path/schema-name additions or removals.
2. For each added path or schema name, inspect the public intervals.icu API docs and decide whether it fits the PRD/roadmap. For a clean key-level report, still use normal maintainer judgment if a separate upstream note suggests semantic changes inside an existing endpoint or schema.
3. If a path or schema change is relevant, create a focused Taskplane/backlog task. Include the endpoint path, HTTP method, terse/full response-shaping expectations, safety/delete-mode impact, schema/model impact, and fixture requirements.
4. For removed paths or schema names, check whether existing icuvisor tools depend on them and open a regression task if needed.
5. After the triage decision is recorded, intentionally update `scripts/openapidiff/baseline/intervals-openapi.json` to the latest accepted path and schema-name inventory so future reports show only new drift.

Do not auto-generate tools from this report. New endpoints still need normal clean-room implementation, tests, docs, and review.
