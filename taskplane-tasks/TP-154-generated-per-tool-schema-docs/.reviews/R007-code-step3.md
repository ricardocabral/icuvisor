# Code Review: Step 3 — Render docs and refine UX

**Verdict:** REVISE

## Findings

### 1. Missing per-tool schema entries are silently omitted

- **Location:** `web/layouts/partials/tool-catalog.html:36-85`
- **Severity:** Medium

The partial correctly errors when `web/data/tool_schemas.json` is missing or empty, but each row then does:

```go-html-template
{{ $schema := index $schemas .name }}
...
{{ with $schema }} ... {{ end }}
```

If a future `make docs-tools`/data edit drops one tool from `tool_schemas.json`, Hugo will still build and that tool will render with no arguments/examples. That undermines the task goal of keeping the public reference aligned with the live registry, and it misses the Step 3 plan note to fail loudly for rows without a matching schema.

Please add a per-row guard, for example immediately after `$schema := index $schemas .name`:

```go-html-template
{{ if not $schema }}{{ errorf "web/data/tool_schemas.json is missing schema for tool %q; run make docs-tools" .name }}{{ end }}
```

### 2. Tools with zero arguments render no input-state fallback

- **Location:** `web/layouts/partials/tool-catalog.html:43-76`
- **Severity:** Low

For schema entries with an empty `arguments` array, the `{{ with .arguments }}` block is skipped entirely. The current generated data has zero-argument tools such as `icuvisor_list_advanced_capabilities` and `list_athletes`, so the “Summary and inputs” column gives no explicit indication that these tools take no arguments. The approved Step 3 plan asked for tidy fallback text for no-argument/no-example cases.

Consider rendering a collapsed details row or short inline text like `Arguments (0): none` when a schema exists but has no arguments. For examples, it may be enough to show a fallback only for write tools without examples to avoid cluttering every read tool.

## Verification run

- `make docs-tools` — passed
- `make web-build` — passed, with existing Hugo deprecation warnings for `.Site.Data` and `.Language.LanguageDirection`
