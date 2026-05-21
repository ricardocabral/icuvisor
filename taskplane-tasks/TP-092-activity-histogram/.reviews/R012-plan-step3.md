# R012 plan review — Step 3: Tool registration and activation hint

Verdict: APPROVE

The Step 3 checklist is narrow enough to implement from the current Step 2 state. The new tool constructor and handler already exist, so this step should stay focused on making `get_activity_histogram` discoverable through the runtime registry, shared tool catalog, toolset tiering, coach ACL surface, and generated docs while preserving the analyzer metadata contract.

## Approval notes

- Register the tool through the existing shared registration path, not a parallel catalog list: add `newGetActivityHistogramTool(client, client, client, ...)` to `registryBaseTools` and keep it as `fullTool`/read-only. This should make the runtime registry and `tools.Catalog()` see the same descriptor.
- Add the name to `internal/toolcatalog/catalog.go` as an athlete-scoped tool. Otherwise `defaultRegistry.Register` will reject it as unknown, and coach mode will not inject/resolve `athlete_id` correctly.
- Add the activity-group mapping in `toolCatalogGroup` and update tier expectations such as `TestRegisteredToolTierMembership` so the new tool is explicitly `full`, not implicitly accepted by default behavior.
- Update generated catalog artifacts with the generator (`make docs-tools`) rather than hand-editing website tables. In this repo that means at least `web/data/tools.json` and `cmd/gendocs/testdata/tools.golden.json`; `web/content/reference/tools.md` should only change if the rendered page text itself changes.
- Keep the activation hint in the MCP description, not only in docs. The first sentence should lead with the user intent (single-activity power/HR/pace distribution or time-in-bucket prompts), and the description must explicitly say not to pull `get_activity_streams` and bin raw samples manually.
- Preserve the Step 2 analyzer meta semantics: `_meta.method:"activity_stream_histogram"`, `_meta.source_tools` naming the upstream reads, and `_meta.n` as the number of valid weighted stream intervals contributing to buckets. Do not treat `n` as bucket count or raw sample count.

## Suggested targeted verification for this step

- `go test ./internal/toolcatalog ./internal/tools ./cmd/gendocs`
- `make docs-tools` followed by `git diff -- web/data/tools.json cmd/gendocs/testdata/tools.golden.json`

Broader histogram behavior tests can remain in Step 4, but Step 3 should not leave catalog/registry tests failing.
