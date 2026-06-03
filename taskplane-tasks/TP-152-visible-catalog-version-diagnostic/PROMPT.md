# Task: TP-152 - Visible catalog/version diagnostic tool

**Created:** 2026-06-03
**Size:** S

## Review Level: 2 (Plan and Code)

**Assessment:** This adds a public MCP tool and changes the core catalog, so it needs schema/catalog review even though implementation is small. It reuses existing catalog-hash metadata patterns.
**Score:** 4/8 — Blast radius: 1, Pattern novelty: 1, Security: 1, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-152-visible-catalog-version-diagnostic/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Add a tiny MCP-visible diagnostic tool that helps assistants detect stale MCP tool catalogs even in clients that hide `_meta` from the model. icuvisor already sends `_meta.catalog_hash` and `_meta.schema_changed`, but public competitor behavior showed a practical workaround: bake the loaded version/hash into a tool description and return the live version/hash from a callable diagnostic. Implement an icuvisor-native equivalent without weakening privacy or expanding write capability.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — MCP conventions and privacy rules.
- `web/content/guides/after-upgrade.md` — current stale-catalog guidance.
- `internal/response/meta.go` — runtime catalog metadata behavior.
- `internal/mcp/catalog_hash.go` — catalog hash calculation.
- `internal/tools/list_advanced_capabilities.go` — meta-tool implementation pattern.

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `internal/tools/catalog.go`
- `internal/tools/list_advanced_capabilities.go`
- `internal/tools/check_server_version.go` (new, or equivalent name)
- `internal/tools/*check*test.go`
- `internal/toolcatalog/catalog.go`
- `internal/toolcatalog/catalog_test.go`
- `internal/mcp/catalog_hash.go`
- `internal/mcp/server.go`
- `internal/response/meta.go`
- `web/data/tools.json`
- `web/content/guides/after-upgrade.md`
- `web/content/guides/troubleshooting.md`
- `web/content/reference/tools.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm current `_meta.catalog_hash`/`schema_changed` tests and docs

### Step 1: Design diagnostic contract

**Plan-review checkpoint** — Review before adding a new public tool.

- [ ] Choose stable snake_case tool name (for example `icuvisor_check_server_version` or similar) and verify it does not conflict
- [ ] Define response shape: server version, live catalog hash, toolset, delete mode, and a short user action message when mismatch is detected
- [ ] Decide how to expose the baseline hash/version in the description without relying on per-session mutable state not available at registration
- [ ] Confirm the tool returns no secrets, athlete IDs, filesystem paths, or API-key-derived data

**Artifacts:**
- `internal/tools/check_server_version.go` (new after plan acceptance)
- `internal/toolcatalog/catalog.go` (modified after plan acceptance)
- `STATUS.md` (discoveries)

### Step 2: Implement tool and tests

- [ ] Add the diagnostic as a core, read-only/meta tool registered in every safe/core catalog
- [ ] Add unit tests for output shape and no-secret/no-athlete leakage
- [ ] Add catalog/hash tests proving description/schema changes affect the catalog hash as expected
- [ ] Ensure the tool does not depend on an intervals.icu HTTP client or network call
- [ ] Run targeted tests: `go test ./internal/tools ./internal/toolcatalog ./internal/mcp ./internal/response -run 'Check|Catalog|Schema|Advanced'`

**Artifacts:**
- `internal/tools/check_server_version.go` (new)
- `internal/tools/*check*test.go` (new/modified)
- `internal/toolcatalog/catalog.go` (modified)
- `internal/tools/catalog.go` (modified)

### Step 3: Update generated docs and stale-catalog guidance

- [ ] Regenerate tool docs/catalog data with `make docs-tools`
- [ ] Update after-upgrade/troubleshooting docs to explain when to call the diagnostic and when to reconnect vs start a new conversation
- [ ] Add CHANGELOG `[Unreleased]` entry
- [ ] Verify wording does not imply telemetry, cloud service, or credential upload

**Artifacts:**
- `web/data/tools.json` (modified)
- `web/content/guides/after-upgrade.md` (modified)
- `web/content/guides/troubleshooting.md` (modified)
- `CHANGELOG.md` (modified)

### Step 4: Testing & Verification

> ZERO test failures allowed. This step runs the FULL test suite as a quality gate.

- [ ] Run FULL test suite: `make test`
- [ ] Run lint: `make lint`
- [ ] Fix all failures
- [ ] Build passes: `make build`

### Step 5: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md
- [ ] Note any client-specific caveats discovered during implementation

## Documentation Requirements

**Must Update:**
- `web/content/guides/after-upgrade.md` — add the new diagnostic flow.
- `web/content/guides/troubleshooting.md` — add stale tool-catalog remediation.
- `CHANGELOG.md` — add user-visible tool addition.

**Check If Affected:**
- `web/content/reference/tools.md` — generated catalog should cover the new tool; hand-edit only if surrounding prose needs adjustment.
- `README.md` — update only if top-level user guidance needs to mention the diagnostic.

## Completion Criteria

- [ ] New read-only diagnostic tool is in core catalog
- [ ] Tool response and description make stale-catalog diagnosis possible without `_meta`
- [ ] Tests prove no secret/athlete/path leakage
- [ ] Docs explain when to reconnect MCP clients
- [ ] All tests passing

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-152): complete Step N — description`
- **Bug fixes:** `fix(TP-152): description`
- **Tests:** `test(TP-152): description`
- **Hydration:** `hydrate: TP-152 expand Step N checkboxes`

## Do NOT

- Add telemetry or network calls.
- Return API keys, athlete IDs, config paths, local usernames, or raw environment values.
- Remove existing `_meta.catalog_hash` / `_meta.schema_changed` behavior.
- Copy competitor implementation; use only the behavior signal summarized in this prompt.
- Skip tests.
- Modify framework/standards docs without explicit user approval.
- Load docs not listed in "Context to Read First".
- Commit without the task ID prefix in the commit message.

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if issues discovered during execution.
     Format:
     ### Amendment N — YYYY-MM-DD HH:MM
     **Issue:** [what was wrong]
     **Resolution:** [what was changed] -->
