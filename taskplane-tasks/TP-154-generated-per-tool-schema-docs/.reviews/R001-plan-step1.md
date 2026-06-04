# Plan Review: Step 1 — Design generated docs data shape

**Verdict:** Changes requested / not reviewable yet

The Step 1 checkpoint has not produced the design artifact required by the prompt. `STATUS.md` only shows unchecked outcome bullets and contains no actual decisions for:

- whether to extend `web/data/tools.json` or generate a separate `web/data/tool_schemas.json`;
- the exact concise projection shape for arguments/examples;
- how nested schemas such as `workout_doc` and large examples will be summarized or collapsed;
- the deterministic/no-secret boundary.

Before implementation starts, please update `STATUS.md` with concrete design notes that answer those points.

## Specific gaps to address

1. **Actual generator path:** preflight found `make docs-tools` uses `go run ./cmd/gendocs --out web/data/tools.json`, not a `scripts/*tools*.go` generator. The plan should name `cmd/gendocs`/its golden test as the implementation target, or explicitly explain any new script/generator split.
2. **Schema projection contract:** define the JSON fields precisely, e.g. per tool `arguments[]` with `name`, `required`, `type`, `enum`, `default`, `description`, and optional `children`/`summary` for objects. Include how arrays, `anyOf`/`oneOf`, nullable fields, and `additionalProperties` will be represented.
3. **Examples policy:** specify whether examples come from `input_examples` only, whether they are capped/truncated, and how write-tool examples remain visible without making the page huge.
4. **Nested/large fields:** document the planned treatment for `workout_doc`, custom item `content`, and other object-heavy inputs so the docs are useful but not raw schema dumps.
5. **Safety/determinism:** state that data is generated from the full registered catalog with deterministic ordering, no network calls, and no secrets/athlete IDs/local paths. If sample IDs in examples are allowed, define what counts as safe placeholder data.

Once those design decisions are recorded, this checkpoint can be reviewed meaningfully.
