# Task: TP-017 - Raise Go test coverage to 80%

**Created:** 2026-05-13
**Size:** M

## Review Level: 1 (Plan Only)

**Assessment:** This is a test-focused task, but it may touch multiple internal packages to find high-value coverage gaps. It does not change product behavior, auth, persistence, or public MCP schemas.
**Score:** 2/8 — Blast radius: 2, Pattern novelty: 0, Security: 0, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-017-raise-test-coverage-to-80/
├── PROMPT.md   ← This file (immutable above --- divider)
├── STATUS.md   ← Execution state (worker updates this)
├── .reviews/   ← Reviewer output (created by the orchestrator runtime)
└── .DONE       ← Created when complete
```

## Mission

Increase the repository's Go statement coverage from the current baseline of 76.9% to at least 80.0% while preserving behavior. Add focused, table-driven tests around currently under-covered branches in the intervals client, wellness parsing, activity streams, extended metrics shaping, and/or tool catalog guards. The goal is stronger regression protection for existing code, not broader product scope or production rewrites.

## Dependencies

- **None**

## Context to Read First

> Only list docs the worker actually needs. Less is better.

**Tier 2 (area context):**

- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**

- `CLAUDE.md` — project Go/testing rules and clean-room constraints
- `CONTRIBUTING.md` — contributor test/lint expectations

## Environment

- **Workspace:** `/Users/jusbrasil/prj/icuvisor`
- **Services required:** None

## File Scope

> The orchestrator uses this to avoid merge conflicts: tasks with overlapping
> file scope run on the same lane (serial), not in parallel.

Expected files and directories:

- `internal/intervals/*_test.go`
- `internal/tools/*_test.go`
- `internal/toolchecks/*_test.go`
- `internal/app/*_test.go` (only if a small, behavior-preserving test seam is needed)
- `taskplane-tasks/TP-017-raise-test-coverage-to-80/STATUS.md`

Avoid production-code changes unless they are tiny, behavior-preserving testability seams.

## Steps

> **Hydration:** STATUS.md tracks outcomes, not individual code changes. Workers
> expand steps when runtime discoveries warrant it. See task-worker agent for rules.

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm baseline with `go test ./... -coverprofile=coverage.out -covermode=atomic` and `go tool cover -func=coverage.out | tail -1`

### Step 1: Pick high-value coverage targets

> **Plan-review checkpoint** — Review the selected target list before implementation; this task should use one consolidated plan review rather than per-step plan reviews.

- [ ] Generate a per-function coverage report and identify the smallest set of tests likely to lift total coverage to at least 80.0%
- [ ] Prioritize behavior-rich low-coverage functions before testing trivial `main` or hard-to-observe branches
- [ ] Record selected target files/functions and rationale in `STATUS.md`

**Suggested target areas:**

- `internal/intervals/activity_streams.go` — `ActivityStream.UnmarshalJSON`, `Client.GetActivityStreams`
- `internal/intervals/wellness.go` — `Wellness.UnmarshalJSON`, `ListWellness`, native-provider extraction helpers
- `internal/tools/get_extended_metrics.go` — Strava-unavailable response, optional-source branches, extended metric shaping helpers
- `internal/toolchecks/confusable_names.go` and `internal/toolchecks/schema_stability.go` — catalog/snapshot guard branches where tests are cheap and deterministic

**Artifacts:**

- `taskplane-tasks/TP-017-raise-test-coverage-to-80/STATUS.md` (modified)

### Step 2: Add focused intervals-package tests

- [ ] Create or extend intervals tests for activity stream JSON decoding, raw-field preservation, required IDs, query parameters, and error wrapping
- [ ] Create `internal/intervals/wellness_test.go` with table-driven tests for native wellness provider extraction, duplicate claimed-key handling, required `oldest`, fields query compaction, and HTTP error wrapping
- [ ] Use `httptest` or existing client test helpers; do not hit the network
- [ ] Run targeted tests: `go test ./internal/intervals -cover`

**Artifacts:**

- `internal/intervals/activity_streams_test.go` (new or modified)
- `internal/intervals/wellness_test.go` (new)

### Step 3: Add focused tool/toolcheck tests until the target is met

- [ ] Add tests for uncovered `internal/tools/get_extended_metrics.go` branches, especially Strava-unavailable payloads, include_full behavior, optional `ErrNotFound`/`ErrUnauthorized` sources, and non-optional error propagation
- [ ] Add or extend `internal/toolchecks` tests for deterministic catalog generation and schema-stability branches only if additional coverage is still needed after Step 2
- [ ] Keep tests behavior-based; avoid asserting implementation details that would make harmless refactors brittle
- [ ] Run targeted tests for touched packages, for example `go test ./internal/tools ./internal/toolchecks -cover`

**Artifacts:**

- `internal/tools/get_extended_metrics_test.go` (new or modified)
- `internal/toolchecks/*_test.go` (modified only if needed)

### Step 4: Testing & Verification

> ZERO test failures allowed. This step runs the FULL test suite as a quality gate.

- [ ] Run coverage gate: `go test -race -count=1 -coverprofile=coverage.txt -covermode=atomic ./...`
- [ ] Verify `go tool cover -func=coverage.txt | tail -1` reports total statement coverage >= 80.0%
- [ ] Run lint: `make lint`
- [ ] Run build: `make build`
- [ ] Remove local coverage artifacts unless they are already ignored and intentionally kept out of git

### Step 5: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md
- [ ] Final coverage percentage recorded in STATUS.md

## Documentation Requirements

**Must Update:**

- `taskplane-tasks/TP-017-raise-test-coverage-to-80/STATUS.md` — selected targets, final coverage percentage, discoveries

**Check If Affected:**

- `CHANGELOG.md` — update only if production behavior changes, which this task should avoid
- `CONTRIBUTING.md` — update only if test commands or contributor expectations change, which this task should avoid

## Completion Criteria

- [ ] Overall Go statement coverage is at least 80.0% using `go test -race -count=1 -coverprofile=coverage.txt -covermode=atomic ./...`
- [ ] Added tests are deterministic, table-driven where appropriate, and do not hit the network
- [ ] All tests, lint, and build pass
- [ ] Any production-code edits are minimal, behavior-preserving, and justified in STATUS.md
- [ ] Documentation requirements are satisfied

## Git Commit Convention

Commits happen at **step boundaries** (not after every checkbox). All commits
for this task MUST include the task ID for traceability:

- **Step completion:** `test(TP-017): complete Step N — description`
- **Bug fixes:** `fix(TP-017): description`
- **Hydration:** `hydrate: TP-017 expand Step N checkboxes`

## Do NOT

- Do not lower the coverage target below 80.0%
- Do not delete or weaken existing tests to make coverage easier
- Do not hit intervals.icu or any external network service from tests
- Do not add GPL/copyleft dependencies or copy code from non-permissive sources
- Do not expand into feature work, MCP schema changes, or broad refactors
- Do not commit generated coverage reports unless the repository already tracks them intentionally
- Do not load docs not listed in "Context to Read First"
- Do not commit without the task ID prefix in the commit message

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if issues discovered during execution.
     Format:
     ### Amendment N — YYYY-MM-DD HH:MM
     **Issue:** [what was wrong]
     **Resolution:** [what was changed] -->
