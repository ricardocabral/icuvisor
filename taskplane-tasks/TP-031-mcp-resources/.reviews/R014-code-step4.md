# Code Review R014 — Step 4: `icuvisor://custom-item-schemas`

**Verdict: REVISE**

I reviewed the diff from `23cb90aa0a443ae563baea48d1643ad7831029a3..HEAD`, read the changed files, and ran:

- `go test ./internal/customitemschemas ./internal/resources ./internal/mcp ./internal/tools`
- `go test ./...`

Both test commands passed, but I found issues that should be fixed before approving Step 4.

## Findings

### 1. `mustSample` introduces a library panic outside `main`

- **File:** `internal/customitemschemas/descriptors.go:71-75`
- **Severity:** Must fix

`mustSample` calls `panic(err)` if a descriptor JSON sample is malformed. The repository hard rule is **“No `panic` outside `main`; return errors.”** This is in an internal library package and can be reached by resource generation/tests, so it should not rely on panic even for static descriptor data.

Suggested fixes:

- Prefer plain Go `map[string]any` / `[]any` literals for the static samples so there is no runtime JSON parse path and no panic helper; or
- Change the descriptor-loading API to return an error, and let `CustomItemSchemasMarkdown()` wrap/report it like the other resource-generation errors.

### 2. The resource is marked as per-`item_type`, but the implementation only emits one schema per broad family

- **File:** `internal/customitemschemas/descriptors.go:21-34`, `internal/resources/custom_item_schemas.go:51-80`
- **Severity:** Should fix

Step 4’s acceptance item says this resource should provide a per-`item_type` schema for the `content` field. The current descriptor model has one `Sample` per family and assigns multiple heterogeneous item types to that same sample, e.g. all of `FITNESS_CHART`, `FITNESS_TABLE`, `TRACE_CHART`, `ACTIVITY_CHART`, `ACTIVITY_HISTOGRAM`, `ACTIVITY_HEATMAP`, and `ACTIVITY_MAP` share the same chart sample; `INPUT_FIELD`, `ACTIVITY_FIELD`, `INTERVAL_FIELD`, and `ACTIVITY_STREAM` share the same field/stream sample.

That means `resources/read` does not actually provide per-`item_type` schemas; it provides family-level representative examples. The tests assert that item type names appear in the Markdown, but they do not prove that each item type has its own sample/schema/path rendering.

Suggested fixes:

- Make the descriptor model keyed by item type, or add per-item-type samples under each family and render subsections for each concrete item type.
- If some item types intentionally share the exact same shape, make that explicit in the descriptor/output (for example, `shares_schema_with`) and test that every documented item type is either backed by a sample or explicitly aliases another schema.
- Add a coverage test that fails when an item type is listed without a corresponding sample/alias and inferred paths.

## Notes

- The shared validation extraction in `internal/customitemschemas/schema.go` preserves the existing create/update validation tests.
- The new resource is registered in `resources.NewRegistry()` and protocol coverage was updated.
- The Markdown correctly states that the static samples are guidance and that writes continue to validate against readable custom items.
