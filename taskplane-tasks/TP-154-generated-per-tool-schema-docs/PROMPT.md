# Task: TP-154 - Generated per-tool schema docs

**Created:** 2026-06-03
**Size:** M

## Review Level: 2 (Plan and Code)

**Assessment:** This changes generated website/reference data and may affect docs build/rendering. It adapts existing catalog generation but adds a new documentation shape.
**Score:** 4/8 — Blast radius: 1, Pattern novelty: 2, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-154-generated-per-tool-schema-docs/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Improve the public tool reference so users and maintainers can see per-tool inputs and examples, not just name/group/tier/safety/summary. The current generated `web/data/tools.json` is summary-only. Generate a concise per-tool argument/reference data shape from the registered MCP schemas, with special attention to write tools and `input_examples`, so docs stay aligned with the actual registry.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — documentation and MCP schema conventions.
- `web/content/reference/tools.md` — current generated reference page.
- `web/layouts/partials/tool-catalog.html` — current catalog rendering.
- `scripts/snapshot_tool_schemas.go` — existing schema serialization patterns.
- `internal/tools/input_examples_test.go` — write-tool examples contract.

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `scripts/*tools*.go`
- `scripts/snapshot_tool_schemas.go`
- `web/data/tools.json`
- `web/data/tool_schemas.json` (new, or equivalent)
- `web/layouts/partials/tool-catalog.html`
- `web/layouts/partials/tool-toc.html`
- `web/content/reference/tools.md`
- `internal/tools/catalog.go`
- `internal/tools/catalog_test.go`
- `internal/toolcatalog/*`
- `README.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Current `make docs-tools` behavior identified

### Step 1: Design generated docs data shape

**Plan-review checkpoint** — Review the docs/data design before implementation.

- [ ] Decide whether to extend `web/data/tools.json` or add a separate generated `web/data/tool_schemas.json`
- [ ] Define a concise schema projection: arguments, required flags, type/enum/default, description, and examples without dumping noisy full JSON schemas inline
- [ ] Decide how to handle nested schemas like `workout_doc` and large examples without overwhelming the docs page
- [ ] Ensure output contains no secrets, athlete IDs, local paths, or nondeterministic values

**Artifacts:**
- `scripts/*tools*.go` (modified after plan acceptance)
- `web/layouts/partials/tool-catalog.html` (modified after plan acceptance)
- `STATUS.md` (design notes)

### Step 2: Implement generator and tests

- [ ] Update the docs-tools generator to emit the chosen per-tool schema/reference data from the live registry
- [ ] Add tests or deterministic checks that generated data includes write-tool examples and key fields like `include_full`, `date`, `category`, and `workout_doc`
- [ ] Keep generation no-network and stable across runs
- [ ] Run targeted tests: `go test ./internal/tools ./internal/toolcatalog ./internal/toolchecks -run 'Catalog|Schema|Examples'`

**Artifacts:**
- `scripts/*tools*.go` (modified)
- `web/data/tool_schemas.json` or `web/data/tools.json` (generated/modified)
- `internal/tools/catalog_test.go` or relevant generator tests (modified if applicable)

### Step 3: Render docs and refine UX

- [ ] Update Hugo partials/reference page to show per-tool arguments in a scannable way
- [ ] Ensure write tools show `input_examples` or a link/expandable section for examples
- [ ] Keep the main catalog page readable on small screens and avoid massive always-expanded payloads
- [ ] Run `make docs-tools` and any available docs build/check command; if no docs build command exists, record that in STATUS.md

**Artifacts:**
- `web/layouts/partials/tool-catalog.html` (modified)
- `web/content/reference/tools.md` (modified only if needed)
- `web/data/*.json` (generated/modified)

### Step 4: Update contributor/user guidance

- [ ] Update README or contributing docs only if the docs generation workflow changes
- [ ] Add CHANGELOG `[Unreleased]` entry for docs/reference improvement
- [ ] Verify generated docs do not expose internal-only tool names outside the registered catalog
- [ ] Review whether TP-153 snapshot policy changes would affect this docs generator; log dependency/caveat if so

**Artifacts:**
- `README.md` (modified if workflow changes)
- `CHANGELOG.md` (modified)
- `STATUS.md` (caveats)

### Step 5: Testing & Verification

> ZERO test failures allowed. This step runs the FULL test suite as a quality gate.

- [ ] Run FULL test suite: `make test`
- [ ] Run lint: `make lint`
- [ ] Fix all failures
- [ ] Build passes: `make build`

### Step 6: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md
- [ ] Include generated-file summary in delivery notes

## Documentation Requirements

**Must Update:**
- `web/content/reference/tools.md` or its rendering partials — expose per-tool schema docs.
- `CHANGELOG.md` — add docs/reference improvement note.

**Check If Affected:**
- `README.md` — update docs generation instructions if they change.
- `CONTRIBUTING.md` — update only if contributor workflow changes.
- `web/README.md` — update if website data files/build workflow changes.

## Completion Criteria

- [ ] Generated reference includes per-tool arguments and examples/projections
- [ ] Output is deterministic and no-network
- [ ] Docs remain concise/readable, especially for large nested schemas
- [ ] Existing `make docs-tools` workflow still works or is documented if changed
- [ ] All tests passing

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-154): complete Step N — description`
- **Bug fixes:** `fix(TP-154): description`
- **Tests:** `test(TP-154): description`
- **Hydration:** `hydrate: TP-154 expand Step N checkboxes`

## Do NOT

- Dump entire raw schemas into an always-expanded docs page if that makes reference unusable.
- Include secrets, local paths, athlete IDs, or nondeterministic values in generated data.
- Change public tool schemas as part of docs generation unless strictly necessary and reviewed.
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
