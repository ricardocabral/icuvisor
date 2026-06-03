# OpenAPI endpoint-diff triage

icuvisor keeps a pinned intervals.icu OpenAPI path baseline in `scripts/openapidiff/baseline/intervals-openapi.json`. Maintainers use it to spot newly documented or removed upstream endpoint paths and then decide whether the change deserves a Taskplane/backlog item.

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
  -latest-url https://intervals.icu/api/openapi.json \
  -output openapi-endpoint-diff.md
```

Normal tests never fetch the network; the tool's tests use in-memory fixture specs. The scheduled/manual GitHub workflow is the only committed automation that performs a live fetch.

## Triage process

1. Read the generated Markdown report. It compares only OpenAPI `paths` keys and is not a product-scope decision.
2. For each added path, inspect the public intervals.icu API docs and decide whether it fits the PRD/roadmap.
3. If a path is relevant, create a focused Taskplane/backlog task. Include the endpoint path, HTTP method, terse/full response-shaping expectations, safety/delete-mode impact, and fixture requirements.
4. For removed paths, check whether existing icuvisor tools depend on the path and open a regression task if needed.
5. After the triage decision is recorded, intentionally update `scripts/openapidiff/baseline/intervals-openapi.json` to the latest accepted path inventory so future reports show only new drift.

Do not auto-generate tools from this report. New endpoints still need normal clean-room implementation, tests, docs, and review.
