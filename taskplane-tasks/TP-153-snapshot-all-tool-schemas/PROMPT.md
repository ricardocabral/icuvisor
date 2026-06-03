# Task: TP-153 - Snapshot every registered MCP tool schema

**Created:** 2026-06-03
**Size:** M

## Review Level: 1 (Plan Only)

**Assessment:** This is primarily test infrastructure and committed golden data, with low runtime risk. It still needs plan review because broad snapshot expansion can create noisy maintenance costs.
**Score:** 3/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-153-snapshot-all-tool-schemas/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Expand icuvisor's schema-stability guard so every registered MCP tool has input-schema snapshot coverage, or an explicit documented reason for exclusion. The current guard snapshots only a curated subset while the generated public catalog exposes 60 tools. Full coverage makes accidental MCP API drift visible in PRs and protects users from stale-client surprises.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — MCP schema and testing conventions.
- `internal/toolchecks/schema_stability.go` — current schema snapshot guard.
- `internal/tools/schema_snapshot/` — committed snapshot format.
- `scripts/snapshot_tool_schemas.go` — snapshot generation command.

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `internal/toolchecks/schema_stability.go`
- `internal/toolchecks/schema_stability_test.go`
- `internal/tools/schema_snapshot/*.json`
- `scripts/snapshot_tool_schemas.go`
- `internal/tools/catalog_test.go`
- `internal/tools/registry_test.go`
- `web/content/reference/tools.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Count current registered tools and current schema snapshots

### Step 1: Decide snapshot coverage policy

**Plan-review checkpoint** — Review the maintenance policy before generating broad snapshot churn.

- [ ] Compare live registered catalog against `schemaCatalogToolNames`
- [ ] Decide whether to snapshot all tool input schemas in all/full mode, core/safe mode, or both
- [ ] Decide how to handle coach-mode injected `athlete_id` schemas
- [ ] Document any intentionally excluded tools and why in code comments/tests

**Artifacts:**
- `internal/toolchecks/schema_stability.go` (modified after plan acceptance)
- `STATUS.md` (coverage counts/discoveries)

### Step 2: Implement full coverage guard

- [ ] Replace or extend the curated whitelist so new registered tools cannot silently avoid snapshot coverage
- [ ] Add tests that fail when a registered public tool has no snapshot and no explicit exclusion
- [ ] Keep generation no-network and deterministic
- [ ] Run targeted tests: `go test ./internal/toolchecks ./internal/tools ./internal/mcp -run 'Schema|Catalog|Registry'`

**Artifacts:**
- `internal/toolchecks/schema_stability.go` (modified)
- `internal/toolchecks/schema_stability_test.go` (modified)
- `scripts/snapshot_tool_schemas.go` (modified if needed)

### Step 3: Regenerate snapshots and review churn

- [ ] Run `go run ./scripts/snapshot_tool_schemas.go`
- [ ] Review added/changed JSON snapshots for accidental secrets, local paths, or unstable ordering
- [ ] If descriptions/examples produce excessive noise, document the chosen policy rather than weakening structural coverage silently
- [ ] Run targeted tests again after snapshot refresh

**Artifacts:**
- `internal/tools/schema_snapshot/*.json` (added/modified)

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
- [ ] Summarize final snapshot policy and maintenance tradeoffs

## Documentation Requirements

**Must Update:**
- `CHANGELOG.md` — add a developer-facing test/tooling note if snapshot policy changes.

**Check If Affected:**
- `web/content/reference/tools.md` — update surrounding prose only if docs generation assumptions change.
- `CONTRIBUTING.md` — update only if contributors need a new snapshot workflow command beyond existing `go run ./scripts/snapshot_tool_schemas.go` / `make docs-tools`.

## Completion Criteria

- [ ] Every registered public MCP tool input schema is snapshotted or explicitly excluded with a test-enforced reason
- [ ] Snapshot generation remains deterministic and no-network
- [ ] Added snapshots contain no secrets/local paths
- [ ] All tests passing

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits for this task MUST include the task ID for traceability:

- **Step completion:** `feat(TP-153): complete Step N — description`
- **Bug fixes:** `fix(TP-153): description`
- **Tests:** `test(TP-153): description`
- **Hydration:** `hydrate: TP-153 expand Step N checkboxes`

## Do NOT

- Weaken schema stability checks just to reduce snapshot churn.
- Include live athlete data, API keys, local filesystem paths, or nondeterministic values in snapshots.
- Add network calls to snapshot generation.
- Change public schemas unless needed for the guard; this is not a feature task.
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
