# Task: TP-105 - Tool routing smoke eval

**Created:** 2026-05-26
**Size:** M

## Review Level: 2

**Assessment:** New opt-in evaluation harness and fixtures; it should not affect runtime behavior, but provider/API-key handling and catalog loading need review.
**Score:** 4/8 — Blast radius: 1, Pattern novelty: 2, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-105-tool-routing-smoke-eval/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Add a first-tool-call smoke evaluation suite that detects MCP tool-routing regressions before they reach users. The eval should load icuvisor's registered tool definitions, ask a configurable model which tool it would call first for fixed prompts, and report expected versus actual tool choices without executing any icuvisor tool handlers.

Tracking issue: https://github.com/ricardocabral/icuvisor/issues/32

## Dependencies

- **Task:** TP-015 (catalog disambiguation/schema stability exists)
- **Task:** TP-030 (toolset tiers exist)
- **Task:** TP-034 (benchmark harness patterns may be reusable)

## Context to Read First

**Tier 2 (area context):**
- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**
- `CLAUDE.md` — repository rules and test/network constraints.
- `docs/prd/PRD-icuvisor.md` — tool catalog and token-efficiency constraints.
- `internal/tools/catalog.go` — registered catalog metadata.
- `internal/mcp/catalog_hash.go` — catalog construction without starting transports.
- `docs/kr5-benchmark.md` and benchmark scripts, if present — reuse patterns where appropriate.

## Environment

- **Workspace:** repository root (`/Users/jusbrasil/prj/icuvisor`)
- **Services required:** None for unit tests. The eval command may call a model provider only when an explicit provider API key is configured by the operator.

## File Scope

- `scripts/**/*`
- `tools/**/*`
- `internal/toolcatalog/**/*`
- `internal/mcp/catalog_hash.go`
- `internal/tools/catalog.go`
- `testdata/**/*`
- `docs/**/*eval*`
- `CONTRIBUTING.md`
- `Makefile`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Identify any existing benchmark/eval harness patterns to reuse

### Step 1: Design eval fixture and expected-result format

- [ ] Define prompt-case fixture format with ID, prompt, expected first tool, allowed catalog mode/toolset, and notes.
- [ ] Add initial cases for activity reads, event writes, workout-library operations, and analyzer helpers.
- [ ] Include safe/full catalog expectations for destructive-tool cases where practical.
- [ ] Add unit tests for fixture loading and expected-result comparison.

**Artifacts:**
- Eval fixture files under `testdata/`, `scripts/`, or `tools/` (new)
- Fixture-loading tests (new)

### Step 2: Implement opt-in first-tool-call runner

- [ ] Load icuvisor tool definitions for core/full catalogs without executing handlers.
- [ ] Call a configurable model provider only when required environment variables are present.
- [ ] Capture first selected tool name or explicit no-tool result.
- [ ] Produce clear pass/fail output with per-case detail.
- [ ] Ensure normal tests remain network-free.

**Artifacts:**
- Eval runner under `scripts/` or `tools/` (new)
- Supporting Go/Python code/tests as appropriate (new/modified)

### Step 3: Wire command and documentation

- [ ] Add a Make target or documented command, for example `make eval-tools`.
- [ ] Document required environment variables and the no-tool-execution guarantee.
- [ ] Document that the eval is opt-in and not a default CI gate unless configured.
- [ ] Update `CHANGELOG.md` if adding user/developer-visible tooling.

**Artifacts:**
- `Makefile` (modified if command added)
- `CONTRIBUTING.md` or `docs/*` (modified/new)
- `CHANGELOG.md` (modified if needed)

### Step 4: Testing & Verification

- [ ] Run targeted tests added/affected by this task
- [ ] Run FULL test suite: `make test`
- [ ] Build passes: `make build`
- [ ] Lint passes: `make lint`
- [ ] If provider credentials are available, run the new eval command and record summary output in STATUS.md
- [ ] Fix all failures or document pre-existing unrelated failures in STATUS.md

### Step 5: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] `STATUS.md` discoveries and final notes updated
- [ ] Commit at step boundary with the task ID in the message

## Documentation Requirements

**Must Update:**
- `CONTRIBUTING.md` or a dedicated `docs/` page — explain how to run the opt-in eval.
- `STATUS.md` — keep execution state current.

**Check If Affected:**
- `Makefile` help output — include the eval target if one is added.
- `CHANGELOG.md` — update if developer tooling changes are user-visible.
- `docs/kr5-benchmark.md` — reference only if this reuses or extends benchmark tooling.

## Completion Criteria

- Eval harness loads tool definitions and expected cases without executing handlers.
- Fixture suite covers confusable read tools, write/update routing, workout-library tools, and analyzer helpers.
- Results show pass/fail counts plus per-case details.
- Unit tests for fixture/result logic pass without network access.
- Full quality gate passes or pre-existing failures are documented.

## Git Commit Convention

Commits happen at **step boundaries**. All commits for this task MUST include `TP-105` for traceability. Examples:

- `feat(TP-105): complete step 1 — add routing eval fixtures`
- `test(TP-105): cover eval result comparison`
- `hydrate: TP-105 expand step checkboxes`

## Do NOT

- Do not execute MCP tool handlers from the eval.
- Do not hit intervals.icu from tests or eval fixtures.
- Do not require provider credentials for `make test`.
- Do not log provider API keys or model credentials.
- Do not make the eval a required CI gate unless explicitly configured and documented.

---

## Amendments

_Add amendments below this line only._
