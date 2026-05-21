# Code Review — Step 2: Reference section first

Verdict: **request changes**

## Finding

### 1. Coach config example includes an invalid ACL tool name

- **Severity:** Blocking
- **File:** `web/content/reference/config-file.md:56`

The coach config example documents this roster entry:

```json
"allowed_tools": ["get_*", "icuvisor_list_advanced_capabilities"]
```

That config will fail to load. Coach ACL patterns are validated by `toolcatalog.ValidateACLPattern`, which accepts only `*`, a prefix wildcard matching athlete-scoped tools, or an exact athlete-scoped tool name. `icuvisor_list_advanced_capabilities` is deliberately outside the athlete-scoped ACL set (`internal/toolcatalog/catalog.go:92` appends it only to `allToolNames`, while `ValidateACLPattern` rejects anything for which `IsAthleteScopedTool` is false at lines 133-134). In coach mode, that tool is handled as an always-visible control/meta tool, not something that belongs in per-athlete `allowed_tools`.

This is especially confusing because the table immediately below the example correctly says `allowed_tools` accepts exact athlete-scoped tool names, `*`, or prefix wildcards.

**Suggested fix:** remove `icuvisor_list_advanced_capabilities` from the example, or replace it with a valid athlete-scoped exact tool such as `get_fitness` if the example needs to demonstrate mixing a wildcard and an exact allow. If you want to mention `icuvisor_list_advanced_capabilities`, do it in prose as a meta/control tool that is not configured via per-athlete ACLs.

## Other checks

- Ran `cd web && hugo --minify --gc`; the site builds successfully.
- Verified the `reference/cli.md` full help block exactly matches `internal/app/testdata/help.golden`.
- Verified R007's broken generated-tool-catalog link is fixed: rendered full-tier badges now point to `/reference/safety-modes/#toolset-tier`, the `#toolset-tier` target exists, and no rendered `web/public` page links to `/reference/toolset-tiers/`.
